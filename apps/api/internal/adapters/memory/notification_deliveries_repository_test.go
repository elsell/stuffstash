package memory

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestNotificationDeliveryAtomicPublicationAndFencing(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	scope := ports.NotificationScope{TenantID: "home", InventoryID: "main", PrincipalID: "owner"}
	saveMemoryTenant(t, ctx, store, scope.TenantID)
	saveMemoryInventory(t, ctx, store, scope.TenantID, scope.InventoryID)
	now := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	token, _ := notification.ParseDeviceToken("abcd")
	record := memoryAuditRecord(t, "device-audit", scope.TenantID)
	record.InventoryID = audit.InventoryID(scope.InventoryID)
	record.PrincipalID = audit.PrincipalID(scope.PrincipalID)
	device := ports.NotificationDevice{ID: "phone", Scope: scope, InstallationID: "phone", Transport: notification.PushAPNS, Token: token, Active: true, Revision: 1, CreatedAt: now, UpdatedAt: now}
	if err := store.SaveNotificationDevice(ctx, device, 0, record); err != nil {
		t.Fatal(err)
	}
	date, _ := expirationdate.ParseDate("2026-09-20", expirationdate.Day)
	notice := ports.NotificationRecord{ID: "notice", Scope: scope, CreatedAt: now, Milestone: notification.Milestone{AssetID: "item", Date: date, Kind: notification.MilestoneUpcoming}}
	state, _ := notification.NewDelivery(now)
	delivery := ports.NotificationDelivery{ID: "delivery", Scope: scope, NotificationID: "notice", DeviceID: "phone", DeviceRevision: 1, CreatedAt: now, State: state}
	record.ID = "notice-audit"
	invalid := delivery
	invalid.DeviceRevision = 2
	if _, _, err := store.InsertNotificationWithDeliveries(ctx, notice, []ports.NotificationDelivery{invalid}, record); err == nil {
		t.Fatal("stale destination accepted")
	}
	if _, found, _ := store.NotificationByID(ctx, scope, "notice"); found {
		t.Fatal("failed publication wrote inbox")
	}
	wrongScope := delivery
	wrongScope.Scope.PrincipalID = "another-user"
	for _, batch := range [][]ports.NotificationDelivery{{wrongScope}, {delivery, delivery}} {
		if _, _, err := store.InsertNotificationWithDeliveries(ctx, notice, batch, record); err == nil {
			t.Fatal("invalid destination batch accepted")
		}
	}
	conflictingAudit := record
	conflictingAudit.ID = "device-audit"
	if _, _, err := store.InsertNotificationWithDeliveries(ctx, notice, []ports.NotificationDelivery{delivery}, conflictingAudit); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("audit collision: %v", err)
	}
	if _, found, _ := store.NotificationByID(ctx, scope, notice.ID); found {
		t.Fatal("audit collision wrote inbox")
	}
	if _, created, err := store.InsertNotificationWithDeliveries(ctx, notice, []ports.NotificationDelivery{delivery}, record); err != nil || !created {
		t.Fatalf("publish: %v", err)
	}
	if _, created, err := store.InsertNotificationWithDeliveries(ctx, notice, []ports.NotificationDelivery{delivery}, record); err != nil || created {
		t.Fatalf("duplicate: %v", err)
	}
	policy := notification.RetryPolicy{MaxAttempts: 2, InitialDelay: time.Second, MaximumDelay: time.Minute}
	first, err := store.ClaimNotificationDeliveries(ctx, now, "first", time.Minute, policy, 10)
	if err != nil || len(first) != 1 {
		t.Fatalf("claim: %v %v", first, err)
	}
	second, err := store.ClaimNotificationDeliveries(ctx, now.Add(time.Minute), "second", time.Minute, policy, 10)
	if err != nil || len(second) != 1 {
		t.Fatalf("reclaim: %v", err)
	}
	if err := store.SettleNotificationDelivery(ctx, "delivery", "first", now.Add(time.Minute), ports.NotificationDeliveryAccepted, policy, time.Time{}); !errors.Is(err, notification.ErrStaleDeliveryLease) {
		t.Fatalf("stale fence: %v", err)
	}
	if err := store.SettleNotificationDelivery(ctx, "delivery", "second", now.Add(time.Minute), ports.NotificationDeliveryAccepted, policy, time.Time{}); err != nil {
		t.Fatal(err)
	}
	pending, err := store.ClaimNotificationDeliveries(ctx, now.Add(time.Hour), "third", time.Minute, policy, 10)
	if err != nil || len(pending) != 0 {
		t.Fatal("accepted delivery redelivered")
	}
	notice.ID = "retry-notice"
	notice.Milestone.AssetID = "another-item"
	delivery.ID = "retry-delivery"
	delivery.NotificationID = notice.ID
	record.ID = "retry-audit"
	if _, _, err := store.InsertNotificationWithDeliveries(ctx, notice, []ports.NotificationDelivery{delivery}, record); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ClaimNotificationDeliveries(ctx, now, "retry-one", time.Minute, policy, 1); err != nil {
		t.Fatal(err)
	}
	if err := store.SettleNotificationDelivery(ctx, delivery.ID, "retry-one", now, ports.NotificationDeliveryRetry, policy, now.Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	early, err := store.ClaimNotificationDeliveries(ctx, now.Add(time.Hour), "early", time.Minute, policy, 1)
	if err != nil || len(early) != 0 {
		t.Fatal("retry ignored backoff")
	}
	due, err := store.ClaimNotificationDeliveries(ctx, now.Add(2*time.Hour), "retry-two", time.Minute, policy, 1)
	if err != nil || len(due) != 1 {
		t.Fatal("retry not due")
	}
	if err := store.SettleNotificationDelivery(ctx, delivery.ID, "retry-two", now.Add(2*time.Hour), ports.NotificationDeliveryRetry, policy, time.Time{}); err != nil {
		t.Fatal(err)
	}
	exhausted, err := store.ClaimNotificationDeliveries(ctx, now.Add(3*time.Hour), "retry-three", time.Minute, policy, 1)
	if err != nil || len(exhausted) != 0 {
		t.Fatal("exhausted retry reclaimed")
	}

}
