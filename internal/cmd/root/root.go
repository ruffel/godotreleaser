package root

import (
	"github.com/ruffel/godotreleaser/internal/cmd/build"
	"github.com/ruffel/godotreleaser/internal/cmd/version"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.AddCommand(build.NewBuildCmd())
	cmd.AddCommand(version.NewCmdVersion())

	return cmd
}
