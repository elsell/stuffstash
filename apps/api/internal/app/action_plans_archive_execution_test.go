package app

import (
	"context"
	"errors"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestExecuteActionPlanArchivesAssetAndMarksExecuted(t *testing.T) {
	t.Parallel()

	repository := &fakeActionPlanRepository{
		records: map[string]ports.ActionPlanRecord{
			"plan-1": actionPlanRecordWithCommand("plan-1", actionplan.StateApproved, actionplan.CommandKindArchiveAsset, `{"assetId":"asset-1"}`),
		},
	}
	item := assetItem("asset-1", "tenant-home", "inventory-home", asset.KindItem, "")
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		item.ID: item,
	}}
	application := newActionPlanExecutionTestApp(repository, assets, &fakeIDGenerator{ids: []string{"undo-1", "audit-1"}})

	executed, err := application.ExecuteActionPlan(context.Background(), ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "plan-1",
	})
	if err != nil {
		t.Fatalf("execute action plan: %v", err)
	}
	if executed.State != actionplan.StateExecuted || executed.ExecutedAt.IsZero() {
		t.Fatalf("unexpected executed plan: %+v", executed)
	}
	archived := assets.items[asset.ID("asset-1")]
	if archived.LifecycleState != asset.LifecycleStateArchived {
		t.Fatalf("expected archived asset, got %+v", archived)
	}
	if len(assets.auditRecords) != 1 || assets.auditRecords[0].Action != audit.ActionAssetArchived {
		t.Fatalf("expected audited archive through asset service, got %+v", assets.auditRecords)
	}
	if len(assets.undoables) != 1 || assets.undoables["undo-1"].OriginalAction != audit.ActionAssetArchived {
		t.Fatalf("expected archive undoable operation, got %+v", assets.undoables)
	}
}

func TestExecuteArchiveActionPlanAllowsEditAccessWithoutCreateAccess(t *testing.T) {
	t.Parallel()

	repository := &fakeActionPlanRepository{
		records: map[string]ports.ActionPlanRecord{
			"plan-1": actionPlanRecordWithCommand("plan-1", actionplan.StateApproved, actionplan.CommandKindArchiveAsset, `{"assetId":"asset-1"}`),
		},
	}
	item := assetItem("asset-1", "tenant-home", "inventory-home", asset.KindItem, "")
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		item.ID: item,
	}}
	application := newActionPlanExecutionTestAppWithAuthorizer(repository, assets, &fakeIDGenerator{ids: []string{"undo-1", "audit-1"}}, &permissionAuthorizer{
		allowed: map[ports.InventoryPermission]struct{}{
			ports.InventoryPermissionEditAsset: {},
			ports.InventoryPermissionView:      {},
		},
	})

	executed, err := application.ExecuteActionPlan(context.Background(), ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "plan-1",
	})
	if err != nil {
		t.Fatalf("execute archive action plan: %v", err)
	}
	if executed.State != actionplan.StateExecuted {
		t.Fatalf("expected executed plan, got %+v", executed)
	}
	if assets.items[asset.ID("asset-1")].LifecycleState != asset.LifecycleStateArchived {
		t.Fatalf("expected edit-authorized archive to update lifecycle, got %+v", assets.items[asset.ID("asset-1")])
	}
}

func TestExecuteArchiveActionPlanRequiresEditAccessWithoutMutatingAsset(t *testing.T) {
	t.Parallel()

	repository := &fakeActionPlanRepository{
		records: map[string]ports.ActionPlanRecord{
			"plan-1": actionPlanRecordWithCommand("plan-1", actionplan.StateApproved, actionplan.CommandKindArchiveAsset, `{"assetId":"asset-1"}`),
		},
	}
	item := assetItem("asset-1", "tenant-home", "inventory-home", asset.KindItem, "")
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		item.ID: item,
	}}
	application := newActionPlanExecutionTestAppWithAuthorizer(repository, assets, &fakeIDGenerator{ids: []string{"undo-1", "audit-1"}}, &permissionAuthorizer{
		allowed: map[ports.InventoryPermission]struct{}{
			ports.InventoryPermissionCreateAsset: {},
			ports.InventoryPermissionView:        {},
		},
	})

	_, err := application.ExecuteActionPlan(context.Background(), ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "plan-1",
	})
	if !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected forbidden archive execution, got %v", err)
	}
	if repository.records["plan-1"].State != actionplan.StateApproved {
		t.Fatalf("expected unauthorized archive plan to remain approved, got %+v", repository.records["plan-1"])
	}
	if assets.items[asset.ID("asset-1")].LifecycleState != asset.LifecycleStateActive {
		t.Fatalf("expected unauthorized archive to leave asset active, got %+v", assets.items[asset.ID("asset-1")])
	}
	if len(assets.auditRecords) != 0 || len(assets.undoables) != 0 {
		t.Fatalf("expected no audit or undoable records, got audit=%+v undoables=%+v", assets.auditRecords, assets.undoables)
	}
}

func TestExecuteArchiveActionPlanRejectsActiveChildrenAndMarksFailed(t *testing.T) {
	t.Parallel()

	repository := &fakeActionPlanRepository{
		records: map[string]ports.ActionPlanRecord{
			"plan-1": actionPlanRecordWithCommand("plan-1", actionplan.StateApproved, actionplan.CommandKindArchiveAsset, `{"assetId":"asset-1"}`),
		},
	}
	item := assetItem("asset-1", "tenant-home", "inventory-home", asset.KindContainer, "")
	child := assetItem("child-1", "tenant-home", "inventory-home", asset.KindItem, "asset-1")
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		item.ID:  item,
		child.ID: child,
	}}
	application := newActionPlanExecutionTestApp(repository, assets, &fakeIDGenerator{ids: []string{"undo-1", "audit-1"}})

	failed, err := application.ExecuteActionPlan(context.Background(), ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "plan-1",
	})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected conflict for active child archive, got %v", err)
	}
	if failed.State != actionplan.StateFailed || failed.FailedAt.IsZero() {
		t.Fatalf("expected failed plan state, got %+v", failed)
	}
	if assets.items[asset.ID("asset-1")].LifecycleState != asset.LifecycleStateActive {
		t.Fatalf("expected archive rejection to leave parent active, got %+v", assets.items[asset.ID("asset-1")])
	}
	if len(assets.auditRecords) != 0 || len(assets.undoables) != 0 {
		t.Fatalf("expected no audit or undoable records, got audit=%+v undoables=%+v", assets.auditRecords, assets.undoables)
	}
}

func TestExecuteActionPlanRestoresAssetAndMarksExecuted(t *testing.T) {
	t.Parallel()

	repository := &fakeActionPlanRepository{
		records: map[string]ports.ActionPlanRecord{
			"plan-1": actionPlanRecordWithCommand("plan-1", actionplan.StateApproved, actionplan.CommandKindRestoreAsset, `{"assetId":"asset-1"}`),
		},
	}
	item := assetItem("asset-1", "tenant-home", "inventory-home", asset.KindItem, "")
	item.LifecycleState = asset.LifecycleStateArchived
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		item.ID: item,
	}}
	application := newActionPlanExecutionTestApp(repository, assets, &fakeIDGenerator{ids: []string{"undo-1", "audit-1"}})

	executed, err := application.ExecuteActionPlan(context.Background(), ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "plan-1",
	})
	if err != nil {
		t.Fatalf("execute action plan: %v", err)
	}
	if executed.State != actionplan.StateExecuted || executed.ExecutedAt.IsZero() {
		t.Fatalf("unexpected executed plan: %+v", executed)
	}
	restored := assets.items[asset.ID("asset-1")]
	if restored.LifecycleState != asset.LifecycleStateActive {
		t.Fatalf("expected restored asset, got %+v", restored)
	}
	if len(assets.auditRecords) != 1 || assets.auditRecords[0].Action != audit.ActionAssetRestored {
		t.Fatalf("expected audited restore through asset service, got %+v", assets.auditRecords)
	}
	if len(assets.undoables) != 1 || assets.undoables["undo-1"].OriginalAction != audit.ActionAssetRestored {
		t.Fatalf("expected restore undoable operation, got %+v", assets.undoables)
	}
}

func TestExecuteRestoreActionPlanAllowsEditAccessWithoutCreateAccess(t *testing.T) {
	t.Parallel()

	repository := &fakeActionPlanRepository{
		records: map[string]ports.ActionPlanRecord{
			"plan-1": actionPlanRecordWithCommand("plan-1", actionplan.StateApproved, actionplan.CommandKindRestoreAsset, `{"assetId":"asset-1"}`),
		},
	}
	item := assetItem("asset-1", "tenant-home", "inventory-home", asset.KindItem, "")
	item.LifecycleState = asset.LifecycleStateArchived
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		item.ID: item,
	}}
	application := newActionPlanExecutionTestAppWithAuthorizer(repository, assets, &fakeIDGenerator{ids: []string{"undo-1", "audit-1"}}, &permissionAuthorizer{
		allowed: map[ports.InventoryPermission]struct{}{
			ports.InventoryPermissionEditAsset: {},
			ports.InventoryPermissionView:      {},
		},
	})

	executed, err := application.ExecuteActionPlan(context.Background(), ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "plan-1",
	})
	if err != nil {
		t.Fatalf("execute restore action plan: %v", err)
	}
	if executed.State != actionplan.StateExecuted {
		t.Fatalf("expected executed plan, got %+v", executed)
	}
	if assets.items[asset.ID("asset-1")].LifecycleState != asset.LifecycleStateActive {
		t.Fatalf("expected edit-authorized restore to update lifecycle, got %+v", assets.items[asset.ID("asset-1")])
	}
}

func TestExecuteRestoreActionPlanRequiresEditAccessWithoutMutatingAsset(t *testing.T) {
	t.Parallel()

	repository := &fakeActionPlanRepository{
		records: map[string]ports.ActionPlanRecord{
			"plan-1": actionPlanRecordWithCommand("plan-1", actionplan.StateApproved, actionplan.CommandKindRestoreAsset, `{"assetId":"asset-1"}`),
		},
	}
	item := assetItem("asset-1", "tenant-home", "inventory-home", asset.KindItem, "")
	item.LifecycleState = asset.LifecycleStateArchived
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		item.ID: item,
	}}
	application := newActionPlanExecutionTestAppWithAuthorizer(repository, assets, &fakeIDGenerator{ids: []string{"undo-1", "audit-1"}}, &permissionAuthorizer{
		allowed: map[ports.InventoryPermission]struct{}{
			ports.InventoryPermissionCreateAsset: {},
			ports.InventoryPermissionView:        {},
		},
	})

	_, err := application.ExecuteActionPlan(context.Background(), ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "plan-1",
	})
	if !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected forbidden restore execution, got %v", err)
	}
	if repository.records["plan-1"].State != actionplan.StateApproved {
		t.Fatalf("expected unauthorized restore plan to remain approved, got %+v", repository.records["plan-1"])
	}
	if assets.items[asset.ID("asset-1")].LifecycleState != asset.LifecycleStateArchived {
		t.Fatalf("expected unauthorized restore to leave asset archived, got %+v", assets.items[asset.ID("asset-1")])
	}
	if len(assets.auditRecords) != 0 || len(assets.undoables) != 0 {
		t.Fatalf("expected no audit or undoable records, got audit=%+v undoables=%+v", assets.auditRecords, assets.undoables)
	}
}

func TestExecuteRestoreActionPlanRejectsArchivedParentAndMarksFailed(t *testing.T) {
	t.Parallel()

	repository := &fakeActionPlanRepository{
		records: map[string]ports.ActionPlanRecord{
			"plan-1": actionPlanRecordWithCommand("plan-1", actionplan.StateApproved, actionplan.CommandKindRestoreAsset, `{"assetId":"asset-1"}`),
		},
	}
	parent := assetItem("parent-1", "tenant-home", "inventory-home", asset.KindContainer, "")
	parent.LifecycleState = asset.LifecycleStateArchived
	item := assetItem("asset-1", "tenant-home", "inventory-home", asset.KindItem, "parent-1")
	item.LifecycleState = asset.LifecycleStateArchived
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		parent.ID: parent,
		item.ID:   item,
	}}
	application := newActionPlanExecutionTestApp(repository, assets, &fakeIDGenerator{ids: []string{"undo-1", "audit-1"}})

	failed, err := application.ExecuteActionPlan(context.Background(), ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "plan-1",
	})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected conflict for archived parent restore, got %v", err)
	}
	if failed.State != actionplan.StateFailed || failed.FailedAt.IsZero() {
		t.Fatalf("expected failed plan state, got %+v", failed)
	}
	if assets.items[asset.ID("asset-1")].LifecycleState != asset.LifecycleStateArchived {
		t.Fatalf("expected restore rejection to leave asset archived, got %+v", assets.items[asset.ID("asset-1")])
	}
	if len(assets.auditRecords) != 0 || len(assets.undoables) != 0 {
		t.Fatalf("expected no audit or undoable records, got audit=%+v undoables=%+v", assets.auditRecords, assets.undoables)
	}
}
