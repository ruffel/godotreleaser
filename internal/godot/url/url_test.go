package url

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//nolint:funlen
func Test_selectTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
		mono    bool
		goos    string
		arch    string
		want    string
	}{
		{
			name:    "macos",
			version: "4.2.2",
			mono:    false,
			goos:    "darwin",
			arch:    "amd64",
			want:    "Godot_v4.2.2-stable_macos.universal.zip",
		},
		{
			name:    "darwin-arm64",
			version: "4.2.2",
			mono:    false,
			goos:    "darwin",
			arch:    "arm64",
			want:    "Godot_v4.2.2-stable_macos.universal.zip",
		},
		{
			name:    "darwin-arm-mono",
			version: "4.2.2",
			mono:    true,
			goos:    "darwin",
			arch:    "arm",
			want:    "mono/Godot_v4.2.2-stable_mono_macos.universal.zip",
		},
		{
			name:    "linux-arm",
			version: "4.2.2",
			mono:    false,
			goos:    "linux",
			arch:    "arm",
			want:    "Godot_v4.2.2-stable_linux.arm32.zip",
		},
		{
			name:    "linux-arm64",
			version: "4.2.2",
			mono:    false,
			goos:    "linux",
			arch:    "arm64",
			want:    "Godot_v4.2.2-stable_linux.arm64.zip",
		},
		{
			name:    "linux-amd64",
			version: "4.2.2",
			mono:    false,
			goos:    "linux",
			arch:    "amd64",
			want:    "Godot_v4.2.2-stable_linux.x86_64.zip",
		},
		{
			name:    "linux-386",
			version: "4.2.2",
			mono:    false,
			goos:    "linux",
			arch:    "386",
			want:    "Godot_v4.2.2-stable_linux.x86_32.zip",
		},
		{
			name:    "linux-arm-mono",
			version: "4.2.2",
			mono:    true,
			goos:    "linux",
			arch:    "arm",
			want:    "mono/Godot_v4.2.2-stable_mono_linux_arm32.zip",
		},
		{
			name:    "windows-amd64",
			version: "4.2.2",
			mono:    false,
			goos:    "windows",
			arch:    "amd64",
			want:    "Godot_v4.2.2-stable_win64.exe.zip",
		},
		{
			name:    "windows-386",
			version: "4.2.2",
			mono:    false,
			goos:    "windows",
			arch:    "386",
			want:    "Godot_v4.2.2-stable_win32.exe.zip",
		},
		{
			name:    "windows-amd64-mono",
			version: "4.2.2",
			mono:    true,
			goos:    "windows",
			arch:    "amd64",
			want:    "mono/Godot_v4.2.2-stable_mono_win64.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, _ := selectTemplate(tt.version, tt.mono, tt.goos, tt.arch)
			assert.Equal(t, tt.want, got)
		})
	}
}
