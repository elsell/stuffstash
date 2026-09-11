package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm/clause"
	"os"
	"testing"
	"time"
)

func TestPostgresNotificationUnreadPreservesAuditedHistory(t *testing.T) {
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
	scope := ports.NotificationScope{TenantID: "unread-home", InventoryID: "unread-inventory", PrincipalID: "owner"}
	cleanup := func() {
		for _, model := range []any{&notificationInboxModel{}, &auditRecordModel{}, &inventoryModel{}, &tenantModel{}} {
			column := "tenant_id"
			if _, ok := model.(*tenantModel); ok {
				column = "id"
			}
			if err := db.Where(clause.Eq{Column: column, Value: scope.TenantID.String()}).Delete(model).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	cleanup()
	t.Cleanup(cleanup)
	ctx := context.Background()
	store := NewStore(db)
	saveTenant(t, ctx, store, scope.TenantID, "Home")
	saveInventory(t, ctx, store, scope.InventoryID.String(), scope.TenantID, "Inventory")
	now := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	date, _ := expirationdate.ParseDate("2026-09-12", expirationdate.Day)
	notice := ports.NotificationRecord{ID: "unread-notice", Scope: scope, Milestone: notification.Milestone{AssetID: "bottle", Date: date, Kind: notification.MilestoneUpcoming}, CreatedAt: now}
	record := auditRecord(t, "unread-created", scope.TenantID, scope.InventoryID, audit.ActionNotificationCreated)
	record.PrincipalID = "owner"
	record.TargetType = audit.TargetNotification
	record.TargetID = notice.ID
	if _, _, err := store.InsertNotification(ctx, notice, record); err != nil {
		t.Fatal(err)
	}
	record.ID = "unread-read"
	record.Action = audit.ActionNotificationRead
	if changed, err := store.MarkNotificationRead(ctx, scope, notice.ID, now, record); err != nil || !changed {
		t.Fatalf("read %v %v", changed, err)
	}
	record.ID = "unread-restored"
	record.Action = audit.ActionNotificationUnread
	if changed, err := store.MarkNotificationUnread(ctx, scope, notice.ID, record); err != nil || !changed {
		t.Fatalf("unread %v %v", changed, err)
	}
	if changed, err := store.MarkNotificationUnread(ctx, scope, notice.ID, record); err != nil || changed {
		t.Fatalf("idempotent unread %v %v", changed, err)
	}
	current, found, err := store.NotificationByID(ctx, scope, notice.ID)
	if err != nil || !found || current.ReadAt != nil || !current.CreatedAt.Equal(notice.CreatedAt) {
		t.Fatal("unread changed original notification")
	}
	var count int64
	if err := db.Model(&auditRecordModel{}).Where(clause.Eq{Column: "id", Value: record.ID}).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("unread audit missing: %v %d", err, count)
	}
}
