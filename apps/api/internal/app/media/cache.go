package media

import (
	"context"
	"errors"

	domain "github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func ThumbnailCacheKey(original domain.StorageKey, variant domain.ThumbnailVariant) (domain.StorageKey, error) {
	normalized, valid := domain.NewThumbnailVariant(variant.String())
	if original == "" || !valid {
		return "", errors.New("invalid thumbnail cache identity")
	}
	key, valid := domain.NewStorageKey(original.String() + ".thumb/" + normalized.String())
	if !valid {
		return "", errors.New("invalid thumbnail cache key")
	}
	return key, nil
}

func thumbnailMetadataKey(key domain.StorageKey) domain.StorageKey {
	return domain.StorageKey(key.String() + ".meta")
}

func readCachedThumbnail(ctx context.Context, blobs ports.BlobStorage, key domain.StorageKey) (ports.ImageDerivative, error) {
	metadata, err := blobs.GetBlob(ctx, thumbnailMetadataKey(key))
	if err != nil {
		return ports.ImageDerivative{}, err
	}
	contentType, valid := domain.NewContentType(string(metadata))
	if !valid || !contentType.IsImage() {
		return ports.ImageDerivative{}, ports.ErrBlobNotFound
	}
	content, err := blobs.GetBlob(ctx, key)
	if err != nil {
		return ports.ImageDerivative{}, err
	}
	if len(content) == 0 {
		return ports.ImageDerivative{}, ports.ErrBlobNotFound
	}
	return ports.ImageDerivative{ContentType: contentType, Content: content}, nil
}
