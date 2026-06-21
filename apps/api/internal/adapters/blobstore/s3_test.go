package blobstore

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestNewS3StoreRequiresConnectionSettings(t *testing.T) {
	if _, err := NewS3Store(S3Config{}); err == nil {
		t.Fatalf("expected missing S3 settings to fail")
	}
}

func TestNewS3StoreTrimsConnectionSettings(t *testing.T) {
	store, err := NewS3Store(S3Config{
		Endpoint:  " 127.0.0.1:3900 ",
		AccessKey: " access ",
		SecretKey: " secret ",
		Bucket:    " bucket ",
		Region:    " garage ",
		MaxBytes:  8,
	})
	if err != nil {
		t.Fatalf("create S3 store: %v", err)
	}
	if store.bucket != "bucket" {
		t.Fatalf("expected trimmed bucket, got %q", store.bucket)
	}
	if store.maxBytes != 8 {
		t.Fatalf("expected max bytes to be retained, got %d", store.maxBytes)
	}
}

func TestMapS3MissingObjectErrors(t *testing.T) {
	err := mapS3Error(minio.ErrorResponse{Code: "NoSuchKey"})
	if !errors.Is(err, ports.ErrBlobNotFound) {
		t.Fatalf("expected missing key to map to blob not found, got %v", err)
	}
}

func TestS3StoreRejectsOversizedBlobReads(t *testing.T) {
	if _, err := readBlobBytes(bytes.NewReader([]byte("012345678")), 8); err == nil {
		t.Fatalf("expected oversized blob read to fail")
	}
}

func TestS3StoreRejectsOversizedBlobWritesBeforeNetwork(t *testing.T) {
	store := S3Store{maxBytes: 4}
	key, ok := media.NewStorageKey("tenant/inventory/asset/attachment")
	if !ok {
		t.Fatal("invalid storage key")
	}

	err := store.PutBlob(context.Background(), key, media.ContentTypePNG, []byte("12345"))
	if err == nil {
		t.Fatalf("expected oversized blob write to fail")
	}
}

func TestS3StoreAllowsBlobReadsWithinLimit(t *testing.T) {
	data, err := readBlobBytes(bytes.NewReader([]byte("01234567")), 8)
	if err != nil {
		t.Fatalf("read blob bytes: %v", err)
	}
	if !bytes.Equal(data, []byte("01234567")) {
		t.Fatalf("expected original bytes, got %q", string(data))
	}
}
