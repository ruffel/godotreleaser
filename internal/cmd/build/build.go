package build

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	project "github.com/ruffel/godotreleaser/internal/godot/project"
	"github.com/ruffel/godotreleaser/internal/paths"
	"github.com/samber/lo"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type buildOpts struct {
	ProjectDir string
	Version    string
	// Dependencies
	fs afero.Fs
}

func NewBuildCmd() *cobra.Command {
	opts := &buildOpts{
		fs: afero.NewOsFs(),
	}

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build the Godot project",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runBuild(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.ProjectDir, "project", "p", "", "Path to the Godot project directory (defaults to the current directory)")
	cmd.Flags().StringVarP(&opts.Version, "version", "v", "", "Godot version to use")

	return cmd
}

func runBuild(opts *buildOpts) error {
	path, err := findProjectFile(opts.fs, opts.ProjectDir)
	if err != nil {
		return fmt.Errorf("failed to find project file: %w", err)
	}

	slog.Debug("Found Godot project file", "path", path)

	if _, err := project.New(path); err != nil {
		return fmt.Errorf("project file is not valid: %w", err)
	}

	//--------------------------------------------------------------------------
	// We need a Godot binary and export templates to build the project.
	//
	// Download the Godot binary and export templates if they don't exist.
	//--------------------------------------------------------------------------
	// TODO: Can we derive this from the project file?
	version := lo.Ternary(opts.Version == "", "4.2.2", opts.Version)
	useMono := false

	if err := downloadGodot(opts.fs, version, useMono); err != nil {
		return fmt.Errorf("failed to configure godot: %w", err)
	}

	//--------------------------------------------------------------------------
	// We have access to a compatible Godot binary and export templates.
	//
	// Build the project.
	//--------------------------------------------------------------------------
	if err := os.MkdirAll(paths.Version(version, useMono), 0o0755); err != nil {
		return err //nolint:wrapcheck
	}

	_ = os.MkdirAll(filepath.Join(lo.Must(os.Getwd()), "examples", "exampleA", "bin"), 0o755)

	binary, err := paths.Binary(version, useMono)
	if err != nil {
		return err //nolint:wrapcheck
	}

	// e, err := exports.New(filepath.Join(filepath.Dir(path), "export_presets.cfg"))

	cmd := exec.Command(binary, "--verbose", "--headless", "--quit", "--export-release", "Windows", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run() //nolint:wrapcheck
}

func findProjectFile(fs afero.Fs, path string) (string, error) {
	const filename = "project.godot"

	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to find CWD: %w", err)
		}

		path = cwd
	}

	// Check if the path is a valid file or directory
	info, err := fs.Stat(path)
	if err != nil {
		return "", err //nolint:wrapcheck
	}

	// If the path is a directory, append the filename to the path
	if info.IsDir() {
		path = filepath.Join(path, filename)
	}

	// Check if the file exists
	exists, err := afero.Exists(fs, path)
	if err != nil {
		return "", err //nolint:wrapcheck
	}

	if !exists {
		return "", errors.New("project file not found")
	}

	// Ensure the file is indeed named 'project.godot'
	if filepath.Base(path) != filename {
		return "", errors.New("path does not point to a 'project.godot' file")
	}

	// Get the absolute path to return
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err //nolint:wrapcheck
	}

	return absPath, nil
}
