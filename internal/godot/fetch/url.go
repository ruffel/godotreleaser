package fetch

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/samber/lo"
)

const (
	BaseURL = "https://downloads.tuxfamily.org/godotengine"

	ExportTemplate     = "Godot_v%s-stable_export_templates.tpz"
	BinaryTemplate     = "Godot_v%s-stable_%s.zip"
	ExportMonoTemplate = "Godot_v%s-stable_mono_export_templates.tpz"
	BinaryMonoTemplate = "Godot_v%s-stable_mono_%s.zip"
)

var archMap = map[string]map[string]string{
	"darwin":  {"any": "macos.universal"},
	"linux":   {"arm": "arm32", "arm64": "arm64", "amd64": "x86_64", "386": "x86_32"},
	"windows": {"amd64": "win64", "386": "win32"},
}

func selectTemplate(version string, mono bool, goos, arch string) (string, error) {
	template := lo.Ternary(mono, BinaryMonoTemplate, BinaryTemplate)

	var osArch string

	switch goos {
	case "darwin":
		osArch = "macos.universal"
	case "linux":
		linuxArch, ok := archMap["linux"][arch]
		if !ok {
			return "", fmt.Errorf("unsupported architecture for Linux: %s", arch)
		}

		osArch = "linux_" + linuxArch
	case "windows":
		windowsArch, ok := archMap["windows"][arch]
		if !ok {
			return "", fmt.Errorf("unsupported architecture for Windows: %s", arch)
		}

		osArch = windowsArch
		if !mono {
			osArch += ".exe"
		}
	default:
		return "", fmt.Errorf("unsupported OS: %s", goos)
	}

	return fmt.Sprintf(template, version, osArch), nil
}

// BuildBinaryURL constructs the download URL for the specified Godot version and mono flag.
func BuildBinaryURL(version string, mono bool) (string, error) {
	if version == "" {
		return "", errors.New("version cannot be empty")
	}

	template, err := selectTemplate(version, mono, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", BaseURL, template), nil
}

// BuildTemplateURL constructs the template download URL for the specified Godot version and mono flag.
func BuildTemplateURL(version string, mono bool) (string, error) {
	if version == "" {
		return "", errors.New("version cannot be empty")
	}

	template := lo.Ternary(mono, ExportMonoTemplate, ExportTemplate)

	return fmt.Sprintf("%s/%s", BaseURL, fmt.Sprintf(template, version)), nil
}
