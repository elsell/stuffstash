package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

func TestAtomicCreatePrintRejectsParentArchivedAfterPreparation(t *testing.T) {
	s := NewStore()
	scope := printing.Scope{TenantID: "tenant", InventoryID: "inventory"}
	s.inventories["inventory"] = inventory.Inventory{ID: "inventory", TenantID: "tenant", LifecycleState: inventory.LifecycleStateActive}
	parent := asset.Asset{ID: "parent", TenantID: "tenant", InventoryID: "inventory", Kind: asset.KindItem, LifecycleState: asset.LifecycleStateArchived}
	s.assets[parent.ID] = parent
	promoted := parent
	promoted.Kind = asset.KindContainer
	promoted.LifecycleState = asset.LifecycleStateActive
	s.labelInstance = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	s.printingPrinters = map[printing.PrinterID]printing.Printer{"printer": {ID: "printer", Scope: scope, Revision: 1, MediaFingerprint: "media"}}
	record := func(id, target string) audit.Record {
		return audit.Record{ID: audit.ID(id), TenantID: "tenant", InventoryID: "inventory", TargetID: target}
	}
	pa := record("parent-audit", "parent")
	input := ports.PreparedAssetPrint{Asset: ports.PreparedCreateAsset{Asset: asset.Asset{ID: "child", TenantID: "tenant", InventoryID: "inventory", ParentAssetID: "parent", Kind: asset.KindItem, LifecycleState: asset.LifecycleStateActive}, AuditRecord: record("asset-audit", "child"), PromotedParent: &promoted, ParentPromotionRecord: &pa}, Label: printing.Label{ID: "01ARZ3NDEKTSV4RRFFQ69G5FAW", InstanceID: s.labelInstance, TenantID: "tenant", InventoryID: "inventory", AssetID: "child"}, LabelAudit: record("label-audit", "child"), Job: ports.PrintJobCreate{Job: printing.Job{ID: "job", Scope: scope, PrinterID: "printer", AssetID: "child", LabelReference: "01ARZ3NDEKTSV4RRFFQ69G5FAW", RequestedBy: "owner", IdempotencyKey: "request", Status: printing.JobQueued, Revision: 1, MediaFingerprint: "media"}, PrinterRevision: 1, RequestFingerprint: "request", Audit: record("job-audit", "job")}}
	if _, err := s.CreateAssetWithPrint(context.Background(), input); err == nil {
		t.Fatal("stale promotion resurrected archived parent")
	}
	if s.assets[parent.ID].LifecycleState != asset.LifecycleStateArchived {
		t.Fatal("parent changed")
	}
	if _, found := s.assets["child"]; found {
		t.Fatal("partial child")
	}
}
