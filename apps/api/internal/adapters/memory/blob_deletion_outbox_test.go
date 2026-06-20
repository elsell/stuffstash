package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestBlobDeletionOutboxClaimFailureRetryAndProcessedLifecycle(t *testing.T) {
	store := NewStore()
	store.blobDeletions["event-one"] = ports.BlobDeletionEvent{
		ID:         "event-one",
		StorageKey: media.StorageKey("tenant/inventory/asset/attachment"),
	}

	ctx := context.Background()
	claimed, err := store.ClaimPendingBlobDeletionEvents(ctx, "claim-one", 1, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim blob deletion events: %v", err)
	}
	if len(claimed) != 1 || claimed[0].ID != "event-one" || claimed[0].ClaimID != "claim-one" {
		t.Fatalf("unexpected claimed events: %+v", claimed)
	}

	if err := store.MarkBlobDeletionEventFailed(ctx, "event-one", "claim-one", "storage unavailable"); err != nil {
		t.Fatalf("mark failed: %v", err)
	}
	reclaimed, err := store.ClaimPendingBlobDeletionEvents(ctx, "claim-two", 1, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("reclaim blob deletion events: %v", err)
	}
	if len(reclaimed) != 1 || reclaimed[0].Attempts != 1 || reclaimed[0].LastError != "storage unavailable" {
		t.Fatalf("expected retryable failed event, got %+v", reclaimed)
	}

	if err := store.MarkBlobDeletionEventProcessed(ctx, "event-one", "claim-two"); err != nil {
		t.Fatalf("mark processed: %v", err)
	}
	empty, err := store.ClaimPendingBlobDeletionEvents(ctx, "claim-three", 1, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim after processed: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected processed event not to be claimed, got %+v", empty)
	}
}
