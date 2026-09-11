package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestNotificationInboxRepositoryContract(t *testing.T) {
	ctx := context.Background()
	fake := memory.NewStore()
	name, _ := tenant.NewName("Home")
	if err := fake.SaveTenant(ctx, tenant.Tenant{ID: "tenant-one", Name: name}); err != nil {
		t.Fatal(err)
	}
	invName, _ := inventory.NewName("Main")
	if err := fake.SaveInventory(ctx, inventory.Inventory{ID: "inventory-one", TenantID: "tenant-one", Name: invName}); err != nil {
		t.Fatal(err)
	}
	for label, repository := range map[string]ports.NotificationInboxRepository{"memory": fake, "gorm": newUndoableOperationTestStore(t, ctx)} {
		t.Run(label, func(t *testing.T) {
			scope := ports.NotificationScope{TenantID: "tenant-one", InventoryID: "inventory-one", PrincipalID: "owner"}
			date, _ := expirationdate.ParseDate("2026-09", expirationdate.Month)
			now := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
			value := ports.NotificationRecord{ID: "notice-1", Scope: scope, Milestone: notification.Milestone{AssetID: "bottle", Date: date, Kind: notification.MilestoneUpcoming}, CreatedAt: now}
			record := auditRecord(t, "inbox-created", scope.TenantID, scope.InventoryID, audit.ActionNotificationCreated)
			record.PrincipalID = "owner"
			record.TargetType = audit.TargetNotification
			record.TargetID = value.ID
			inserted, created, err := repository.InsertNotification(ctx, value, record)
			if err != nil || !created || inserted.ID != value.ID {
				t.Fatalf("insert %+v %v %v", inserted, created, err)
			}
			duplicate := value
			duplicate.ID = "notice-duplicate"
			record.ID = "duplicate-audit"
			same, created, err := repository.InsertNotification(ctx, duplicate, record)
			if err != nil || created || same.ID != value.ID {
				t.Fatal("duplicate milestone inserted")
			}
			if _, _, err := repository.NotificationByID(ctx, scope, ""); err == nil {
				t.Fatal("empty id accepted")
			}
			if _, err := repository.MarkNotificationRead(ctx, scope, "", now, record); err == nil {
				t.Fatal("empty id modified all notices")
			}
			other := scope
			other.PrincipalID = "other"
			if _, found, err := repository.NotificationByID(ctx, other, value.ID); err != nil || found {
				t.Fatal("cross-principal read")
			}
			record.ID = "read"
			record.Action = audit.ActionNotificationRead
			record.PrincipalID = "other"
			if _, err := repository.MarkNotificationRead(ctx, other, value.ID, now, record); err == nil {
				t.Fatal("cross-principal mutation accepted")
			}
			beforeRead, _, err := repository.NotificationByID(ctx, scope, value.ID)
			if err != nil || beforeRead.ReadAt != nil {
				t.Fatal("other principal changed owner's read state")
			}
			record.PrincipalID = "owner"
			changed, err := repository.MarkNotificationRead(ctx, scope, value.ID, now, record)
			if err != nil || !changed {
				t.Fatalf("mark read %v %v", changed, err)
			}
			changed, err = repository.MarkNotificationRead(ctx, scope, value.ID, now.Add(time.Hour), record)
			if err != nil || changed {
				t.Fatal("read not idempotent")
			}
			current, found, err := repository.NotificationByID(ctx, scope, value.ID)
			if err != nil || !found || current.ReadAt == nil || !current.ReadAt.Equal(now) {
				t.Fatal("read state lost")
			}
			next := value
			next.ID = "notice-2"
			next.Milestone.Kind = notification.MilestoneExpired
			record.ID = "inbox-created"
			record.Action = audit.ActionNotificationCreated
			if _, _, err := repository.InsertNotification(ctx, next, record); err == nil {
				t.Fatal("audit conflict accepted")
			}
			if _, found, err := repository.NotificationByID(ctx, scope, next.ID); err != nil || found {
				t.Fatal("failed audit committed inbox")
			}
			record.ID = "expired-created"
			if _, created, err := repository.InsertNotification(ctx, next, record); err != nil || !created {
				t.Fatalf("expired insert %v", err)
			}
			record.Action = audit.ActionNotificationRead
			if _, err := repository.MarkNotificationRead(ctx, scope, next.ID, now, record); err == nil {
				t.Fatal("read audit conflict accepted")
			}
			nextRead, _, _ := repository.NotificationByID(ctx, scope, next.ID)
			if nextRead.ReadAt != nil {
				t.Fatal("failed audit marked read")
			}
			page, err := repository.ListNotifications(ctx, scope, "", 1)
			if err != nil || len(page) != 1 || page[0].ID != next.ID {
				t.Fatalf("first page %+v %v", page, err)
			}
			page, err = repository.ListNotifications(ctx, scope, next.ID, 1)
			if err != nil || len(page) != 1 || page[0].ID != value.ID {
				t.Fatal("stable cursor")
			}
			other = scope
			other.TenantID = "elsewhere"
			page, err = repository.ListNotifications(ctx, other, "", 10)
			if err != nil || len(page) != 0 {
				t.Fatal("cross tenant list")
			}
			if _, err := repository.ListNotifications(ctx, scope, "", 0); err == nil {
				t.Fatal("unbounded page accepted")
			}
		})
	}
}
