package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

func TestNotificationPreferencesAuditFailureRollsBack(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	scope := ports.NotificationScope{TenantID: "home", InventoryID: "main", PrincipalID: "owner"}
	saveMemoryTenant(t, ctx, store, scope.TenantID)
	saveMemoryInventory(t, ctx, store, scope.TenantID, scope.InventoryID)
	value := ports.NotificationPreferencesRecord{ID: "preferences", Scope: scope, Revision: 1, Settings: notification.DefaultSettings("UTC")}
	record := memoryAuditRecord(t, "preferences-audit", scope.TenantID)
	record.InventoryID = audit.InventoryID(scope.InventoryID.String())
	record.PrincipalID = audit.PrincipalID(scope.PrincipalID.String())
	if err := store.SaveNotificationPreferences(ctx, value, 0, record); err != nil {
		t.Fatal(err)
	}
	changed := value.Clone()
	changed.Revision = 2
	changed.Settings.Defaults.Enabled = false
	if err := store.SaveNotificationPreferences(ctx, changed, 1, record); err == nil {
		t.Fatal("duplicate audit accepted")
	}
	actual, found, err := store.NotificationPreferences(ctx, scope)
	if err != nil || !found || actual.Revision != 1 || !actual.Settings.Defaults.Enabled {
		t.Fatal("failed audit committed preferences")
	}
}
