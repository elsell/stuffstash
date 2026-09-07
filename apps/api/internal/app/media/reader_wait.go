package media

import (
	"context"
	"errors"

	domain "github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type readerAdmissionResult struct {
	release func()
	err     error
}

// Keep one foreground waiter queued while observing incremental publication.
// Cache checks never download the original or consume image-processing capacity.
func (r *Reader) waitForThumbnail(ctx context.Context, key domain.StorageKey, published <-chan struct{}) (func(), *ports.ImageDerivative, error) {
	waiting, cancel := context.WithCancel(ctx)
	admitted := make(chan readerAdmissionResult, 1)
	go func() {
		release, err := r.admission.Acquire(waiting, ports.ImageWorkForeground)
		admitted <- readerAdmissionResult{release: release, err: err}
	}()
	received := false
	defer func() {
		cancel()
		if !received {
			result := <-admitted
			if result.release != nil {
				result.release()
			}
		}
	}()
	for {
		select {
		case result := <-admitted:
			received = true
			if result.err == nil {
				result.err = ctx.Err()
			}
			if result.err != nil {
				if result.release != nil {
					result.release()
				}
				return nil, nil, result.err
			}
			return result.release, nil, nil
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-published:
			published = nil // One hint; a deleted cache must not cause a busy loop.
			cached, err := readCachedThumbnail(ctx, r.processor.blobs, key)
			if ctx.Err() != nil {
				return nil, nil, ctx.Err()
			}
			if err == nil {
				return nil, &cached, nil
			}
			if !errors.Is(err, ports.ErrBlobNotFound) {
				return nil, nil, err
			}
		}
	}
}
