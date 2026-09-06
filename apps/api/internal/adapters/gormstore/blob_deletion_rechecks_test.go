package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestBlobDeletionRechecksAreFairAndLeaseFenced(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	now := time.Date(2026, 9, 6, 22, 0, 0, 0, time.UTC)
	processed := now.Add(-2 * time.Hour)
	for _, id := range []string{"cleanup-a", "cleanup-b"} {
		if err := store.db.Create(&blobDeletionEventModel{ID: id, StorageKey: id, ProcessedAt: &processed}).Error; err != nil {
			t.Fatal(err)
		}
	}
	first, err := store.ClaimBlobDeletionRechecks(ctx, "claim", 1, now, now.Add(time.Minute), time.Hour)
	if err != nil || len(first) != 1 || first[0].ID != "cleanup-a" {
		t.Fatal("initial recheck failed", first, err)
	}
	if err := store.ResolveBlobDeletionRecheck(ctx, first[0], now, true); err != nil {
		t.Fatal(err)
	}
	second, err := store.ClaimBlobDeletionRechecks(ctx, "claim", 1, now, now.Add(time.Minute), time.Hour)
	if err != nil || len(second) != 1 || second[0].ID != "cleanup-b" {
		t.Fatal("failed recheck starved next tombstone", second, err)
	}
	later := now.Add(2 * time.Minute)
	replacement, err := store.ClaimBlobDeletionRechecks(ctx, "claim", 1, later, later.Add(time.Minute), time.Hour)
	if err != nil || len(replacement) != 1 {
		t.Fatal("expired recheck lease did not recover", err)
	}
	if err := store.ResolveBlobDeletionRecheck(ctx, second[0], later, false); !errors.Is(err, ports.ErrOutboxClaimLost) {
		t.Fatal("stale recheck resolved replacement", err)
	}
	if err := store.ResolveBlobDeletionRecheck(ctx, replacement[0], later, false); err != nil {
		t.Fatal(err)
	}
	immediate, err := store.ClaimBlobDeletionRechecks(ctx, "next", 1, later, later.Add(time.Minute), time.Hour)
	if err != nil || len(immediate) != 0 {
		t.Fatal("recheck ignored interval", err)
	}
	var model blobDeletionEventModel
	if err := store.db.First(&model, &blobDeletionEventModel{ID: "cleanup-a"}).Error; err != nil {
		t.Fatal(err)
	}
	if model.ProcessedAt == nil || !model.ProcessedAt.Equal(processed) {
		t.Fatal("recheck rewrote original processed status")
	}
}
