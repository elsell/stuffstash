package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestNotificationPreferencesPersistenceIsolationAndRevision(t *testing.T) {
	ctx := context.Background()
	store := newUndoableOperationTestStore(t, ctx)
	scope := ports.NotificationScope{TenantID: tenant.ID("tenant-one"), InventoryID: inventory.InventoryID("inventory-one"), PrincipalID: identity.PrincipalID("owner")}
	value := ports.NotificationPreferencesRecord{ID: "preferences-one", Scope: scope, Revision: 1, Settings: notification.DefaultSettings("America/New_York"), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	record := auditRecord(t, "preferences-audit", scope.TenantID, scope.InventoryID, audit.ActionNotificationPreferencesUpdated)
	record.PrincipalID = audit.PrincipalID(scope.PrincipalID.String())
	if err := store.SaveNotificationPreferences(ctx, value, 0, record); err != nil {
		t.Fatal(err)
	}
	read, found, err := store.NotificationPreferences(ctx, scope)
	if err != nil || !found || !read.Settings.Equal(value.Settings) {
		t.Fatalf("roundtrip %+v %v", read, err)
	}
	other := scope
	other.PrincipalID = "other"
	if _, found, err := store.NotificationPreferences(ctx, other); err != nil || found {
		t.Fatal("cross-principal settings leak")
	}
	other = scope
	other.TenantID = "other-tenant"
	if _, found, err := store.NotificationPreferences(ctx, other); err != nil || found {
		t.Fatal("cross-tenant settings leak")
	}
	changed := read.Clone()
	changed.Revision = 2
	changed.Settings.Defaults.Enabled = false
	changed.Settings.Overrides["medicine"] = notification.ExpirationPreferences{Enabled: true, Upcoming: true, Expired: true, AdvanceDays: 60}
	record.ID = "preferences-update"
	if err := store.SaveNotificationPreferences(ctx, changed, 1, record); err != nil {
		t.Fatal(err)
	}
	failed := changed.Clone()
	failed.Revision = 3
	failed.Settings.Defaults.AdvanceDays = 99
	if err := store.SaveNotificationPreferences(ctx, failed, 2, record); err == nil {
		t.Fatal("duplicate audit accepted")
	}
	unchanged, found, err := store.NotificationPreferences(ctx, scope)
	if err != nil || !found || unchanged.Revision != 2 || unchanged.Settings.Defaults.AdvanceDays == 99 {
		t.Fatal("failed audit committed preferences")
	}
	record.ID = "preferences-stale"
	if err := store.SaveNotificationPreferences(ctx, changed, 1, record); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("stale write %v", err)
	}
	rows, err := store.ListNotificationRecipients(ctx, "", 10)
	if err != nil || len(rows) != 1 || rows[0].Revision != 2 || !rows[0].Settings.ForType("medicine").Enabled || rows[0].Settings.Defaults.Enabled {
		t.Fatalf("recipient discovery %+v %v", rows, err)
	}
}
