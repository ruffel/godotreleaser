package build

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pterm/pterm"
	"github.com/ruffel/godotreleaser/internal/godot/url"
	"github.com/ruffel/godotreleaser/internal/paths"
	"github.com/ruffel/godotreleaser/internal/utils/downloader"
	"github.com/ruffel/godotreleaser/internal/utils/unzip"
	"github.com/samber/lo"
	"github.com/spf13/afero"
)

//nolint:cyclop,funlen
func downloadGodot(fs afero.Fs, version string, mono bool) error {
	_ = fs
	//--------------------------------------------------------------------------
	// Check if this configuration already exists...
	//--------------------------------------------------------------------------
	// binaryPath :=

	// found, err := afero.Exists(fs, )

	//--------------------------------------------------------------------------
	// Check if this configuration already exists...
	//--------------------------------------------------------------------------

	multi := pterm.DefaultMultiPrinter

	binaryTracker := NewDownloadTracker(lo.Must(pterm.DefaultProgressbar.WithWriter(multi.NewWriter()).
		WithTitle("Downloading Godot Binaries").
		WithShowCount(false).
		WithShowPercentage(true).
		WithShowElapsedTime(false).
		Start()))

	_ = binaryTracker

	templateTracker := NewDownloadTracker(lo.Must(pterm.DefaultProgressbar.WithWriter(multi.NewWriter()).
		WithTitle("Downloading Godot Templates").
		WithShowCount(false).
		WithShowPercentage(true).
		WithShowElapsedTime(false).
		Start()))

	_ = templateTracker

	if _, err := multi.Start(); err != nil {
		pterm.Error.Println("Failed to start multi printer:", err)

		return err //nolint:wrapcheck
	}

	var wg sync.WaitGroup

	wg.Add(2) //nolint:mnd

	binaryPath := filepath.Join(paths.Version(version, mono), "godot.zip")

	// Download the files concurrently
	go func() {
		defer wg.Done()

		binaryAddress, err := url.BuildBinaryURL(version, mono)
		if err != nil {
			pterm.Error.Println("Failed to build Godot binary URL:", err)

			return
		}

		err = downloader.DownloadFile(context.Background(), binaryAddress, binaryPath, downloader.WithProgress(binaryTracker)) //nolint:lll
		if err != nil {
			pterm.Error.Println("Failed to download Godot binary:", err)
		}
	}()

	templatePath := filepath.Join(paths.Version(version, mono), "templates.tpz")

	go func() {
		defer wg.Done()

		templateAddress, _ := url.BuildTemplateURL(version, mono)

		err := downloader.DownloadFile(context.Background(), templateAddress, templatePath, downloader.WithProgress(templateTracker)) //nolint:lll
		if err != nil {
			pterm.Error.Println("Failed to download Godot templates:", err)
		}
	}()

	wg.Wait()

	versionDir := paths.Version(version, mono)

	// Now that we have the files, we can extract them.
	pterm.Info.Println("Extracting Godot binary...")

	if err := unzip.Extract(binaryPath, filepath.Join(versionDir, "editor")); err != nil {
		pterm.Error.Println("Failed to extract Godot binary:", err)

		return err //nolint:wrapcheck
	}

	pterm.Info.Println("Extracting Godot templates...")

	if err := unzip.Extract(templatePath, paths.TemplatePath(version, mono)); err != nil {
		pterm.Error.Println("Failed to extract Godot templates:", err)

		return err //nolint:wrapcheck
	}

	// // Clean up zip files
	// if err := os.Remove(binaryPath); err != nil {
	// 	pterm.Warning.Printf("Failed to remove binary zip file: %v\n", err)
	// }

	// if err := os.Remove(templatePath); err != nil {
	// 	pterm.Warning.Printf("Failed to remove template zip file: %v\n", err)
	// }

	pterm.Success.Println("Godot and templates extracted successfully")

	var godotBinary string

	err := filepath.Walk(versionDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasPrefix(info.Name(), "Godot") {
			godotBinary = path

			return filepath.SkipDir
		}

		return nil
	})
	if err != nil {
		pterm.Error.Printf("Failed to find Godot binary: %v\n", err)

		return err //nolint:wrapcheck
	}

	if godotBinary == "" {
		pterm.Error.Println("Godot binary not found")

		return errors.New("godot binary not found")
	}

	// Create a symlink to the Godot binary
	symlinkPath := filepath.Join(versionDir, "godot")

	os.Remove(symlinkPath)

	err = os.Symlink(godotBinary, symlinkPath)
	if err != nil {
		pterm.Error.Printf("Failed to create symlink: %v\n", err)

		return err //nolint:wrapcheck
	}

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
