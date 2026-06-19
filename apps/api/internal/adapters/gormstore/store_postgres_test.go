package gormstore

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func TestPostgresStoreClaimsOutboxEventOnceAcrossWorkers(t *testing.T) {
	dsn := os.Getenv("STUFF_STASH_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set STUFF_STASH_TEST_POSTGRES_DSN to run Postgres outbox concurrency verification")
	}

	ctx := context.Background()
	db, err := OpenPostgres(dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("postgres db handle: %v", err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Fatalf("close postgres: %v", err)
		}
	})
	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("migrate postgres: %v", err)
	}

	store := NewStore(db)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	eventID := "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	cleanupAuthorizationOutboxTestRows(t, ctx, store, eventID, tenantID)
	saveTenantWithOutbox(t, ctx, store, eventID, tenantID, "Concurrency Home")

	claims := make(chan string, 2)
	var wg sync.WaitGroup
	for _, claimID := range []string{"claim-one", "claim-two"} {
		wg.Add(1)
		go func(claimID string) {
			defer wg.Done()
			events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, claimID, 1, time.Now().Add(time.Minute))
			if err != nil {
				t.Errorf("claim %s: %v", claimID, err)
				return
			}
			for _, event := range events {
				claims <- event.ClaimID
			}
		}(claimID)
	}
	wg.Wait()
	close(claims)

	claimedBy := []string{}
	for claimID := range claims {
		claimedBy = append(claimedBy, claimID)
	}
	if len(claimedBy) != 1 {
		t.Fatalf("expected exactly one worker to claim event, got %+v", claimedBy)
	}

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-three", 1, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim while lease active: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected active lease to hide event from third worker, got %+v", events)
	}
}

func cleanupAuthorizationOutboxTestRows(t *testing.T, ctx context.Context, store Store, eventID string, tenantID tenant.ID) {
	t.Helper()

	if err := store.db.WithContext(ctx).Delete(&authorizationOutboxEventModel{ID: eventID}).Error; err != nil {
		t.Fatalf("clean outbox row: %v", err)
	}
	if err := store.db.WithContext(ctx).Delete(&tenantModel{ID: tenantID.String()}).Error; err != nil {
		t.Fatalf("clean tenant row: %v", err)
	}
}
