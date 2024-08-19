package root

import (
	"log/slog"
	"os"
	"time"

	charmlog "github.com/charmbracelet/log"
	"github.com/ruffel/godotreleaser/internal/cmd/build"
	"github.com/ruffel/godotreleaser/internal/cmd/dependencies"
	"github.com/ruffel/godotreleaser/internal/cmd/version"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			slog.SetDefault(slog.New(charmlog.NewWithOptions(os.Stderr, charmlog.Options{
				ReportCaller:    true,
				ReportTimestamp: true,
				TimeFormat:      time.Kitchen,
				Level:           charmlog.DebugLevel,
			})))

			return nil
		},
	}

	cmd.AddCommand(build.NewBuildCmd())
	cmd.AddCommand(version.NewCmdVersion())
	cmd.AddCommand(dependencies.NewDependenciesCmd())

	return cmd
}
