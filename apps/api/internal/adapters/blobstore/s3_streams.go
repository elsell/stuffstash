package blobstore

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

var _ ports.StreamingBlobStorage = S3Store{}

func (s S3Store) OpenBlobStream(ctx context.Context, key media.StorageKey) (ports.BlobReadStream, int64, error) {
	stat, err := s.client.StatObject(ctx, s.bucket, key.String(), minio.StatObjectOptions{})
	if err != nil {
		return nil, 0, mapS3Error(err)
	}
	options := minio.GetObjectOptions{}
	if err = options.SetMatchETag(stat.ETag); err != nil {
		return nil, 0, err
	}
	object, err := s.client.GetObject(ctx, s.bucket, key.String(), options)
	if err != nil {
		return nil, 0, mapS3Error(err)
	}
	return object, stat.Size, nil
}
func (s S3Store) PutBlobStream(ctx context.Context, input ports.BlobStreamWrite) error {
	if err := validateStream(ctx, input); err != nil {
		return err
	}
	// A known-size single PUT avoids multipart leftovers after cancellation.
	_, err := s.client.PutObject(ctx, s.bucket, input.Key.String(), streamContextReader{ctx: ctx, r: input.Content}, input.SizeBytes, minio.PutObjectOptions{ContentType: input.ContentType, DisableMultipart: true})
	return mapS3Error(err)
}
