package exports

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/ruffel/godotreleaser/pkg/godot/config/parser"
	"github.com/samber/lo"
)

// PresetCollection represents a collection of Preset configurations.
type PresetCollection []Preset

// Preset represents an individual configuration preset.
type Preset struct {
	Name                     string        `koanf:"name"`
	Platform                 string        `koanf:"platform"`
	Runnable                 bool          `koanf:"runnable"`
	DedicatedServer          bool          `koanf:"dedicated_server"`
	CustomFeatures           string        `koanf:"custom_features"`
	IncludeFilter            string        `koanf:"include_filter"`
	ExcludeFilter            string        `koanf:"exclude_filter"`
	ExportPath               string        `koanf:"export_path"`
	EncryptionIncludeFilters string        `koanf:"encryption_include_filters"`
	EncryptionExcludeFilters string        `koanf:"encryption_exclude_filters"`
	EncryptPck               bool          `koanf:"encrypt_pck"`
	EncryptDirectory         bool          `koanf:"encrypt_directory"`
	Options                  PresetOptions `koanf:"-"`
}

// PresetOptions contains various configuration options for the Preset.
type PresetOptions struct {
	CustomTemplateDebug      string `koanf:"custom_template/debug"`
	BinaryFormatArchitecture string `koanf:"binary_format/architecture"`
}

// Config holds the loaded configuration and its presets.
type Config struct {
	presets PresetCollection
	data    *koanf.Koanf
}

// Presets returns the collection of loaded presets.
func (c *Config) Presets() PresetCollection {
	return c.presets
}

// PresetNames returns a list of all preset names.
func (c *Config) PresetNames() []string {
	return lo.Map(c.presets, func(preset Preset, _ int) string {
		return preset.Name
	})
}

// New loads the configuration from the specified file and returns a Config object.
func New(path string) (*Config, error) {
	k := koanf.New(".")

	// Load configuration file using the specified parser.
	if err := k.Load(file.Provider(path), parser.Godot{}); err != nil {
		return nil, fmt.Errorf("failed to load config file %s: %w", path, err)
	}

	// Parse the presets from the configuration.
	presets, err := loadPresets(k)
	if err != nil {
		return nil, fmt.Errorf("failed to load presets: %w", err)
	}

	return &Config{
		presets: presets,
		data:    k,
	}, nil
}

// loadPresets loads and parses the preset configurations from the Koanf instance.
func loadPresets(k *koanf.Koanf) (PresetCollection, error) {
	var presets PresetCollection

	for i := 0; ; i++ {
		key := fmt.Sprintf("preset.%d", i)
		if !k.Exists(key) {
			break // Exit loop when no more presets are found.
		}

		var preset Preset
		if err := k.Unmarshal(key, &preset); err != nil {
			slog.Warn("Failed to unmarshal preset", "index", i, "error", err)

			continue
		}

		// Load and attach preset options if available.
		optionsKey := key + ".options"

		if k.Exists(optionsKey) {
			var options PresetOptions
			if err := k.Unmarshal(optionsKey, &options); err != nil {
				slog.Warn("Failed to unmarshal preset options", "index", i, "error", err)
			} else {
				preset.Options = options
			}
		}

		presets = append(presets, preset)
	}

	if len(presets) == 0 {
		return nil, errors.New("no valid presets found")
	}

	return presets, nil
}
