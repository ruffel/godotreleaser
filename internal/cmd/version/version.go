package version

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var version = "0.0.0-unset"

type versionOptions struct {
	Short bool
}

func NewCmdVersion() *cobra.Command {
	opts := &versionOptions{
		Short: false,
	}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version and build information",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runVersion(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Short, "short", "s", opts.Short, "Print just the version number.")

	return cmd
}

func runVersion(opts *versionOptions) error {
	if opts.Short {
		fmt.Println(version)

		return nil
	}

	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Commit: %s\n", getCommitSHA())
	fmt.Printf("Date: %s\n", getCommitDate())

	return nil
}

func getCommitSHA() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	for _, settings := range info.Settings {
		if settings.Key == "vcs.revision" {
			return settings.Value
		}
	}

	return ""
}

func getCommitDate() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	for _, settings := range info.Settings {
		if settings.Key == "vcs.time" {
			return settings.Value
		}
	}

	return ""
}
