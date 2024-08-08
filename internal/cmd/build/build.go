package build

import (
	"os"
	"os/exec"
	"path/filepath"

	project "github.com/ruffel/godotreleaser/internal/godot/project"
	"github.com/ruffel/godotreleaser/internal/paths"
	"github.com/samber/lo"
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
	useMono := false   // TODO: We can probably derive this requirement from the project file.

	//--------------------------------------------------------------------------
	// We have access to a compatible Godot binary and export templates.
	//
	// Build the project.
	//--------------------------------------------------------------------------
	if err := os.MkdirAll(paths.Version(version, useMono), 0755); err != nil { //nolint:mnd
		return err //nolint:wrapcheck
	}

	os.MkdirAll(filepath.Join(lo.Must(os.Getwd()), "examples", "exampleA", "bin"), 0o755)

	// if err := downloadGodot(version, useMono); err != nil {
	// 	return err //nolint:wrapcheck
	// }

	project := filepath.Join(opts.ProjectDir, "project.godot")
	binary := filepath.Join(paths.Version(version, useMono), "godot")

	cmd := exec.Command(binary, "--verbose", "--headless", "--quit", "--export-release", "Windows", project) //nolint:lll
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run() //nolint:wrapcheck
}
