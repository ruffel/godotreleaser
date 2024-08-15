package dependencies

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pterm/pterm"
	"github.com/ruffel/godotreleaser/internal/godot/url"
	"github.com/ruffel/godotreleaser/internal/paths"
	"github.com/ruffel/godotreleaser/internal/terminal"
	"github.com/ruffel/godotreleaser/internal/terminal/messages"
	"github.com/ruffel/godotreleaser/internal/utils/downloader"
	"github.com/ruffel/godotreleaser/internal/utils/unzip"
	"github.com/samber/lo"
	"github.com/spf13/afero"
)

func Run(_ context.Context, fs afero.Fs, version string, mono bool) error {
	terminal.Send(messages.NewStage("Configuring Godot " + version))

	if err := downloadGodot(fs, version, mono); err != nil {
		return err // nolint:wrapcheck
	}

	return nil
}

//nolint:cyclop,funlen
func downloadGodot(fs afero.Fs, version string, mono bool) error {
	slog.Info("Fetching Godot binaries and export templates", "version", version, "mono", mono)

	//--------------------------------------------------------------------------
	// Check if this configuration already exists...
	//--------------------------------------------------------------------------
	exportPath := paths.TemplatePath(version, mono)

	binaryExists, err := paths.CheckBinaryExists(version, mono)
	if err != nil {
		return err //nolint:wrapcheck
	}

	exportExists, err := afero.Exists(fs, exportPath)
	if err != nil {
		return err //nolint:wrapcheck
	}

	if binaryExists && exportExists {
		return nil
	}

	multi := pterm.DefaultMultiPrinter

	binaryTracker := NewDownloadTracker(lo.Must(pterm.DefaultProgressbar.WithWriter(multi.NewWriter()).
		WithTitle("Downloading Godot Binaries").
		WithShowCount(false).
		WithShowPercentage(true).
		WithShowElapsedTime(true).
		Start()))

	_ = binaryTracker

	templateTracker := NewDownloadTracker(lo.Must(pterm.DefaultProgressbar.WithWriter(multi.NewWriter()).
		WithTitle("Downloading Godot Templates").
		WithShowCount(false).
		WithShowPercentage(true).
		WithShowElapsedTime(true).
		Start()))

	_ = templateTracker

	if _, err := multi.Start(); err != nil {
		pterm.Error.Println("Failed to start multi printer:", err)

		return err //nolint:wrapcheck
	}

	var wg sync.WaitGroup

	wg.Add(2) //nolint:mnd

	binaryZipPath := filepath.Join(paths.Version(version, mono), "godot.zip")

	// Download the files concurrently
	go func() {
		defer wg.Done()

		if err := fs.MkdirAll(filepath.Dir(binaryZipPath), 0o0755); err != nil {
			return
		}

		if binaryExists {
			_, _ = binaryTracker.Tracker.Stop()

			return
		}

		binaryAddress, err := url.BuildBinaryURL(version, mono)
		if err != nil {
			pterm.Error.Println("Failed to build Godot binary URL:", err)

			return
		}

		err = downloader.DownloadFile(context.Background(), binaryAddress, binaryZipPath, downloader.WithProgress(binaryTracker))
		if err != nil {
			pterm.Error.Println("Failed to download Godot binary:", err)
		}

		slog.Debug("Downloaded godot binaries", "version", version, "mono", mono, "url", binaryAddress, "dst", binaryZipPath)
	}()

	templatePath := filepath.Join(paths.Version(version, mono), "templates.tpz")

	go func() {
		defer wg.Done()

		if exportExists {
			_, _ = templateTracker.Tracker.Stop()

			return
		}

		if err := fs.MkdirAll(filepath.Dir(templatePath), 0o0755); err != nil {
			return
		}

		templateAddress, _ := url.BuildTemplateURL(version, mono)

		err := downloader.DownloadFile(context.Background(), templateAddress, templatePath, downloader.WithProgress(templateTracker))
		if err != nil {
			pterm.Error.Println("Failed to download Godot templates:", err)
		}

		slog.Debug("Downloaded godot export templates", "version", version, "mono", mono, "url", templateAddress, "dst", templatePath)
	}()

	wg.Wait()

	_, _ = multi.Stop()

	versionDir := paths.Version(version, mono)

	if !binaryExists {
		// Now that we have the files, we can extract them.
		pterm.Info.Println("Extracting Godot binary...")

		src := binaryZipPath
		dst := filepath.Join(versionDir, "editor")

		if err := unzip.Extract(src, dst); err != nil {
			pterm.Error.Println("Failed to extract Godot binary:", err)

			return err //nolint:wrapcheck
		}

		slog.Debug("Extracted godot binaries", "src", src, "dst", dst)
	}

	if !exportExists {
		pterm.Info.Println("Extracting Godot templates...")

		src := templatePath
		dst := paths.TemplatePath(version, mono)

		if err := unzip.Extract(src, dst); err != nil {
			pterm.Error.Println("Failed to extract Godot templates:", err)

			return err //nolint:wrapcheck
		}

		slog.Debug("Extracted Godot export templates", "src", src, "dst", dst)
	}

	// Clean up zip files
	if err := os.Remove(binaryZipPath); err != nil {
		pterm.Warning.Printf("Failed to remove binary zip file: %v\n", err)
	}

	if err := os.Remove(templatePath); err != nil {
		pterm.Warning.Printf("Failed to remove template zip file: %v\n", err)
	}

	pterm.Success.Println("Godot and templates extracted successfully")

	return nil
}

type DownloadTracker struct {
	Downloaded    int64
	Total         int64
	Tracker       *pterm.ProgressbarPrinter
	reader        io.Reader
	startTime     time.Time
	originalTitle string
}

func NewDownloadTracker(progress *pterm.ProgressbarPrinter) *DownloadTracker {
	return &DownloadTracker{
		Tracker:       progress,
		startTime:     time.Now(),
		originalTitle: progress.Title,
	}
}

func (d *DownloadTracker) SetTotal(total int64) {
	d.Total = total
	d.Tracker.Total = int(total)
	d.updateTitle()
}

func (d *DownloadTracker) SetReader(reader io.Reader) {
	d.reader = reader
}

func (d *DownloadTracker) Read(p []byte) (int, error) {
	n, err := d.reader.Read(p)

	d.Downloaded += int64(n)
	d.Tracker.Add(n)
	d.updateTitle()

	return n, err //nolint:wrapcheck
}

func (d *DownloadTracker) updateTitle() {
	// elapsed := time.Since(d.startTime).Seconds()

	// speed := float64(d.Downloaded) / elapsed / 1024 / 1024 // nolint:mnd

	// progress := fmt.Sprintf("%s / %s (%.2f MB/s)",
	// 	formatSize(d.Downloaded),
	// 	formatSize(d.Total),
	// 	speed)

	// d.Tracker.Title = fmt.Sprintf("%s - %s", d.originalTitle, progress)
}

func formatSize(bytes int64) string { // nolint:unused
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0

	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
