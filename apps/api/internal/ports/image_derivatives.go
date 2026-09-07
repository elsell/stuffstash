package ports

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
)

type ImageDerivativesRequest struct {
	Attachment  media.Attachment
	ContentType media.ContentType
	Content     []byte
	Variants    []media.ThumbnailVariant
}

// ImageBatchProcessor decodes once and publishes each requested size as it becomes
// ready. Publishers must consume the derivative before returning; errors stop work.
type ImageBatchProcessor interface {
	CreateThumbnails(context.Context, ImageDerivativesRequest, func(media.ThumbnailVariant, ImageDerivative) error) error
}
