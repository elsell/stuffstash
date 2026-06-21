package blobstore

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type S3Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	Secure    bool
	MaxBytes  int64
}

type S3Store struct {
	client                 *minio.Client
	bucket                 string
	maxBytes               int64
	directUploadSigningKey []byte
}

func NewS3Store(config S3Config) (S3Store, error) {
	endpoint := strings.TrimSpace(config.Endpoint)
	accessKey := strings.TrimSpace(config.AccessKey)
	secretKey := strings.TrimSpace(config.SecretKey)
	bucket := strings.TrimSpace(config.Bucket)
	region := strings.TrimSpace(config.Region)
	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		return S3Store{}, errors.New("s3 endpoint, access key, secret key, and bucket are required")
	}
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: config.Secure,
		Region: region,
	})
	if err != nil {
		return S3Store{}, err
	}
	return S3Store{client: client, bucket: bucket, maxBytes: config.MaxBytes, directUploadSigningKey: []byte(secretKey)}, nil
}

func (s S3Store) PutBlob(ctx context.Context, key media.StorageKey, contentType media.ContentType, data []byte) error {
	if s.maxBytes > 0 && int64(len(data)) > s.maxBytes {
		return errors.New("blob exceeds configured maximum size")
	}
	_, err := s.client.PutObject(ctx, s.bucket, key.String(), bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType.String(),
	})
	return err
}

func (s S3Store) GetBlob(ctx context.Context, key media.StorageKey) ([]byte, error) {
	object, err := s.client.GetObject(ctx, s.bucket, key.String(), minio.GetObjectOptions{})
	if err != nil {
		return nil, mapS3Error(err)
	}
	defer object.Close()

	data, err := readBlobBytes(object, s.maxBytes)
	if err != nil {
		return nil, mapS3Error(err)
	}
	return data, nil
}

func (s S3Store) DeleteBlob(ctx context.Context, key media.StorageKey) error {
	if err := s.client.RemoveObject(ctx, s.bucket, key.String(), minio.RemoveObjectOptions{}); err != nil {
		return mapS3Error(err)
	}
	return nil
}

func mapS3Error(err error) error {
	if err == nil {
		return nil
	}
	response := minio.ToErrorResponse(err)
	switch response.Code {
	case "NoSuchKey", "NoSuchBucket", "NotFound":
		return ports.ErrBlobNotFound
	default:
		return err
	}
}

func readBlobBytes(reader io.Reader, maxBytes int64) ([]byte, error) {
	bounded := reader
	if maxBytes > 0 {
		bounded = io.LimitReader(reader, maxBytes+1)
	}
	data, err := io.ReadAll(bounded)
	if err != nil {
		return nil, err
	}
	if maxBytes > 0 && int64(len(data)) > maxBytes {
		return nil, errors.New("blob exceeds configured maximum size")
	}
	return data, nil
}
