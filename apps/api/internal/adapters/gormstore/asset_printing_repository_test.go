package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestCreateAssetPrintRollsBackLateFailureAndReplaysOriginalResult(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t, ctx)
	saveTenant(t, ctx, s, "tenant", "Home")
	saveInventory(t, ctx, s, "inventory", "tenant", "Garage")
	scope := printing.Scope{TenantID: "tenant", InventoryID: "inventory"}
	p, _ := printingPrinterFromDomain(printing.Printer{ID: "printer", Scope: scope, Revision: 1, MediaFingerprint: "media"})
	if err := s.db.Create(&p).Error; err != nil {
		t.Fatal(err)
	}
	instance, err := s.BootstrapLabelInstance(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAV")
	if err != nil {
		t.Fatal(err)
	}
	item := asset.Asset{ID: "new", TenantID: "tenant", InventoryID: "inventory", Kind: asset.KindItem, Title: "Tools", LifecycleState: asset.LifecycleStateActive}
	ar := auditRecord(t, "asset-audit", "tenant", "inventory", audit.ActionAssetCreated)
	ar.TargetID = "new"
	lr := auditRecord(t, "label-audit", "tenant", "inventory", audit.ActionLabelProvisioned)
	lr.TargetID = "new"
	jr := auditRecord(t, "job-audit", "tenant", "inventory", audit.ActionPrintJobQueued)
	jr.TargetID = "job"
	in := ports.PreparedAssetPrint{Asset: ports.PreparedCreateAsset{Asset: item, AuditRecord: ar}, Label: printing.Label{ID: "01ARZ3NDEKTSV4RRFFQ69G5FAW", InstanceID: instance, TenantID: "tenant", InventoryID: "inventory", AssetID: "new"}, LabelAudit: lr, Job: ports.PrintJobCreate{Job: printing.Job{ID: "job", Scope: scope, PrinterID: "printer", AssetID: "new", LabelReference: "01ARZ3NDEKTSV4RRFFQ69G5FAW", RequestedBy: "owner", IdempotencyKey: "request", Status: printing.JobQueued, Revision: 1, Copies: 1, MediaFingerprint: "media", Artifact: printing.Artifact{ExpiresAt: time.Now().Add(time.Hour)}}, PrinterRevision: 1, RequestFingerprint: "payload", Content: []byte("immutable"), Audit: jr}}
	bad := in
	bad.Job.Audit.ID = lr.ID
	if _, err = s.CreateAssetWithPrint(ctx, bad); err == nil {
		t.Fatal("late audit collision committed")
	}
	if _, found, err := s.AssetByID(ctx, "tenant", "inventory", "new"); err != nil || found {
		t.Fatal("partial asset remained", err)
	}
	if _, found, err := s.LabelForAsset(ctx, "tenant", "inventory", "new"); err != nil || found {
		t.Fatal("partial label remained", err)
	}
	first, err := s.CreateAssetWithPrint(ctx, in)
	if err != nil || !first.Created {
		t.Fatalf("create %v %+v", err, first)
	}
	in.Asset.Asset.ID = "different-prepared-id"
	in.Job.Job.ID = "different-job"
	replay, err := s.CreateAssetWithPrint(ctx, in)
	if err != nil || replay.Created || replay.Asset.ID != "new" || replay.Job.ID != "job" {
		t.Fatalf("replay %+v %v", replay, err)
	}
	in.Job.RequestFingerprint = "changed"
	if _, err = s.CreateAssetWithPrint(ctx, in); err == nil {
		t.Fatal("changed request accepted")
	}
}
