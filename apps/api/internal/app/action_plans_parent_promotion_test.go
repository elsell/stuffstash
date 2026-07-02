package app

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestExecuteActionPlanPromotesItemParentWhenCreatingAsset(t *testing.T) {
	t.Parallel()

	parent := assetItem("asset-parent", "tenant-home", "inventory-home", asset.KindItem, "")
	repository := &fakeActionPlanRepository{
		records: map[string]ports.ActionPlanRecord{
			"plan-1": actionPlanRecordWithCommand("plan-1", actionplan.StateApproved, actionplan.CommandKindCreateAsset, `{"title":"Milk","kind":"item","parentAssetId":"asset-parent"}`),
		},
	}
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{parent.ID: parent}}
	application := newActionPlanExecutionTestApp(repository, assets, &fakeIDGenerator{ids: []string{"asset-1", "audit-parent-promotion", "undo-1", "audit-1"}})

	executed, err := application.ExecuteActionPlan(context.Background(), ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "plan-1",
	})
	if err != nil {
		t.Fatalf("execute action plan: %v", err)
	}
	if executed.State != actionplan.StateExecuted {
		t.Fatalf("unexpected executed plan: %+v", executed)
	}
	if assets.items[parent.ID].Kind != asset.KindContainer {
		t.Fatalf("expected action plan create to promote parent, got %+v", assets.items[parent.ID])
	}
	created := assets.items[asset.ID("asset-1")]
	if created.ParentAssetID != parent.ID {
		t.Fatalf("expected created asset under promoted parent, got %+v", created)
	}
	if len(assets.auditRecords) != 2 || assets.auditRecords[0].Action != audit.ActionAssetUpdated || assets.auditRecords[1].Action != audit.ActionAssetCreated {
		t.Fatalf("expected promotion and create audit records, got %+v", assets.auditRecords)
	}
}
