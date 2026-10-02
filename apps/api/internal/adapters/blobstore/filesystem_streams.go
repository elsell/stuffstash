package blobstore

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

var _ ports.StreamingBlobStorage = FileSystemStore{}

func (s FileSystemStore) OpenBlobStream(ctx context.Context, key media.StorageKey) (ports.BlobReadStream, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	path, err := s.pathForKey(key)
	if err != nil {
		return nil, 0, err
	}
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, 0, ports.ErrBlobNotFound
	}
	if err != nil {
		return nil, 0, err
	}
	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, 0, err
	}
	if !stat.Mode().IsRegular() {
		f.Close()
		return nil, 0, ports.ErrBlobStreamSize
	}
	return f, stat.Size(), nil
}
func (s FileSystemStore) PutBlobStream(ctx context.Context, input ports.BlobStreamWrite) error {
	if err := validateStream(ctx, input); err != nil {
		return err
	}
	path, err := s.pathForKey(input.Key)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".archive-stage-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if err = copyExactStream(ctx, f, input); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

type ScratchSpace struct{ Directory string }

func (s ScratchSpace) NewArchiveScratch(ctx context.Context) (ports.ArchiveScratch, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	file, err := os.CreateTemp(s.Directory, "stuffstash-archive-")
	if err != nil {
		return nil, err
	}
	// Unix keeps the descriptor valid until close and reclaims it on process exit.
	// Never leave a named private archive behind for a crash-recovery sweep.
	if err = os.Remove(file.Name()); err != nil {
		return nil, errors.Join(err, file.Close())
	}
	return &scratchFile{File: file}, nil
}

type scratchFile struct {
	*os.File
	closed bool
}

func (f *scratchFile) Close() error {
	if f.closed {
		return errScratchClosed
	}
	f.closed = true
	return f.File.Close()
}
