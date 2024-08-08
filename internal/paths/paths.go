package paths

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/samber/lo"
)

func Root() string {
	return filepath.Join(lo.Must(os.UserConfigDir()), ".godotreleaser")
}

func Cache() string {
	return filepath.Join(Root(), "cache")
}

func Version(version string, mono bool) string {
	if mono {
		return filepath.Join(Cache(), version+"-mono")
	}

	return filepath.Join(Cache(), version)
}

func templateRoot() (string, error) {
	var dir string

	switch runtime.GOOS {
	case "windows":
		dir = os.Getenv("AppData")
		if dir == "" {
			return "", errors.New("%AppData% is not defined")
		}

	case "darwin", "ios":
		dir = os.Getenv("HOME")
		if dir == "" {
			return "", errors.New("$HOME is not defined")
		}

		dir += "/Library/Application Support"

	case "plan9":
		dir = os.Getenv("home")
		if dir == "" {
			return "", errors.New("$home is not defined")
		}

		dir += "/lib"

	default: // Unix
		dir = os.Getenv("HOME")
		if dir == "" {
			return "", errors.New("$HOME is not defined")
		}

		dir = filepath.Join(dir, ".local", "share")
	}

	return dir, nil
}

func TemplatePath(version string, mono bool) string {
	root := lo.Must(templateRoot())
	name := lo.Ternary(runtime.GOOS == "linux", "godot", "Godot")
	base := fmt.Sprintf("%s.stable%s", version, lo.Ternary(mono, ".mono", ""))

	return filepath.Join(root, name, "export_templates", base)
}
