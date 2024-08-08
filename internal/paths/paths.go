package paths

import (
	"os"
	"path/filepath"

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
