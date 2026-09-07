package media

import (
	"context"
	"errors"

	domain "github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type Reader struct {
	processor *Processor
	admission ports.ImageWorkAdmission
}

var _ ports.ThumbnailReader = (*Reader)(nil)

func NewReader(processor *Processor, admission ports.ImageWorkAdmission) (*Reader, error) {
	if processor == nil || admission == nil {
		return nil, errors.New("thumbnail reader requires processor and shared admission")
	}
	return &Reader{processor: processor, admission: admission}, nil
}

func (r *Reader) ReadThumbnail(ctx context.Context, attachment domain.Attachment, variant domain.ThumbnailVariant) (ports.ImageDerivative, bool, error) {
	if err := ctx.Err(); err != nil {
		return ports.ImageDerivative{}, false, err
	}
	normalized, valid := domain.NewThumbnailVariant(variant.String())
	if !valid || !attachment.ContentType.IsImage() {
		return ports.ImageDerivative{}, false, errors.New("thumbnail requires an image and valid variant")
	}
	key, err := ThumbnailCacheKey(attachment.StorageKey, normalized)
	if err != nil {
		return ports.ImageDerivative{}, false, err
	}
	for {
		if err := ctx.Err(); err != nil {
			return ports.ImageDerivative{}, false, err
		}
		result, cached, retry, err := r.readAttempt(ctx, attachment, normalized, key)
		if retry {
			continue
		}
		return result, cached, err
	}
}

// Each retry subscribes before its cache lookup, closing flight/publication races.
func (r *Reader) readAttempt(ctx context.Context, attachment domain.Attachment, variant domain.ThumbnailVariant, key domain.StorageKey) (ports.ImageDerivative, bool, bool, error) {
	published, unsubscribe := r.processor.readiness.Watch(key)
	defer unsubscribe()
	if cached, err := readCachedThumbnail(ctx, r.processor.blobs, key); err == nil {
		return cached, true, false, nil
	} else if !errors.Is(err, ports.ErrBlobNotFound) {
		return ports.ImageDerivative{}, false, false, err
	}
	flightRelease, finished, owned := r.processor.flights.TryStart(attachment.StorageKey)
	if !owned {
		select {
		case <-ctx.Done():
			return ports.ImageDerivative{}, false, false, ctx.Err()
		case <-published:
		case <-finished:
		}
		return ports.ImageDerivative{}, false, true, nil
	}
	defer flightRelease()
	release, cached, err := r.waitForThumbnail(ctx, key, published)
	if err != nil {
		return ports.ImageDerivative{}, false, false, err
	}
	if cached != nil {
		return *cached, true, false, nil
	}
	defer release()
	if cached, err := readCachedThumbnail(ctx, r.processor.blobs, key); err == nil {
		return cached, true, false, nil
	} else if !errors.Is(err, ports.ErrBlobNotFound) {
		return ports.ImageDerivative{}, false, false, err
	}
	var result ports.ImageDerivative
	err = r.processor.generate(ctx, attachment, nil, []domain.ThumbnailVariant{variant}, func(derivative ports.ImageDerivative) { result = derivative })
	if err != nil {
		return ports.ImageDerivative{}, false, false, err
	}
	if len(result.Content) == 0 {
		return ports.ImageDerivative{}, false, false, errors.New("thumbnail processor produced no output")
	}
	return result, false, false, nil
}
