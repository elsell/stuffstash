package blobstore

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestFileStreamPublishesOnlyCompleteContent(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	s := NewFileSystemStore(root)
	key := storageKey(t, "tenant/job/archive")
	input := ports.BlobStreamWrite{Key: key, ContentType: "application/zip", Content: bytes.NewReader([]byte("archive")), SizeBytes: 7, MaxBytes: 10}
	if err := s.PutBlobStream(ctx, input); err != nil {
		t.Fatal(err)
	}
	r, size, err := s.OpenBlobStream(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	data := make([]byte, 3)
	if n, err := r.ReadAt(data, 2); err != nil || n != 3 || string(data) != "chi" || size != 7 {
		t.Fatalf("random access: %q %v", data, err)
	}
	r.Close()
	for _, length := range []int64{6, 8, 11} {
		bad := input
		bad.Content = bytes.NewReader([]byte("archive"))
		bad.SizeBytes = length
		if err := s.PutBlobStream(ctx, bad); !errors.Is(err, ports.ErrBlobStreamSize) {
			t.Fatalf("bad size accepted: %v", err)
		}
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	input.Content = bytes.NewReader([]byte("changed"))
	if err := s.PutBlobStream(cancelled, input); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel ignored: %v", err)
	}
	final, _, err := s.OpenBlobStream(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	defer final.Close()
	body, err := io.ReadAll(final)
	if err != nil || string(body) != "archive" {
		t.Fatal("failed write replaced published object")
	}
}
func TestArchiveScratchIsPrivateAndUnlinkedWhileOpen(t *testing.T) {
	root := t.TempDir()
	scratch, err := (ScratchSpace{Directory: root}).NewArchiveScratch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = scratch.Write([]byte("temporary")); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatal("scratch has a directory entry that could survive a crash")
	}
	stat, err := scratch.(*scratchFile).Stat()
	if err != nil || stat.Mode().Perm() != 0600 {
		t.Fatal("scratch is not private")
	}
	if _, err = scratch.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	content, err := io.ReadAll(scratch)
	if err != nil || string(content) != "temporary" {
		t.Fatal("unlinked scratch is not readable")
	}
	if err = scratch.Close(); err != nil {
		t.Fatal(err)
	}
	entries, err = os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatal("scratch leaked")
	}
}

type failingArchiveReader struct {
	*bytes.Reader
	read bool
}

func (r *failingArchiveReader) Read(p []byte) (int, error) {
	if r.read {
		return 0, io.ErrUnexpectedEOF
	}
	r.read = true
	if len(p) > 3 {
		p = p[:3]
	}
	return r.Reader.Read(p)
}
func TestFileStreamFailureRemovesStagingAndPreservesExistingBlob(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	store := NewFileSystemStore(root)
	key := storageKey(t, "archive")
	original := ports.BlobStreamWrite{Key: key, ContentType: "application/zip", SizeBytes: 7, MaxBytes: 10, Content: bytes.NewReader([]byte("archive"))}
	if err := store.PutBlobStream(ctx, original); err != nil {
		t.Fatal(err)
	}
	failed := original
	failed.Content = &failingArchiveReader{Reader: bytes.NewReader([]byte("changed"))}
	if err := store.PutBlobStream(ctx, failed); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("copy failure not returned: %v", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 1 || entries[0].Name() != "archive" {
		t.Fatal("partial staging file leaked")
	}
	stream, _, err := store.OpenBlobStream(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	data, err := io.ReadAll(stream)
	if err != nil || string(data) != "archive" {
		t.Fatal("partial content published")
	}
}
