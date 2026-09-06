package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestThumbnailReclaimFencesSameTokenAndCountsCrash(t *testing.T) {
	store := NewStore()
	now := time.Date(2026, 9, 6, 22, 0, 0, 123456789, time.UTC)
	attachment := media.Attachment{ID: "photo", TenantID: "tenant", InventoryID: "inventory", AssetID: "asset", StorageKey: "original", SHA256: "hash", ContentType: media.ContentTypeJPEG, CreatedAt: now}
	job, err := media.NewThumbnailJob(attachment, media.ThumbnailJobNewImage, now)
	if err != nil {
		t.Fatal(err)
	}
	store.attachments[attachment.ID] = attachment
	store.enqueueThumbnailJob(&job)
	ctx := context.Background()
	first, err := store.ClaimThumbnailJobs(ctx, "worker", 1, now.Add(time.Second), now.Add(time.Minute))
	if err != nil || len(first) != 1 {
		t.Fatal("initial claim failed")
	}
	later := now.Add(2 * time.Minute)
	replacement, err := store.ClaimThumbnailJobs(ctx, "worker", 1, later, later.Add(time.Minute))
	if err != nil || len(replacement) != 1 {
		t.Fatal("reclaim failed")
	}
	if first[0].Attempts != 1 || replacement[0].Attempts != 2 {
		t.Fatal("crash did not consume an attempt")
	}
	done := ports.ThumbnailJobResolution{Status: ports.ThumbnailJobCompleted, At: later}
	if err := store.ResolveThumbnailJob(ctx, first[0], done); !errors.Is(err, ports.ErrOutboxClaimLost) {
		t.Fatalf("expired claim resolved replacement: %v", err)
	}
	if err := store.ResolveThumbnailJob(ctx, replacement[0], done); err != nil {
		t.Fatal(err)
	}
}
