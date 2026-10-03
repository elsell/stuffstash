package gormstore

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/idgen"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm/clause"
)

func TestPostgresConcurrentCreateAndPrintCommitsOneAssetAndJob(t *testing.T) {
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
	s := NewStore(db)
	ctx := context.Background()
	ids := idgen.ULIDGenerator{}
	scope := printing.Scope{TenantID: ids.NewID(), InventoryID: ids.NewID()}
	tid, iid := tenant.ID(scope.TenantID), inventory.InventoryID(scope.InventoryID)
	saveTenant(t, ctx, s, tid, "Atomic print")
	saveInventory(t, ctx, s, scope.InventoryID, tid, "Garage")
	printer, err := printingPrinterFromDomain(printing.Printer{ID: printing.PrinterID(ids.NewID()), Scope: scope, Revision: 1, Readiness: printing.PrinterUnknown, MediaFingerprint: "media"})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Create(&printer).Error; err != nil {
		t.Fatal(err)
	}
	instance, err := s.BootstrapLabelInstance(ctx, printing.InstanceID(ids.NewID()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, model := range []any{&printingJobModel{}, &labelModel{}, &assetModel{}, &printingPrinterModel{}, &auditRecordModel{}, &inventoryModel{}} {
			if err := db.Where(clause.Eq{Column: "tenant_id", Value: scope.TenantID}).Delete(model).Error; err != nil {
				t.Error(err)
			}
		}
		if err := db.Where(clause.Eq{Column: "id", Value: scope.TenantID}).Delete(&tenantModel{}).Error; err != nil {
			t.Error(err)
		}
	})
	inputs := make([]ports.PreparedAssetPrint, 2)
	for index := range inputs {
		item := asset.Asset{ID: asset.ID(ids.NewID()), TenantID: asset.TenantID(tid), InventoryID: asset.InventoryID(iid), Kind: asset.KindItem, Title: "Tools", LifecycleState: asset.LifecycleStateActive}
		label := printing.Label{ID: printing.LabelID(ids.NewID()), InstanceID: instance, TenantID: scope.TenantID, InventoryID: scope.InventoryID, AssetID: item.ID.String()}
		job := printing.Job{ID: printing.JobID(ids.NewID()), Scope: scope, PrinterID: printing.PrinterID(printer.ID), AssetID: item.ID.String(), LabelReference: string(label.ID), RequestedBy: "owner", IdempotencyKey: "same-create", Status: printing.JobQueued, Revision: 1, Copies: 1, MediaFingerprint: "media", Artifact: printing.Artifact{ExpiresAt: time.Now().Add(time.Hour)}}
		inputs[index] = ports.PreparedAssetPrint{Asset: ports.PreparedCreateAsset{Asset: item, AuditRecord: auditRecord(t, ids.NewID(), tid, iid, audit.ActionAssetCreated)}, Label: label, LabelAudit: auditRecord(t, ids.NewID(), tid, iid, audit.ActionLabelProvisioned), Job: ports.PrintJobCreate{Job: job, Content: []byte("immutable"), PrinterRevision: 1, RequestFingerprint: "same-payload", Audit: auditRecord(t, ids.NewID(), tid, iid, audit.ActionPrintJobQueued)}}
		inputs[index].Asset.AuditRecord.TargetID = item.ID.String()
		inputs[index].LabelAudit.TargetID = item.ID.String()
		inputs[index].Job.Audit.TargetID = string(job.ID)
	}
	start := make(chan struct{})
	results := make(chan ports.AssetPrintResult, 2)
	failures := make(chan error, 2)
	var done sync.WaitGroup
	for _, input := range inputs {
		done.Add(1)
		go func() {
			defer done.Done()
			<-start
			result, err := s.CreateAssetWithPrint(ctx, input)
			results <- result
			failures <- err
		}()
	}
	close(start)
	done.Wait()
	close(results)
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	var original ports.AssetPrintResult
	created := 0
	for result := range results {
		if result.Created {
			created++
		}
		if original.Asset.ID != "" && (result.Asset.ID != original.Asset.ID || result.Job.ID != original.Job.ID) {
			t.Fatal("concurrent retry returned a different creation")
		}
		original = result
	}
	if created != 1 {
		t.Fatalf("created %d times", created)
	}
	for _, model := range []any{&assetModel{}, &labelModel{}, &printingJobModel{}} {
		var count int64
		if err := db.Model(model).Where(clause.Eq{Column: "tenant_id", Value: scope.TenantID}).Count(&count).Error; err != nil || count != 1 {
			t.Fatalf("duplicate or partial aggregate %T: count=%d err=%v", model, count, err)
		}
	}
}
