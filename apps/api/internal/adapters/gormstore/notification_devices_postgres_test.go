package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm/clause"
	"os"
	"testing"
	"time"
)

func TestPostgresNotificationDevicesConcurrentTokenOwnership(t *testing.T) {
	dsn := os.Getenv("STUFF_STASH_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	db, err := OpenPostgres(dsn)
	if err != nil {
		t.Fatal(err)
	}
	connection, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	if err := runEmbeddedPostgresMigrations(db); err != nil {
		t.Fatal(err)
	}
	cleanup := func() {
		for _, model := range []any{&notificationDeviceModel{}, &auditRecordModel{}, &inventoryModel{}, &tenantModel{}} {
			column := "tenant_id"
			if _, ok := model.(*tenantModel); ok {
				column = "id"
			}
			if err := db.Where(clause.Eq{Column: column, Value: "push-device-home"}).Delete(model).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	cleanup()
	t.Cleanup(cleanup)
	ctx := context.Background()
	store := NewStore(db)
	saveTenant(t, ctx, store, "push-device-home", "Home")
	saveInventory(t, ctx, store, "push-device-inventory", "push-device-home", "Inventory")
	token, _ := notification.ParseDeviceToken("concurrent-native-token")
	now := time.Now().UTC()
	start := make(chan struct{})
	results := make(chan error, 2)
	for _, owner := range []identity.PrincipalID{"first-owner", "second-owner"} {
		value := ports.NotificationDevice{ID: owner.String(), Scope: ports.NotificationScope{TenantID: "push-device-home", InventoryID: "push-device-inventory", PrincipalID: owner}, InstallationID: "installation", Transport: notification.PushAPNS, Token: token, Active: true, Revision: 1, CreatedAt: now, UpdatedAt: now}
		record := auditRecord(t, owner.String(), value.Scope.TenantID, value.Scope.InventoryID, audit.ActionNotificationPreferencesUpdated)
		record.PrincipalID = audit.PrincipalID(owner)
		go func() { <-start; results <- store.SaveNotificationDevice(ctx, value, 0, record) }()
	}
	close(start)
	wins, conflicts := 0, 0
	for range 2 {
		err := <-results
		if err == nil {
			wins++
		} else if errors.Is(err, ports.ErrConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if wins != 1 || conflicts != 1 {
		t.Fatalf("ownership: %d wins %d conflicts", wins, conflicts)
	}
	var count int64
	if err := db.Model(&notificationDeviceModel{}).Where(&notificationDeviceModel{TenantID: "push-device-home", Active: true}).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("active registrations %d %v", count, err)
	}
}
