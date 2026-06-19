package blobstore

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestFileSystemStoreWritesReadsAndDeletesBlob(t *testing.T) {
	store := NewFileSystemStore(t.TempDir())
	key := storageKey(t, "tenant/inventory/asset/attachment")
	content := []byte("blob content")

	if err := store.PutBlob(context.Background(), key, media.ContentTypePNG, content); err != nil {
		t.Fatalf("put blob: %v", err)
	}
	content[0] = 'X'

	read, err := store.GetBlob(context.Background(), key)
	if err != nil {
		t.Fatalf("get blob: %v", err)
	}
	if !bytes.Equal(read, []byte("blob content")) {
		t.Fatalf("expected stored bytes to be immutable, got %q", string(read))
	}

	if err := store.DeleteBlob(context.Background(), key); err != nil {
		t.Fatalf("delete blob: %v", err)
	}
	if _, err := store.GetBlob(context.Background(), key); !errors.Is(err, ports.ErrBlobNotFound) {
		t.Fatalf("expected blob not found after delete, got %v", err)
	}
}

func TestFileSystemStoreCreatesNestedDirectories(t *testing.T) {
	root := t.TempDir()
	store := NewFileSystemStore(root)
	key := storageKey(t, "a/b/c/d")

	if err := store.PutBlob(context.Background(), key, media.ContentTypePDF, []byte("data")); err != nil {
		t.Fatalf("put blob: %v", err)
	}

	if _, err := store.GetBlob(context.Background(), key); err != nil {
		t.Fatalf("get blob: %v", err)
	}
	if _, err := filepath.Abs(root); err != nil {
		t.Fatalf("temp root should be absolute-able: %v", err)
	}
}

func TestFileSystemStoreRejectsEmptyRoot(t *testing.T) {
	store := NewFileSystemStore(" ")

	if err := store.PutBlob(context.Background(), storageKey(t, "key"), media.ContentTypePNG, []byte("data")); err == nil {
		t.Fatalf("expected empty root error")
	}
}

func TestFileSystemStoreRejectsEscapingKey(t *testing.T) {
	store := NewFileSystemStore(t.TempDir())
	key := media.StorageKey("../escape")

	if err := store.PutBlob(context.Background(), key, media.ContentTypePNG, []byte("data")); err == nil {
		t.Fatalf("expected escaping key error")
	}
}

func storageKey(t *testing.T, value string) media.StorageKey {
	t.Helper()

	key, ok := media.NewStorageKey(value)
	if !ok {
		t.Fatalf("invalid storage key %q", value)
	}
	return key
}
