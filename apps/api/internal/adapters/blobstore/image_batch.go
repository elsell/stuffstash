package blobstore

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/jpeg"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

var _ ports.ImageBatchProcessor = StandardImageProcessor{}

func (StandardImageProcessor) CreateThumbnails(ctx context.Context, request ports.ImageDerivativesRequest, publish func(media.ThumbnailVariant, ports.ImageDerivative) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if publish == nil || !request.ContentType.IsImage() || len(request.Content) == 0 || len(request.Variants) == 0 {
		return errors.New("thumbnail batch requires source, variants and publisher")
	}
	requested := make(map[media.ThumbnailVariant]bool, len(request.Variants))
	for _, variant := range request.Variants {
		switch variant {
		case media.ThumbnailVariantSmall, media.ThumbnailVariantMedium, media.ThumbnailVariantLarge:
		default:
			return errors.New("invalid thumbnail variant")
		}
		if requested[variant] {
			return errors.New("duplicate thumbnail variant")
		}
		requested[variant] = true
	}
	if err := validateImageBounds(request.Content); err != nil {
		return err
	}
	source, _, err := image.Decode(bytes.NewReader(request.Content))
	if err != nil {
		return err
	}
	for _, variant := range []media.ThumbnailVariant{media.ThumbnailVariantSmall, media.ThumbnailVariantMedium, media.ThumbnailVariantLarge} {
		if !requested[variant] {
			continue
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		thumbnail := resizeImage(source, thumbnailMaxDimension(variant))
		if err := ctx.Err(); err != nil {
			return err
		}
		var output bytes.Buffer
		if err := jpeg.Encode(&output, thumbnail, &jpeg.Options{Quality: thumbnailJPEGQuality}); err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := publish(variant, ports.ImageDerivative{ContentType: media.ContentTypeJPEG, Content: output.Bytes()}); err != nil {
			return err
		}
	}
	return ctx.Err()
}
