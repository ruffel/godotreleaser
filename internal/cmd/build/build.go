package build

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

type buildOpts struct{}

func NewBuildCmd() *cobra.Command {
	opts := &buildOpts{}

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build the Godot project",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runBuild(opts)
		},
	}

	return cmd
}

func runBuild(_ *buildOpts) error {
	if err := os.MkdirAll("/app/bin", 0o0755); err != nil { //nolint:mnd
		return err //nolint:wrapcheck
	}

	cmd := exec.Command("godot", "--verbose", "--headless", "--quit", "--export-release", "Windows", "/app/project.godot") //nolint:lll
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run() //nolint:wrapcheck
}
