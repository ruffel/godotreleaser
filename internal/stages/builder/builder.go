package builder

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ruffel/godotreleaser/internal/paths"
	"github.com/ruffel/godotreleaser/internal/terminal/messages"
	"github.com/ruffel/godotreleaser/pkg/godot/config/exports"
	"github.com/spf13/afero"
)

func Run(_ context.Context, _ afero.Fs, version string, mono bool, path string) error {
	if err := os.MkdirAll(paths.Version(version, mono), 0o0755); err != nil {
		return err //nolint:wrapcheck
	}

	binary, err := paths.GetBinary(version, mono)
	if err != nil {
		return err //nolint:wrapcheck
	}

	e, err := exports.New(filepath.Join(filepath.Dir(path), "export_presets.cfg"))
	if err != nil {
		return err //nolint:wrapcheck
	}

	for _, name := range e.PresetNames() {
		fmt.Fprintln(os.Stdout, messages.NewStage(fmt.Sprintf("Building Project (%s)", name)))

		cmd := exec.Command(binary, "--verbose", "--headless", "--quit", "--export-release", name, path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to build project: %w", err)
		}
	}

	return nil
}
