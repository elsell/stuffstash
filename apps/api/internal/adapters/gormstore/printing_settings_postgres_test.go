package gormstore

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/idgen"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm/clause"
)

func TestPostgresPrintSettingsConcurrentInitializationKeepsOneAuditedWinner(t *testing.T) {
	dsn := os.Getenv("STUFF_STASH_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	db, err := OpenPostgres(dsn)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err = runEmbeddedPostgresMigrations(db); err != nil {
		t.Fatal(err)
	}
	store := NewStore(db)
	ctx := context.Background()
	ids := idgen.ULIDGenerator{}
	scope := printing.Scope{TenantID: ids.NewID(), InventoryID: ids.NewID()}
	saveTenant(t, ctx, store, tenant.ID(scope.TenantID), "Print settings race")
	saveInventory(t, ctx, store, scope.InventoryID, tenant.ID(scope.TenantID), "Inventory")
	t.Cleanup(func() {
		for _, model := range []any{&printSettingsModel{}, &auditRecordModel{}, &inventoryModel{}} {
			if err := db.Where(clause.Eq{Column: "tenant_id", Value: scope.TenantID}).Delete(model).Error; err != nil {
				t.Error(err)
			}
		}
		if err := db.Where(clause.Eq{Column: "id", Value: scope.TenantID}).Delete(&tenantModel{}).Error; err != nil {
			t.Error(err)
		}
	})
	start := make(chan struct{})
	results := make(chan error, 2)
	var done sync.WaitGroup
	for index := range 2 {
		settings := printing.DefaultInventoryPrintSettings(scope)
		settings.Revision = 1
		settings.UpdatedAt = time.Now().UTC()
		settings.Template.Options.ShowReference = index == 0
		record := auditRecord(t, ids.NewID(), tenant.ID(scope.TenantID), inventory.InventoryID(scope.InventoryID), audit.ActionPrintSettingsUpdated)
		done.Add(1)
		go func() {
			defer done.Done()
			<-start
			_, err := store.SavePrintSettings(ctx, settings, 0, nil, record)
			results <- err
		}()
	}
	close(start)
	done.Wait()
	close(results)
	winners, conflicts := 0, 0
	for err := range results {
		if err == nil {
			winners++
		} else if errors.Is(err, ports.ErrPrintConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if winners != 1 || conflicts != 1 {
		t.Fatalf("initial settings overwrote concurrent change: winners%d conflicts%d", winners, conflicts)
	}
	var history int64
	if err = db.Model(&auditRecordModel{}).Where(clause.Eq{Column: "tenant_id", Value: scope.TenantID}).Where(clause.Eq{Column: "action", Value: string(audit.ActionPrintSettingsUpdated)}).Count(&history).Error; err != nil {
		t.Fatal(err)
	}
	if history != 1 {
		t.Fatalf("settings and audit were not atomic: %d records", history)
	}
}
