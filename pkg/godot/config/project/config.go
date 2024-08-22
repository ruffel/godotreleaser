package project

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/ruffel/godotreleaser/pkg/godot/config/parser"
	"github.com/samber/lo"
)

type Config struct {
	Version   int          `koanf:"DEFAULT.config_version"`
	Name      string       `koanf:"application.config/name"`
	Features  []string     `koanf:"application.config/features"`
	MainScene string       `koanf:"application.run/main_scene"`
	raw       *koanf.Koanf `koanf:"-"`
}

func (c *Config) ContainsMono() bool {
	// TODO: This is a bit hacky
	return lo.ContainsBy(c.raw.Keys(), func(k string) bool {
		return strings.HasPrefix(k, "dotnet")
	})
}

func (c *Config) EngineVersion() *version.Version {
	// Parse any "versions" found in the features list.
	list := lo.FilterMap(c.Features, func(f string, _ int) (*version.Version, bool) {
		version, err := version.NewVersion(f)

		return version, err == nil
	})

	// If we have any versions, return the highest one.
	if len(list) == 0 {
		return nil
	}

	// Sort the versions and return the highest one.
	sort.Sort(version.Collection(list))

	return list[len(list)-1]
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
