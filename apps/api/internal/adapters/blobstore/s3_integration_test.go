package blobstore

import (
	"bytes"
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestS3StoreAgainstGarage(t *testing.T) {
	endpoint := os.Getenv("STUFF_STASH_TEST_S3_ENDPOINT")
	accessKey := os.Getenv("STUFF_STASH_TEST_S3_ACCESS_KEY")
	secretKey := os.Getenv("STUFF_STASH_TEST_S3_SECRET_KEY")
	bucket := os.Getenv("STUFF_STASH_TEST_S3_BUCKET")
	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		t.Skip("Garage S3 test environment is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	store, err := NewS3Store(S3Config{
		Endpoint:  endpoint,
		AccessKey: accessKey,
		SecretKey: secretKey,
		Bucket:    bucket,
		Region:    "garage",
		Secure:    false,
		MaxBytes:  1024,
	})
	if err != nil {
		t.Fatalf("create S3 store: %v", err)
	}
	key := storageKey(t, "tenant/inventory/asset/garage-test")
	content := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}

	if err := store.PutBlob(ctx, key, media.ContentTypePNG, content); err != nil {
		t.Fatalf("put blob: %v", err)
	}
	read, err := store.GetBlob(ctx, key)
	if err != nil {
		t.Fatalf("get blob: %v", err)
	}
	if !bytes.Equal(read, content) {
		t.Fatalf("expected %q, got %q", string(content), string(read))
	}
	if err := store.DeleteBlob(ctx, key); err != nil {
		t.Fatalf("delete blob: %v", err)
	}
	if _, err := store.GetBlob(ctx, key); !errors.Is(err, ports.ErrBlobNotFound) {
		t.Fatalf("expected deleted blob to be missing, got %v", err)
	}
}
