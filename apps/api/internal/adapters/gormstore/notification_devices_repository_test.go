package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestNotificationDeviceScopeRevisionAndAtomicAudit(t *testing.T) {
	ctx := context.Background()
	store := newUndoableOperationTestStore(t, ctx)
	scope := ports.NotificationScope{TenantID: "tenant-one", InventoryID: "inventory-one", PrincipalID: "owner"}
	token, _ := notification.ParseDeviceToken("native-token")
	now := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	device := ports.NotificationDevice{ID: "device", Scope: scope, InstallationID: "installation", Transport: notification.PushAPNS, Token: token, Active: true, Revision: 1, CreatedAt: now, UpdatedAt: now}
	record := auditRecord(t, "device-audit", scope.TenantID, scope.InventoryID, audit.ActionNotificationPreferencesUpdated)
	record.InventoryID = audit.InventoryID(scope.InventoryID)
	record.PrincipalID = audit.PrincipalID(scope.PrincipalID)
	if err := store.SaveNotificationDevice(ctx, device, 0, record); err != nil {
		t.Fatal(err)
	}
	other := scope
	other.PrincipalID = "outsider"
	if _, found, err := store.NotificationDeviceByID(ctx, other, device.ID); err != nil || found {
		t.Fatal("cross-principal device read")
	}
	changed := device
	changed.Revision = 2
	changed.Active = false
	if err := store.SaveNotificationDevice(ctx, changed, 1, record); err == nil {
		t.Fatalf("audit conflict: %v", err)
	}
	actual, found, err := store.NotificationDeviceByID(ctx, scope, device.ID)
	if err != nil || !found || !actual.Active || actual.Revision != 1 {
		t.Fatal("audit failure changed device")
	}
	duplicate := device
	duplicate.ID = "other-device"
	duplicate.Scope = other
	record.ID = "other-audit"
	record.PrincipalID = "outsider"
	if err := store.SaveNotificationDevice(ctx, duplicate, 0, record); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("token ownership: %v", err)
	}
	record.ID = "revoke-audit"
	record.PrincipalID = "owner"
	if err := store.SaveNotificationDevice(ctx, changed, 1, record); err != nil {
		t.Fatal(err)
	}
	active, err := store.ListNotificationDevices(ctx, scope, "", 10)
	if err != nil || len(active) != 0 {
		t.Fatal("revoked device enumerated")
	}
}

func TestNotificationDeviceInstallationAndStaleRevision(t *testing.T) {
	ctx := context.Background()
	store := newUndoableOperationTestStore(t, ctx)
	scope := ports.NotificationScope{TenantID: "tenant-one", InventoryID: "inventory-one", PrincipalID: "owner"}
	token, _ := notification.ParseDeviceToken("native-token")
	now := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	device := ports.NotificationDevice{ID: "device", Scope: scope, InstallationID: "installation", Transport: notification.PushAPNS, Token: token, Active: true, Revision: 1, CreatedAt: now, UpdatedAt: now}
	record := auditRecord(t, "first", scope.TenantID, scope.InventoryID, audit.ActionNotificationPreferencesUpdated)
	record.InventoryID = audit.InventoryID(scope.InventoryID)
	record.PrincipalID = audit.PrincipalID(scope.PrincipalID)
	if err := store.SaveNotificationDevice(ctx, device, 0, record); err != nil {
		t.Fatal(err)
	}
	duplicate := device
	duplicate.ID = "duplicate"
	duplicate.Token, _ = notification.ParseDeviceToken("another-token")
	record.ID = "duplicate"
	if err := store.SaveNotificationDevice(ctx, duplicate, 0, record); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("duplicate installation: %v", err)
	}
	device.Revision = 2
	device.Token = duplicate.Token
	record.ID = "rotate"
	if err := store.SaveNotificationDevice(ctx, device, 1, record); err != nil {
		t.Fatal(err)
	}
	record.ID = "stale"
	if err := store.SaveNotificationDevice(ctx, device, 1, record); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("stale revision: %v", err)
	}
	current, found, err := store.NotificationDeviceByInstallation(ctx, scope, "installation")
	if err != nil || !found || current.Revision != 2 || current.Token != duplicate.Token {
		t.Fatal("rotation not retained")
	}
	wrong := scope
	wrong.TenantID = "another-home"
	if _, found, err := store.NotificationDeviceByInstallation(ctx, wrong, "installation"); err != nil || found {
		t.Fatal("cross-tenant installation")
	}
}
