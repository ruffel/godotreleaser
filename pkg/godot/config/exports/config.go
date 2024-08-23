package exports

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
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
	CustomFeatures           []string      `koanf:"custom_features"`
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
	CustomTemplateDebug           string   `koanf:"custom_template/debug"`
	CustomTemplateRelease         string   `koanf:"custom_template/release"`
	DebugExportConsoleWrapper     int      `koanf:"debug_export/console_wrapper"`
	BinaryFormatArchitecture      string   `koanf:"binary_format/architecture"`
	BinaryFormatEmbedPCK          bool     `koanf:"binary_format/embed_pck"`
	TextureFormatBPTC             bool     `koanf:"texture_format/bptc"`
	TextureFormatS3TC             bool     `koanf:"texture_format/s3tc"`
	TextureFormatETC              bool     `koanf:"texture_format/etc"`
	TextureFormatETC2             bool     `koanf:"texture_format/etc2"`
	CodeSignEnable                bool     `koanf:"codesign/enable"`
	CodeSignTimestamp             bool     `koanf:"codesign/timestamp"`
	CodeSignTimestampServerURL    string   `koanf:"codesign/timestamp_server_url"`
	CodeSignDigestAlgorithm       int      `koanf:"codesign/digest_algorithm"`
	CodeSignDescription           string   `koanf:"codesign/description"`
	CodeSignCustomOptions         []string `koanf:"codesign/custom_options"`
	ApplicationModifyResources    bool     `koanf:"application/modify_resources"`
	ApplicationIcon               string   `koanf:"application/icon"`
	ApplicationConsoleWrapperIcon string   `koanf:"application/console_wrapper_icon"`
	ApplicationIconInterpolation  int      `koanf:"application/icon_interpolation"`
	ApplicationFileVersion        string   `koanf:"application/file_version"`
	ApplicationProductVersion     string   `koanf:"application/product_version"`
	ApplicationCompanyName        string   `koanf:"application/company_name"`
	ApplicationProductName        string   `koanf:"application/product_name"`
	ApplicationFileDescription    string   `koanf:"application/file_description"`
	ApplicationCopyright          string   `koanf:"application/copyright"`
	ApplicationTrademarks         string   `koanf:"application/trademarks"`
	ApplicationExportAngle        int      `koanf:"application/export_angle"`
	SSHRemoteDeployEnabled        bool     `koanf:"ssh_remote_deploy/enabled"`
	SSHRemoteDeployHost           string   `koanf:"ssh_remote_deploy/host"`
	SSHRemoteDeployPort           string   `koanf:"ssh_remote_deploy/port"`
	SSHRemoteDeployExtraArgsSSH   string   `koanf:"ssh_remote_deploy/extra_args_ssh"`
	SSHRemoteDeployExtraArgsSCP   string   `koanf:"ssh_remote_deploy/extra_args_scp"`
	SSHRemoteDeployRunScript      string   `koanf:"ssh_remote_deploy/run_script"`
	SSHRemoteDeployCleanupScript  string   `koanf:"ssh_remote_deploy/cleanup_script"`
	DotNetIncludeScriptsContent   bool     `koanf:"dotnet/include_scripts_content"`
	DotNetIncludeDebugSymbols     bool     `koanf:"dotnet/include_debug_symbols"`
	DotNetEmbedBuildOutputs       bool     `koanf:"dotnet/embed_build_outputs"`
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

func (c *Config) Render() (string, error) {
	// Use a dummy separator to avoid any conflicts with the preset keys.
	k := koanf.New("__DUMMY__")

	// Iterate over the presets and add them to the Koanf instance.
	for i, preset := range c.presets {
		p := koanf.New(".")
		if err := p.Load(structs.Provider(preset, "koanf"), nil); err != nil {
			slog.Error("Failed to load preset", "index", i, "error", err)

			return "", err //nolint:wrapcheck
		}

		if err := k.Set(fmt.Sprintf("preset.%d", i), p.All()); err != nil {
			slog.Error("Failed to set preset", "index", i, "error", err)

			return "", err //nolint:wrapcheck
		}

		o := koanf.New(".")
		if err := o.Load(structs.Provider(preset.Options, "koanf"), nil); err != nil {
			slog.Error("Failed to load preset options", "index", i, "error", err)

			return "", err //nolint:wrapcheck
		}

		if err := k.Set(fmt.Sprintf("preset.%d.options", i), o.All()); err != nil {
			slog.Error("Failed to set preset options", "index", i, "error", err)

			return "", err //nolint:wrapcheck
		}
	}

	data, err := k.Marshal(parser.Godot{})
	if err != nil {
		return "", fmt.Errorf("failed to marshal configuration: %w", err)
	}

	return string(data), nil
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
