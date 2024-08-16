package client

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ruffel/godotreleaser/internal/paths"
)

type Client struct {
	path string
}

func NewFromPath(path string) (*Client, error) {
	return &Client{
		path: path,
	}, nil
}

func NewFromVersion(version string, mono bool) (*Client, error) {
	path, err := paths.GetBinary(version, mono)
	if err != nil {
		return nil, fmt.Errorf("failed to get binary path: %w", err)
	}

	return NewFromPath(path)
}

type BuildOptions struct {
	Preset     string
	Project    string
	ExportType ExportType
}

func (c *Client) Build(ctx context.Context, opts *BuildOptions) error {
	cleanPath := filepath.Clean(c.path)
	cleanPreset := filepath.Clean(opts.Preset)
	cleanPathArg := filepath.Clean(opts.Project)

	cmd := exec.CommandContext(ctx, cleanPath, "--headless", "--quit", opts.ExportType.String(), cleanPreset, cleanPathArg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build project: %w", err)
	}

	return nil
}

func (c *Client) Run(args ...string) error {
	binary := filepath.Clean(c.path)

	cmd := exec.Command(binary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}

	return nil
}
