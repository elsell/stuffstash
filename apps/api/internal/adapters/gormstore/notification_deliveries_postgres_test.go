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

func TestPostgresNotificationDeliveryConcurrentPublishAndClaim(t *testing.T) {
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
	scope := ports.NotificationScope{TenantID: "push-queue-home", InventoryID: "push-queue-inventory", PrincipalID: "owner"}
	cleanup := func() {
		for _, model := range []any{&notificationDeliveryModel{}, &notificationInboxModel{}, &notificationDeviceModel{}, &auditRecordModel{}, &inventoryModel{}, &tenantModel{}} {
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
	now := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	token, _ := notification.ParseDeviceToken("face")
	device := ports.NotificationDevice{ID: "queue-phone", Scope: scope, InstallationID: "phone", Transport: notification.PushAPNS, Token: token, Active: true, Revision: 1, CreatedAt: now, UpdatedAt: now}
	record := auditRecord(t, "queue-device-audit", scope.TenantID, scope.InventoryID, audit.ActionNotificationDeviceUpdated)
	record.PrincipalID = audit.PrincipalID(scope.PrincipalID)
	if err := store.SaveNotificationDevice(ctx, device, 0, record); err != nil {
		t.Fatal(err)
	}
	date, _ := expirationdate.ParseDate("2026-09-20", expirationdate.Day)
	notice := ports.NotificationRecord{ID: "queue-notice", Scope: scope, CreatedAt: now, Milestone: notification.Milestone{AssetID: "item", Date: date, Kind: notification.MilestoneUpcoming}}
	state, _ := notification.NewDelivery(now)
	delivery := ports.NotificationDelivery{ID: "queue-delivery", Scope: scope, NotificationID: notice.ID, DeviceID: device.ID, DeviceRevision: 1, CreatedAt: now, State: state}
	record.ID = "queue-notice-audit"
	record.Action = audit.ActionNotificationCreated
	type publication struct {
		created bool
		err     error
	}
	start := make(chan struct{})
	publications := make(chan publication, 2)
	for range 2 {
		go func() {
			<-start
			_, created, err := store.InsertNotificationWithDeliveries(ctx, notice, []ports.NotificationDelivery{delivery}, record)
			publications <- publication{created, err}
		}()
	}
	close(start)
	created := 0
	for range 2 {
		result := <-publications
		if result.err != nil {
			t.Fatal(result.err)
		}
		if result.created {
			created++
		}
	}
	if created != 1 {
		t.Fatal("concurrent publication did not deduplicate")
	}
	// A fresh store instance sees work committed by the publisher.
	reopened := NewStore(db)
	policy := notification.RetryPolicy{MaxAttempts: 2, InitialDelay: time.Second, MaximumDelay: time.Minute}
	type claim struct {
		rows []ports.NotificationDelivery
		err  error
	}
	claims := make(chan claim, 2)
	start = make(chan struct{})
	for _, fence := range []string{"worker-a", "worker-b"} {
		go func() {
			<-start
			rows, err := reopened.ClaimNotificationDeliveries(ctx, now, fence, time.Minute, policy, 10)
			claims <- claim{rows, err}
		}()
	}
	close(start)
	claimed := 0
	for range 2 {
		result := <-claims
		if result.err != nil {
			t.Fatal(result.err)
		}
		claimed += len(result.rows)
	}
	if claimed != 1 {
		t.Fatalf("workers claimed %d deliveries", claimed)
	}
	var count int64
	if err := db.Model(&notificationDeliveryModel{}).Where(&notificationDeliveryModel{NotificationID: notice.ID}).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("delivery persistence: %d %v", count, err)
	}
}
