package gormstore

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
	store, attachment, record, job := thumbnailQueueFixture(t, ctx)
	if err := store.SaveAttachment(ctx, attachment, record, &job); err != nil {
		t.Fatal(err)
	}
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
	if err := store.db.Delete(&attachmentModel{ID: attachment.ID.String()}).Error; err != nil {
		t.Fatal(err)
	}
	if err := guard.Publish(ctx, attachment, nil, publish); !errors.Is(err, ports.ErrBlobNotFound) {
		t.Fatalf("deleted attachment publication: %v", err)
	}
	if writes != 1 {
		t.Fatal("deleted image was recreated")
	}
}

func TestThumbnailPublicationFencesBackgroundLease(t *testing.T) {
	ctx := context.Background()
	store, attachment, record, job := thumbnailQueueFixture(t, ctx)
	if err := store.SaveAttachment(ctx, attachment, record, &job); err != nil {
		t.Fatal(err)
	}
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
