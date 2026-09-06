package media

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/worklimit"
	domain "github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestReaderUsesForegroundAdmissionAndServesCacheWithoutCapacity(t *testing.T) {
	processor, source, _, guard, _ := processorFixture(t)
	limiter, err := worklimit.New(1)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := NewReader(processor, limiter)
	if err != nil {
		t.Fatal(err)
	}
	release, err := limiter.Acquire(context.Background(), ports.ImageWorkBackground)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, _, err := reader.ReadThumbnail(ctx, source.attachment, domain.ThumbnailVariantSmall); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("cold read bypassed shared admission", err)
	}
	if guard.writes != 0 {
		t.Fatal("blocked reader published")
	}
	release()
	first, cached, err := reader.ReadThumbnail(context.Background(), source.attachment, domain.ThumbnailVariantSmall)
	if err != nil || cached || len(first.Content) == 0 {
		t.Fatal("cold read failed", err)
	}
	release, err = limiter.Acquire(context.Background(), ports.ImageWorkBackground)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
	defer cancel2()
	second, cached, err := reader.ReadThumbnail(ctx2, source.attachment, domain.ThumbnailVariantSmall)
	if err != nil || !cached || len(second.Content) == 0 {
		t.Fatal("cache hit waited for admission", err)
	}
	if guard.writes != 1 {
		t.Fatal("cache hit generated again")
	}
}

func TestReaderGuardRejectsAttachmentDeletedAfterAuthorization(t *testing.T) {
	processor, source, blobs, guard, _ := processorFixture(t)
	limiter, err := worklimit.New(1)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := NewReader(processor, limiter)
	if err != nil {
		t.Fatal(err)
	}
	guard.before = func() { source.missing = true }
	if _, _, err := reader.ReadThumbnail(context.Background(), source.attachment, domain.ThumbnailVariantSmall); err == nil {
		t.Fatal("deleted attachment published")
	}
	key, _ := ThumbnailCacheKey(source.attachment.StorageKey, domain.ThumbnailVariantSmall)
	if _, err := blobs.GetBlob(context.Background(), key); !errors.Is(err, ports.ErrBlobNotFound) {
		t.Fatal("deleted derivative recreated")
	}
}
