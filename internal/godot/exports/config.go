package exports

import (
	"fmt"

	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/ruffel/godotreleaser/internal/godot/parser"
)

type Config struct {
	raw map[string]interface{}
}

func New(path string) (*Config, error) {
	k := koanf.New(".")

	if err := k.Load(file.Provider(path), parser.ProjectParser{}); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &Config{
		raw: k.Raw(),
	}, nil
}
