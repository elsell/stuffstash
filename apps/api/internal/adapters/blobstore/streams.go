package blobstore

import (
	"context"
	"errors"
	"io"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

// Validate length from the seekable private source before publication. The caller
// owns it exclusively throughout the write; producers must close before calling.
func validateStream(ctx context.Context, input ports.BlobStreamWrite) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if input.Content == nil || input.Key.String() == "" || input.ContentType == "" || input.SizeBytes < 0 || input.MaxBytes <= 0 || input.SizeBytes > input.MaxBytes || input.SizeBytes > ports.MaxSinglePutBytes {
		return ports.ErrBlobStreamSize
	}
	size, err := input.Content.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	if size != input.SizeBytes {
		return ports.ErrBlobStreamSize
	}
	_, err = input.Content.Seek(0, io.SeekStart)
	return err
}

type streamContextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r streamContextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}

func copyExactStream(ctx context.Context, w io.Writer, input ports.BlobStreamWrite) error {
	n, err := io.Copy(w, io.LimitReader(streamContextReader{ctx: ctx, r: input.Content}, input.SizeBytes+1))
	if err != nil {
		return err
	}
	if n != input.SizeBytes {
		return ports.ErrBlobStreamSize
	}
	return ctx.Err()
}

var errScratchClosed = errors.New("archive scratch is already closed")
