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

func TestReaderRechecksCacheAfterWaitingForWorker(t *testing.T) {
	processor, source, blobs, guard, claim := processorFixture(t)
	limiter, err := worklimit.New(1)
	if err != nil {
		t.Fatal(err)
	}
	release, err := limiter.Acquire(context.Background(), ports.ImageWorkBackground)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	waiting := make(chan struct{})
	reader, err := NewReader(processor, readerAdmissionSignal{ImageWorkAdmission: limiter, waiting: waiting})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		output, cached, err := reader.ReadThumbnail(ctx, source.attachment, domain.ThumbnailVariantSmall)
		if err == nil && (!cached || len(output.Content) == 0) {
			err = errors.New("reader did not reuse worker output")
		}
		done <- err
	}()
	select {
	case <-waiting:
	case <-ctx.Done():
		t.Fatal("reader never requested admission")
	}
	if err := processor.ProcessThumbnailJob(ctx, claim); err != nil {
		t.Fatal(err)
	}
	if err := blobs.DeleteBlob(ctx, source.attachment.StorageKey); err != nil {
		t.Fatal(err)
	}
	release()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if guard.writes != 3 {
		t.Fatal("reader regenerated worker output")
	}
}

type readerAdmissionSignal struct {
	ports.ImageWorkAdmission
	waiting chan struct{}
}

func (a readerAdmissionSignal) Acquire(ctx context.Context, priority ports.ImageWorkPriority) (func(), error) {
	close(a.waiting)
	return a.ImageWorkAdmission.Acquire(ctx, priority)
}

func TestReaderServesPublishedSmallBeforeBackgroundReleasesCapacity(t *testing.T) {
	processor, source, blobs, _, _ := processorFixture(t)
	limiter, err := worklimit.New(1)
	if err != nil {
		t.Fatal(err)
	}
	release, err := limiter.Acquire(context.Background(), ports.ImageWorkBackground)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	waiting := make(chan struct{})
	reader, err := NewReader(processor, readerAdmissionSignal{ImageWorkAdmission: limiter, waiting: waiting})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		value, cached, err := reader.ReadThumbnail(ctx, source.attachment, domain.ThumbnailVariantSmall)
		if err == nil && (!cached || string(value.Content) != "ready-small") {
			err = errors.New("published small was not served")
		}
		done <- err
	}()
	select {
	case <-waiting:
	case <-ctx.Done():
		t.Fatal("reader did not wait")
	}
	key, _ := ThumbnailCacheKey(source.attachment.StorageKey, domain.ThumbnailVariantSmall)
	if err := blobs.PutBlob(ctx, key, domain.ContentType("image/jpeg"), []byte("ready-small")); err != nil {
		t.Fatal(err)
	}
	if err := blobs.PutBlob(ctx, thumbnailMetadataKey(key), domain.ContentType("text/plain"), []byte("image/jpeg")); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal("small waited for remaining background variants", err)
	}
	blocked, stop := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer stop()
	if accidentalRelease, err := limiter.Acquire(blocked, ports.ImageWorkForeground); err == nil {
		accidentalRelease()
		t.Fatal("reader released background capacity")
	}
	release()
	available, stopAvailable := context.WithTimeout(context.Background(), time.Second)
	defer stopAvailable()
	permit, err := limiter.Acquire(available, ports.ImageWorkForeground)
	if err != nil {
		t.Fatal("reader leaked capacity", err)
	}
	permit()
}
