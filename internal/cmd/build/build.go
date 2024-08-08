package build

import (
	"os"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	project "github.com/ruffel/godotreleaser/internal/godot/project"
	"github.com/ruffel/godotreleaser/internal/paths"
	"github.com/spf13/cobra"
)

type buildOpts struct {
	ProjectDir string
}

func NewBuildCmd() *cobra.Command {
	opts := &buildOpts{}

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build the Godot project",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runBuild(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.ProjectDir, "project", "p", "", "Path to the Godot project directory (defaults to the current directory)") //nolint:lll

	return cmd
}

func runBuild(opts *buildOpts) error {
	//--------------------------------------------------------------------------
	// Check that we have a valid project directory.
	//--------------------------------------------------------------------------
	if opts.ProjectDir == "" {
		// TODO: Handle this error.
		opts.ProjectDir, _ = os.Getwd()
	}

	_, err := project.New(filepath.Join(opts.ProjectDir, "project.godot"))
	if err != nil {
		return err //nolint:wrapcheck
	}

	//--------------------------------------------------------------------------
	// We need a Godot binary and export templates to build the project.
	//
	// Download the Godot binary and export templates if they don't exist.
	//--------------------------------------------------------------------------
	version := "4.2.2" // TODO: Can we derive this from the project file?
	useMono := true    // TODO: We can probably derive this requirement from the project file.

	//--------------------------------------------------------------------------
	// We have access to a compatible Godot binary and export templates.
	//
	// Build the project.
	//--------------------------------------------------------------------------
	if err := os.MkdirAll(paths.Version(version, useMono), 0755); err != nil { //nolint:mnd
		return err //nolint:wrapcheck
	}

	spew.Dump(paths.Version(version, useMono))

	if err := downloadGodot(version, useMono); err != nil {
		return err //nolint:wrapcheck
	}

	return nil

}
