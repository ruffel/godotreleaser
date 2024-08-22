package build

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/ruffel/godotreleaser/internal/stages/builder"
	"github.com/ruffel/godotreleaser/internal/stages/dependencies"
	"github.com/ruffel/godotreleaser/internal/terminal"
	"github.com/ruffel/godotreleaser/internal/terminal/messages"
	"github.com/ruffel/godotreleaser/pkg/godot/config/project"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type buildOpts struct {
	ProjectDir string
	Version    string
	Mono       bool
	MonoSet    bool
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			// TODO: Ehhhh... this is a bit hacky
			opts.MonoSet = cmd.Flags().Changed("with-mono")

			return runBuild(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.ProjectDir, "project", "p", "", "Path to the Godot project directory (defaults to the current directory)")
	cmd.Flags().StringVarP(&opts.Version, "version", "v", "", "Godot version to use")
	cmd.Flags().BoolVar(&opts.Mono, "with-mono", false, "Mono version of Godot")

	return cmd
}

//nolint:funlen
func runBuild(ctx context.Context, opts *buildOpts) error {
	terminal.Send(messages.NewSequence("Building Godot Project"))

	path, err := findProjectFile(opts.fs, opts.ProjectDir)
	if err != nil {
		return fmt.Errorf("failed to find project file: %w", err)
	}

	slog.Debug("Found Godot project file", "path", path)

	project, err := project.New(path)
	if err != nil {
		return fmt.Errorf("project file is not valid: %w", err)
	}

	//--------------------------------------------------------------------------
	// We need a Godot binary and export templates to build the project.
	//
	// Download the Godot binary and export templates if they don't exist.
	//--------------------------------------------------------------------------
	version := func() string {
		if opts.Version != "" {
			slog.Debug("Using version provided by user", "version", opts.Version)

			return opts.Version
		}

		if project.EngineVersion() != nil {
			version := project.EngineVersion().Original()

			slog.Debug("Using version found in project.godot", "version", version)

			return version
		}

		slog.Debug("Using default version", "version", "4.3")

		return "4.3" // Or error?
	}()

	useMono := func() bool {
		if opts.MonoSet {
			return opts.Mono
		}

		return project.ContainsMono()
	}()

	if err := dependencies.Run(ctx, opts.fs, version, useMono); err != nil {
		return err //nolint:wrapcheck
	}

	if err := builder.Run(ctx, opts.fs, version, useMono, path); err != nil {
		return err //nolint:wrapcheck
	}

	terminal.Send(messages.NewFooter("Project Built"))

	return nil
}

var ErrProjectFileNotFound = errors.New("project.godot file not found")

func findProjectFile(fs afero.Fs, path string) (string, error) {
	const filename = "project.godot"

	// Iterate through each path and search for the project file
	for _, searchPath := range determineSearchPaths(path) {
		projectFilePath, err := checkProjectFile(fs, searchPath, filename)
		if err != nil {
			return "", err
		}

		if projectFilePath != "" {
			return resolveAbsolutePath(projectFilePath)
		}
	}

	return "", ErrProjectFileNotFound
}

func checkProjectFile(fs afero.Fs, basePath, filename string) (string, error) {
	// Check if the provided basePath is a file or a directory
	info, err := fs.Stat(basePath)
	if err != nil {
		return "", fmt.Errorf("failed to stat path %s: %w", basePath, err)
	}

	if info.IsDir() {
		// If it's a directory, check for the project.godot file inside it
		projectPath := filepath.Join(basePath, filename)
		slog.Debug("Checking directory for project file", "path", projectPath)

		exists, err := afero.Exists(fs, projectPath)
		if err != nil {
			return "", fmt.Errorf("failed to check if project file exists in directory %s: %w", basePath, err)
		}

		if exists {
			return projectPath, nil
		}
	} else {
		slog.Debug("Checking if the path is the project file", "path", basePath)

		if filepath.Base(basePath) == filename {
			return basePath, nil
		}
	}

	return "", nil
}

func determineSearchPaths(path string) []string {
	idiomaticPaths := []string{
		"/app",
		"/workspaces",
		"/src",
		"/code",
	}

	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			slog.Warn("Failed to find CWD, cannot use as search path", "error", err)

			return idiomaticPaths
		}

		return append([]string{cwd}, idiomaticPaths...)
	}

	return []string{path}
}

func resolveAbsolutePath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	if filepath.Base(absPath) != "project.godot" {
		return "", fmt.Errorf("path does not point to a 'project.godot' file: %s", absPath)
	}

	return absPath, nil
}
