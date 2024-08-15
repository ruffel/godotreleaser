package builder

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ruffel/godotreleaser/internal/paths"
	"github.com/ruffel/godotreleaser/internal/terminal"
	"github.com/ruffel/godotreleaser/internal/terminal/messages"
	"github.com/ruffel/godotreleaser/pkg/godot/config/exports"
	"github.com/spf13/afero"
)

func Run(_ context.Context, fs afero.Fs, version string, mono bool, path string) error {
	if err := os.MkdirAll(paths.Version(version, mono), 0o0755); err != nil {
		return err //nolint:wrapcheck
	}

	e, err := exports.New(filepath.Join(filepath.Dir(path), "export_presets.cfg"))
	if err != nil {
		return err //nolint:wrapcheck
	}

	binary, err := paths.GetBinary(version, mono)
	if err != nil {
		return err //nolint:wrapcheck
	}

	for _, preset := range e.Presets() {
		name := preset.Name
		dst := filepath.Join(filepath.Dir(path), filepath.Dir(preset.ExportPath))

		terminal.Send(messages.NewStage(fmt.Sprintf("Building Project (%s)", name)))

		found, err := afero.DirExists(fs, dst)
		if err != nil {
			return err //nolint:wrapcheck
		}

		if !found {
			if err := fs.MkdirAll(dst, 0o0755); err != nil {
				return fmt.Errorf("failed to create export directory: %w", err)
			}

			slog.Debug("Created preset output directory", "preset", name, "dst", dst)
		}

		cmd := exec.Command(binary, "--verbose", "--headless", "--quit", "--export-release", name, path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to build project: %w", err)
		}

		slog.Info("Successfully built target preset", "preset", name, "dst", dst)
	}

	return nil
}
