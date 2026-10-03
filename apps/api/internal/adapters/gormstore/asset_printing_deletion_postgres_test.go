package gormstore

import (
	"context"
	"errors"
	"fmt"
	"github.com/stuffstash/stuff-stash/internal/adapters/idgen"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sync"
	"testing"
	"time"
)

type atomicPrintDeletionContext struct{}

func TestPostgresPrintingCreateAndDeleteShareScopeFirstLockOrder(t *testing.T) {
	store, scope, inputs := postgresAtomicPrintFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Observe scheduling only: every query, insert, FK check and lock still runs
	// against PostgreSQL. The held inventory lock makes deletion win this race.
	waiting := make(chan struct{})
	var once sync.Once
	observe := func(tx *gorm.DB) {
		if tx.Statement.Context.Value(atomicPrintDeletionContext{}) == true && (tx.Statement.Table == "inventories" || tx.Statement.Table == "assets") {
			once.Do(func() { close(waiting) })
		}
	}
	if err := store.db.Callback().Query().Before("gorm:query").Register("printing-deletion-schedule", observe); err != nil {
		t.Fatal(err)
	}
	if err := store.db.Callback().Create().Before("gorm:create").Register("printing-deletion-schedule", observe); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = store.db.Callback().Query().Remove("printing-deletion-schedule")
		_ = store.db.Callback().Create().Remove("printing-deletion-schedule")
	})
	result := make(chan error, 1)
	ids := idgen.ULIDGenerator{}
	err := store.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var scopeRow inventoryModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&inventoryModel{ID: scope.InventoryID, TenantID: scope.TenantID}).First(&scopeRow).Error; err != nil {
			return err
		}
		go func() {
			_, err := store.CreateAssetWithPrint(context.WithValue(ctx, atomicPrintDeletionContext{}, true), inputs[0])
			result <- err
		}()
		select {
		case <-waiting:
		case <-ctx.Done():
			return ctx.Err()
		}
		// A creator waiting for the scope must not already own the printer. NOWAIT
		// proves the order without waiting for PostgreSQL's deadlock detector.
		var printer printingPrinterModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "NOWAIT"}).Where(&printingPrinterModel{ID: string(inputs[0].Job.Job.PrinterID), TenantID: scope.TenantID, InventoryID: scope.InventoryID}).First(&printer).Error; err != nil {
			return fmt.Errorf("create acquired printer before scope: %w", err)
		}
		record := auditRecord(t, ids.NewID(), tenant.ID(scope.TenantID), inventory.InventoryID(scope.InventoryID), audit.ActionInventoryDeleted)
		effects := &ports.InventoryDeletionEffects{Audit: func(action audit.Action, target audit.TargetType, id string) (audit.Record, error) {
			r := auditRecord(t, ids.NewID(), tenant.ID(scope.TenantID), inventory.InventoryID(scope.InventoryID), action)
			r.TargetID = id
			r.TargetType = target
			return r, nil
		}}
		return NewStore(tx).DeleteInventory(ctx, tenant.ID(scope.TenantID), inventory.InventoryID(scope.InventoryID), record, effects)
	})
	if err != nil {
		cancel()
		select {
		case <-result:
		case <-time.After(time.Second):
		}
		t.Fatal(err)
	}
	select {
	case err := <-result:
		if !errors.Is(err, ports.ErrPrintNotFound) {
			t.Fatalf("creation after committed deletion: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("create/delete did not settle")
	}
	for _, model := range []any{&assetModel{}, &labelModel{}, &printingJobModel{}} {
		var count int64
		if err := store.db.Model(model).Where(clause.Eq{Column: "inventory_id", Value: scope.InventoryID}).Count(&count).Error; err != nil || count != 0 {
			t.Fatalf("partial creation after deletion %T: %d %v", model, count, err)
		}
	}
}

func TestPostgresPrintingCreationWinsBeforeScopeDeletion(t *testing.T) {
	store, scope, inputs := postgresAtomicPrintFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	locked := make(chan struct{})
	waiting := make(chan struct{})
	release := make(chan struct{})
	var lockedOnce, waitingOnce, releaseOnce sync.Once
	releaseCreate := func() { releaseOnce.Do(func() { close(release) }) }
	defer releaseCreate()
	hold := func(tx *gorm.DB) {
		if tx.Statement.Context.Value(atomicPrintDeletionContext{}) == "create" && tx.Statement.Table == "inventories" {
			lockedOnce.Do(func() {
				close(locked)
				select {
				case <-release:
				case <-ctx.Done():
				}
			})
		}
	}
	observe := func(tx *gorm.DB) {
		if tx.Statement.Context.Value(atomicPrintDeletionContext{}) == "delete" && tx.Statement.Table == "inventories" {
			waitingOnce.Do(func() { close(waiting) })
		}
	}
	if err := store.db.Callback().Query().After("gorm:query").Register("printing-create-held", hold); err != nil {
		t.Fatal(err)
	}
	if err := store.db.Callback().Query().Before("gorm:query").Register("printing-delete-waiting", observe); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = store.db.Callback().Query().Remove("printing-create-held")
		_ = store.db.Callback().Query().Remove("printing-delete-waiting")
	})
	created := make(chan error, 1)
	deleted := make(chan error, 1)
	go func() {
		_, err := store.CreateAssetWithPrint(context.WithValue(ctx, atomicPrintDeletionContext{}, "create"), inputs[0])
		created <- err
	}()
	select {
	case <-locked:
	case <-ctx.Done():
		t.Fatal("creation did not acquire scope")
	}
	ids := idgen.ULIDGenerator{}
	record := auditRecord(t, ids.NewID(), tenant.ID(scope.TenantID), inventory.InventoryID(scope.InventoryID), audit.ActionInventoryDeleted)
	go func() {
		deleted <- store.DeleteInventory(context.WithValue(ctx, atomicPrintDeletionContext{}, "delete"), tenant.ID(scope.TenantID), inventory.InventoryID(scope.InventoryID), record, nil)
	}()
	select {
	case <-waiting:
	case <-ctx.Done():
		t.Fatal("deletion did not contend for scope")
	}
	releaseCreate()
	select {
	case err := <-created:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("creation did not settle")
	}
	select {
	case err := <-deleted:
		if !errors.Is(err, ports.ErrForbidden) {
			t.Fatalf("deletion did not preserve concurrently created asset: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("deletion did not settle")
	}
	for _, model := range []any{&assetModel{}, &labelModel{}, &printingJobModel{}} {
		var count int64
		if err := store.db.Model(model).Where(clause.Eq{Column: "inventory_id", Value: scope.InventoryID}).Count(&count).Error; err != nil || count != 1 {
			t.Fatalf("partial creation %T: %d %v", model, count, err)
		}
	}
}
