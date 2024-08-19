package dependencies

import (
	"context"

	"github.com/ruffel/godotreleaser/internal/stages/dependencies"
	"github.com/ruffel/godotreleaser/internal/terminal"
	"github.com/ruffel/godotreleaser/internal/terminal/messages"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type dependenciesOpts struct {
	Version string
	Mono    bool
	fs      afero.Fs
}

func NewDependenciesCmd() *cobra.Command {
	opts := &dependenciesOpts{
		fs: afero.NewOsFs(),
	}

	cmd := &cobra.Command{
		Use:   "dependencies",
		Short: "Install Godot dependencies",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDependencies(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Version, "version", "v", "4.2.2", "Godot version to use")
	cmd.Flags().BoolVar(&opts.Mono, "with-mono", false, "Mono version of Godot")

	return cmd
}

func runDependencies(ctx context.Context, opts *dependenciesOpts) error {
	terminal.Send(messages.NewSequence("Fetching Godot dependencies"))

	if err := dependencies.Run(ctx, opts.fs, opts.Version, opts.Mono); err != nil {
		return err //nolint:wrapcheck
	}

	terminal.Send(messages.NewFooter("Godot dependencies installed"))

	return nil
}
