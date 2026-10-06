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
			if os.IsExist(err) {
				return ports.Failure("file", "The output path already exists. Choose another --output path. The existing file was not changed.")
			}
			return ports.Failure("file", "Cannot publish the file. Choose a new output path on a filesystem that supports hard links.")
		}
		return nil
	}, nil
}
