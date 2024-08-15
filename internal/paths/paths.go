package paths

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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

// CheckBinaryExists checks if the Godot binary exists in the specified version directory.
func CheckBinaryExists(version string, mono bool) (bool, error) {
	dirPath := Version(version, mono)

	// Check if the directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return false, nil
	}

	var godotBinaryFound bool

	walkFn := func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasPrefix(info.Name(), "Godot") {
			godotBinaryFound = true

			return filepath.SkipDir
		}

		return nil
	}

	if err := filepath.Walk(dirPath, walkFn); err != nil {
		return false, err //nolint:wrapcheck
	}

	return godotBinaryFound, nil
}

// GetBinary retrieves the Godot binary's path for the specified version.
func GetBinary(version string, mono bool) (string, error) {
	dirPath := Version(version, mono)

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return "", errors.New("directory does not exist, please download the binary")
	}

	var godotBinary string

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasPrefix(info.Name(), "Godot") {
			godotBinary = path

			return filepath.SkipDir
		}

		return nil
	}

	if err := filepath.Walk(dirPath, walkFn); err != nil {
		return "", err //nolint:wrapcheck
	}

	if godotBinary == "" {
		return "", errors.New("godot binary not found")
	}

	return godotBinary, nil
}
