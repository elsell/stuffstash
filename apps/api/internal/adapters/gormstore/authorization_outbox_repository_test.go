package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestStoreMarksOutboxEventsProcessedAndFailed(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	tenantName, ok := tenant.NewName("Home")
	if !ok {
		t.Fatalf("expected valid tenant name")
	}
	if err := store.SaveTenantAndEnqueueOwnerGrant(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAW", tenant.Tenant{
		ID:   tenantID,
		Name: tenantName,
	}, identity.Principal{ID: identity.PrincipalID("user-one")}, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID, "", audit.ActionTenantCreated)); err != nil {
		t.Fatalf("save tenant and enqueue owner grant: %v", err)
	}

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-one", 10, time.Now(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 claimed event, got %+v", events)
	}

	if err := store.MarkAuthorizationOutboxEventFailed(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAW", "claim-one", "spicedb unavailable"); err != nil {
		t.Fatalf("mark outbox failed: %v", err)
	}
	events, err = store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-two", 10, time.Now(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 1 || events[0].Attempts != 1 || events[0].LastError != "spicedb unavailable" {
		t.Fatalf("expected failed event to remain pending, got %+v", events)
	}

	if err := store.MarkAuthorizationOutboxEventProcessed(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAW", "wrong-claim"); !errors.Is(err, ports.ErrOutboxClaimLost) {
		t.Fatalf("expected claim lost error, got %v", err)
	}
	events, err = store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-three", 10, time.Now(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected active claim to hide event from wrong processor, got %+v", events)
	}

	if err := store.MarkAuthorizationOutboxEventProcessed(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAW", "claim-two"); err != nil {
		t.Fatalf("mark outbox processed: %v", err)
	}
	events, err = store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-three", 10, time.Now(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected processed event hidden from pending list, got %+v", events)
	}
}

func TestStoreMarksOutboxEventsDeadLettered(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	saveTenantWithOutbox(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FAW", tenantID, "Home")

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-one", 10, time.Now(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 claimed event, got %+v", events)
	}

	if err := store.MarkAuthorizationOutboxEventDeadLettered(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAW", "wrong-claim", "invalid event"); !errors.Is(err, ports.ErrOutboxClaimLost) {
		t.Fatalf("expected claim lost error, got %v", err)
	}
	if err := store.MarkAuthorizationOutboxEventDeadLettered(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAW", "claim-one", "invalid event"); err != nil {
		t.Fatalf("mark outbox dead-lettered: %v", err)
	}

	var model authorizationOutboxEventModel
	if err := store.db.WithContext(ctx).Where(&authorizationOutboxEventModel{ID: "01ARZ3NDEKTSV4RRFFQ69G5FAW"}).First(&model).Error; err != nil {
		t.Fatalf("load outbox event: %v", err)
	}
	if model.DeadLetteredAt == nil || model.DeadLetterReason != "invalid event" {
		t.Fatalf("expected dead-letter details, got %+v", model)
	}

	events, err = store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-two", 10, time.Now(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected dead-lettered event hidden from pending list, got %+v", events)
	}
}

func TestStoreClaimsHideEventsUntilLeaseExpires(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	saveTenantWithOutbox(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FAW", tenantID, "Home")

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-one", 10, time.Now(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 1 || events[0].ClaimID != "claim-one" {
		t.Fatalf("expected claim-one to own event, got %+v", events)
	}

	events, err = store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-two", 10, time.Now(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected active lease to hide event, got %+v", events)
	}
}

func TestStoreReclaimsEventsAfterLeaseExpires(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	saveTenantWithOutbox(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FAW", tenantID, "Home")

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-one", 10, time.Now(), time.Now().Add(-time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected claim-one to claim event, got %+v", events)
	}

	events, err = store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-two", 10, time.Now(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 1 || events[0].ClaimID != "claim-two" {
		t.Fatalf("expected expired lease to be reclaimed, got %+v", events)
	}
}
