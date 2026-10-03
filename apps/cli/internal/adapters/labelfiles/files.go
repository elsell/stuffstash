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
		return ports.Failure("file", "cannot create label output file")
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
		return ports.Failure("file", "could not write label output")
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	// Same-directory link publishes atomically and never replaces a concurrent file or symlink.
	if err = os.Link(temporary, path); err != nil {
		return ports.Failure("file", "cannot publish label; output must be a new path on a filesystem supporting hard links")
	}
	return nil
}
