package build

import (
	"fmt"

	validator "github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/ruffel/godotreleaser/internal/config"
	"github.com/spf13/cobra"
)

type buildOpts struct {
	config string
}

func NewBuildCmd() *cobra.Command {
	opts := &buildOpts{
		config: "",
	}

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build the Godot project",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runBuild(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.config, "config", "c", "", "Load configuration from file")

	return cmd
}

func runBuild(opts *buildOpts) error {
	cfg, err := loadConfiguration(opts.config)
	if err != nil {
		return err
	}

	fmt.Println("Building project", cfg.ProjectName)

	return nil
}

func loadConfiguration(filepath string) (*config.Config, error) {
	k := koanf.New(".")

	if filepath != "" {
		if err := k.Load(file.Provider(filepath), yaml.Parser()); err != nil {
			return nil, err //nolint:wrapcheck
		}
	}

	var cfg config.Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, err //nolint:wrapcheck
	}

	if err := validator.New().Struct(cfg); err != nil {
		return nil, err //nolint:wrapcheck
	}

	return &cfg, nil
}
