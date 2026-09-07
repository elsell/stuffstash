package media

import (
	"context"
	"errors"
	"time"

	domain "github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type readerAdmissionResult struct {
	release func()
	err     error
}

// Keep one foreground waiter queued while observing incremental publication.
// Cache checks never download the original or consume image-processing capacity.
func (r *Reader) waitForThumbnail(ctx context.Context, key domain.StorageKey) (func(), *ports.ImageDerivative, error) {
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
	ticker := time.NewTicker(r.cachePollInterval)
	defer ticker.Stop()
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
		case <-ticker.C:
			// A slow storage probe must not hold a concurrently granted permit
			// indefinitely. A probe timeout is inconclusive, not a read failure.
			probe, stop := context.WithTimeout(ctx, r.cachePollInterval)
			cached, err := readCachedThumbnail(probe, r.processor.blobs, key)
			stop()
			if ctx.Err() != nil {
				return nil, nil, ctx.Err()
			}
			if err == nil {
				return nil, &cached, nil
			}
			if !errors.Is(err, ports.ErrBlobNotFound) && !errors.Is(err, context.DeadlineExceeded) {
				return nil, nil, err
			}
		}
	}
}
