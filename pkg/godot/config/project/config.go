package project

import (
	"fmt"

	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/ruffel/godotreleaser/pkg/godot/config/parser"
)

type Config struct {
	Version   int          `koanf:"DEFAULT.config_version"`
	Name      string       `koanf:"application.config/name"`
	Features  []string     `koanf:"application.config/features"`
	MainScene string       `koanf:"application.run/main_scene"`
	raw       *koanf.Koanf `koanf:"-"`
}

func New(path string) (*Config, error) {
	k := koanf.New(".")

	if err := k.Load(file.Provider(path), parser.Godot{}); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	var config *Config

	if err := k.UnmarshalWithConf("", &config, koanf.UnmarshalConf{Tag: "koanf", FlatPaths: true}); err != nil {
		return nil, err //nolint:wrapcheck
	}

	config.raw = k

	return config, nil
}

func (c *Config) ProjectName() string {
	return c.Name
}
