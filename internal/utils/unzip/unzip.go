package unzip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Extract removes the first directory level and extracts the zip archive to the specified destination directory.
func Extract(src string, dst string) error {
	// Create the destination directory if it doesn't exist
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("failed to create dst: %w", err)
	}

	// Open the zip archive
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip archive: %w", err)
	}
	defer r.Close()

	// Extract the contents of the archive
	for _, f := range r.File {
		// Remove the first directory level
		relativePath := f.Name
		if idx := strings.Index(f.Name, "/"); idx != -1 {
			relativePath = f.Name[idx+1:]
		}

		fpath := filepath.Join(dst, relativePath)

		// Check if the file is a directory
		if f.FileInfo().IsDir() {
			// Create the directory
			if err := os.MkdirAll(fpath, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		// Create the file and extract its contents
		if err := func() error {
			// Create the directory structure for the file if needed
			if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
				return fmt.Errorf("failed to create directory for file: %w", err)
			}

			// Open the file inside the archive
			src, err := f.Open()
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer src.Close()

			// Create the target file
			dst, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			defer dst.Close()

			// Copy the contents
			_, err = io.Copy(dst, src)
			if err != nil {
				return fmt.Errorf("failed to copy file: %w", err)
			}

			return nil
		}(); err != nil {
			return err
		}
	}

	return nil
}
