//go:build !windows

package binaryfiles

import (
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"os"
	"path/filepath"
)

func createPrivate(path string) (*os.File, func() error, error) {
	f, err := os.CreateTemp(filepath.Dir(path), ".stuffstash-download-*")
	if err != nil {
		return nil, nil, ports.Failure("file", "Cannot create the file. Check the output directory and write permissions.")
	}
	return f, func() error {
		if err := os.Link(f.Name(), path); err != nil {
			return ports.Failure("file", "Cannot publish the file. Choose a new output path on a filesystem that supports hard links.")
		}
		return nil
	}, nil
}
