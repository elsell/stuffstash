package ports

import (
	"context"
	"errors"
	"io"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
)

var ErrBlobStreamSize = errors.New("blob stream length is invalid or exceeds its limit")

const MaxSinglePutBytes int64 = 5 * 1024 * 1024 * 1024

type BlobReadStream interface {
	io.Reader
	io.ReaderAt
	io.Closer
}
type BlobStreamWrite struct {
	Key                 media.StorageKey
	ContentType         string
	SizeBytes, MaxBytes int64
	Content             io.ReadSeeker
}
type StreamingBlobStorage interface {
	OpenBlobStream(context.Context, media.StorageKey) (BlobReadStream, int64, error)
	PutBlobStream(context.Context, BlobStreamWrite) error
	DeleteBlob(context.Context, media.StorageKey) error
}

// Closing scratch content removes its private temporary file.
type ArchiveScratch interface {
	io.ReadWriteSeeker
	io.ReaderAt
	io.Closer
}
type ArchiveScratchSpace interface {
	NewArchiveScratch(context.Context) (ArchiveScratch, error)
}
