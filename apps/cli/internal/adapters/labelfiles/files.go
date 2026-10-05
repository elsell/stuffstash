//go:build !windows

package labelfiles

import (
	"context"
	"os"
	"path/filepath"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type Files struct{}

func (Files) Publish(ctx context.Context, path string, content []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".stuffstash-label-*")
	if err != nil {
		return ports.Failure("file", "Cannot create the label file. Check that the output directory exists and that you can write to it.")
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	defer file.Close()
	if _, err = file.Write(content); err == nil {
		err = file.Sync()
	}
	if err == nil {
		err = file.Close()
	}
	if err != nil {
		return ports.Failure("file", "Cannot save the label file. Check available disk space and write access to the output directory. Then repeat the command.")
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	// Same-directory link publishes atomically and never replaces a concurrent file or symlink.
	if err = os.Link(temporary, path); err != nil {
		if os.IsExist(err) {
			return ports.Failure("file", "The output path already exists. Choose a new --output path.")
		}
		return ports.Failure("file", "Cannot save the label at the requested path. Choose a new --output path on a filesystem that supports hard links.")
	}
	return nil
}
