package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/adapters/idgen"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"gorm.io/gorm/clause"
	"os"
	"sync"
	"testing"
	"time"
)

func TestPostgresLabelProvisionIsAtomicAcrossProcesses(t *testing.T) {
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
	if err := runEmbeddedPostgresMigrations(db); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	store := NewStore(db)
	tid := tenant.ID("label-concurrency-tenant")
	iid := inventory.InventoryID("label-concurrency-inv")
	cleanup := func() {
		for _, model := range []any{&labelRenderModel{}, &labelModel{}, &auditRecordModel{}, &assetModel{}, &inventoryModel{}, &tenantModel{}} {
			column := "tenant_id"
			if _, ok := model.(*tenantModel); ok {
				column = "id"
			}
			if err := db.Where(clause.Eq{Column: column, Value: tid.String()}).Delete(model).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	cleanup()
	t.Cleanup(cleanup)
	saveTenant(t, ctx, store, tid, "Labels")
	saveInventory(t, ctx, store, iid.String(), tid, "Tools")
	item := assetItem("label-concurrency-asset", tid.String(), iid.String(), asset.KindItem, "")
	if err := createAsset(t, ctx, store, item); err != nil {
		t.Fatal(err)
	}
	ids := idgen.NewULIDGenerator()
	instance, err := store.BootstrapLabelInstance(ctx, printing.InstanceID(ids.NewID()))
	if err != nil {
		t.Fatal(err)
	}
	type result struct {
		label   printing.Label
		created bool
		err     error
	}
	results := make(chan result, 4)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		value := printing.Label{ID: printing.LabelID(ids.NewID()), InstanceID: instance, TenantID: tid.String(), InventoryID: iid.String(), AssetID: item.ID.String(), CreatedAt: time.Now()}
		record := auditRecord(t, ids.NewID(), tid, iid, audit.ActionLabelProvisioned)
		record.TargetID = item.ID.String()
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, c, e := NewStore(db).ProvisionLabel(ctx, value, record)
			results <- result{v, c, e}
		}()
	}
	wg.Wait()
	close(results)
	var canonical printing.LabelID
	created := 0
	for result := range results {
		if result.err != nil {
			t.Fatal(result.err)
		}
		if canonical == "" {
			canonical = result.label.ID
		}
		if canonical != result.label.ID {
			t.Fatal("concurrent provisioning created multiple canonical identities")
		}
		if result.created {
			created++
		}
	}
	if created != 1 {
		t.Fatalf("created %d labels", created)
	}
	var count int64
	if err := db.Model(&auditRecordModel{}).Where(&auditRecordModel{TenantID: tid.String(), Action: audit.ActionLabelProvisioned.String()}).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("provisioning audit count %d: %v", count, err)
	}
}
