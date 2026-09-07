package memory

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestBlobDeletionRecheckFailureDoesNotStarveAnotherRecord(t *testing.T) {
	store := NewStore()
	ctx := context.Background()
	now := time.Date(2026, 9, 6, 22, 0, 0, 0, time.UTC)
	for _, id := range []string{"a", "b"} {
		store.blobDeletions[id] = ports.BlobDeletionEvent{ID: id, StorageKey: media.StorageKey(id), ProcessedAt: now.Add(-2 * time.Hour)}
	}
	first, err := store.ClaimBlobDeletionRechecks(ctx, "claim", 1, now, now.Add(time.Minute), time.Hour)
	if err != nil || len(first) != 1 || first[0].ID != "a" {
		t.Fatal("initial claim failed", err)
	}
	for _, key := range []media.StorageKey{"", "wrong"} {
		changed := first[0]
		changed.StorageKey = key
		if err := store.ResolveBlobDeletionRecheck(ctx, changed, now, false); !errors.Is(err, ports.ErrOutboxClaimLost) {
			t.Fatal("changed cleanup identity resolved", err)
		}
	}
	if err := store.ResolveBlobDeletionRecheck(ctx, first[0], now, true); err != nil {
		t.Fatal(err)
	}
	next, err := store.ClaimBlobDeletionRechecks(ctx, "claim", 1, now, now.Add(time.Minute), time.Hour)
	if err != nil || len(next) != 1 || next[0].ID != "b" {
		t.Fatal("failed record starved another", err)
	}
	later := now.Add(2 * time.Minute)
	replacement, err := store.ClaimBlobDeletionRechecks(ctx, "claim", 1, later, later.Add(time.Minute), time.Hour)
	if err != nil || len(replacement) != 1 {
		t.Fatal("lease did not recover", err)
	}
	if err := store.ResolveBlobDeletionRecheck(ctx, next[0], later, false); !errors.Is(err, ports.ErrOutboxClaimLost) {
		t.Fatal("stale recheck resolved", err)
	}
	if err := store.ResolveBlobDeletionRecheck(ctx, replacement[0], later, false); err != nil {
		t.Fatal(err)
	}
}
