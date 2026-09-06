package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type publicationClock struct{ now time.Time }

func (c *publicationClock) Now() time.Time { return c.now }

func TestThumbnailPublicationChecksAuthoritativeIdentity(t *testing.T) {
	ctx := context.Background()
	store, attachment, job := memoryPublicationFixture(t)
	clock := &publicationClock{now: job.CreatedAt}
	guard, err := NewThumbnailPublicationGuard(store, clock)
	if err != nil {
		t.Fatal(err)
	}
	writes := 0
	publish := func(context.Context) error { writes++; return nil }
	if err := guard.Publish(ctx, attachment, nil, publish); err != nil || writes != 1 {
		t.Fatalf("legitimate publication failed: %v", err)
	}
	for _, change := range []func(*media.Attachment){
		func(a *media.Attachment) { a.TenantID = "other" },
		func(a *media.Attachment) { a.InventoryID = "other" },
		func(a *media.Attachment) { a.AssetID = "other" },
		func(a *media.Attachment) { a.TenantID = "" },
		func(a *media.Attachment) { a.SHA256 = "changed" },
		func(a *media.Attachment) { a.StorageKey = "changed" },
	} {
		changed := attachment
		change(&changed)
		if err := guard.Publish(ctx, changed, nil, publish); err == nil {
			t.Fatal("invalid identity was published")
		}
	}
	if writes != 1 {
		t.Fatal("invalid identity reached publisher")
	}
	store.mu.Lock()
	delete(store.attachments, attachment.ID)
	store.mu.Unlock()
	if err := guard.Publish(ctx, attachment, nil, publish); !errors.Is(err, ports.ErrBlobNotFound) {
		t.Fatalf("deleted attachment publication: %v", err)
	}
	if writes != 1 {
		t.Fatal("deleted image was recreated")
	}
}

func TestThumbnailPublicationFencesBackgroundLease(t *testing.T) {
	ctx := context.Background()
	store, attachment, job := memoryPublicationFixture(t)
	clock := &publicationClock{now: job.CreatedAt}
	guard, err := NewThumbnailPublicationGuard(store, clock)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := store.ClaimThumbnailJobs(ctx, "worker", 1, clock.Now(), clock.Now().Add(time.Minute))
	if err != nil || len(claims) != 1 {
		t.Fatal("claim failed")
	}
	writes := 0
	publish := func(context.Context) error { writes++; return nil }
	if err := guard.Publish(ctx, attachment, &claims[0], publish); err != nil {
		t.Fatal(err)
	}
	clock.now = clock.now.Add(2 * time.Minute)
	if err := guard.Publish(ctx, attachment, &claims[0], publish); !errors.Is(err, ports.ErrOutboxClaimLost) {
		t.Fatalf("expired publisher admitted: %v", err)
	}
	replacement, err := store.ClaimThumbnailJobs(ctx, "worker", 1, clock.Now(), clock.Now().Add(time.Minute))
	if err != nil || len(replacement) != 1 {
		t.Fatal("reclaim failed")
	}
	if err := guard.Publish(ctx, attachment, &claims[0], publish); !errors.Is(err, ports.ErrOutboxClaimLost) {
		t.Fatalf("replaced publisher admitted: %v", err)
	}
	if writes != 1 {
		t.Fatal("stale worker reached blob publication")
	}
	sentinel := errors.New("controlled publication failure")
	if err := guard.Publish(ctx, attachment, &replacement[0], func(context.Context) error { return sentinel }); !errors.Is(err, sentinel) {
		t.Fatal("publication failure hidden")
	}
}

func memoryPublicationFixture(t *testing.T) (*Store, media.Attachment, media.ThumbnailJob) {
	t.Helper()
	store := NewStore()
	now := time.Date(2026, 9, 6, 22, 0, 0, 0, time.UTC)
	attachment := media.Attachment{ID: "photo", TenantID: "tenant", InventoryID: "inventory", AssetID: "asset", StorageKey: "original", SHA256: "hash", ContentType: media.ContentTypeJPEG, CreatedAt: now}
	job, err := media.NewThumbnailJob(attachment, media.ThumbnailJobNewImage, now)
	if err != nil {
		t.Fatal(err)
	}
	store.attachments[attachment.ID] = attachment
	store.enqueueThumbnailJob(&job)
	return store, attachment, job
}

func TestThumbnailPublicationCanWriteBlobsWhileHoldingLifecycleLock(t *testing.T) {
	store, attachment, job := memoryPublicationFixture(t)
	guard, err := NewThumbnailPublicationGuard(store, &publicationClock{now: job.CreatedAt})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		done <- guard.Publish(context.Background(), attachment, nil, func(ctx context.Context) error {
			if store.mu.TryLock() {
				store.mu.Unlock()
				return errors.New("lifecycle lock not held")
			}
			return store.PutBlob(ctx, "derivative", media.ContentTypeJPEG, []byte("data"))
		})
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("publication deadlocked with blob storage")
	}
	if _, err := store.GetBlob(context.Background(), "derivative"); err != nil {
		t.Fatal(err)
	}
}
