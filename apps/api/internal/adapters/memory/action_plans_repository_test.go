package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestActionPlanRepositoryScopesReadsAndTransitions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewStore()
	record := memoryActionPlanRecord("plan-1", time.Date(2026, 6, 26, 18, 0, 0, 0, time.UTC))
	if err := store.SaveActionPlan(ctx, record); err != nil {
		t.Fatalf("save action plan: %v", err)
	}

	if _, found, err := store.ActionPlanByID(ctx, tenant.ID("tenant-other"), record.InventoryID, record.ID); err != nil || found {
		t.Fatalf("expected wrong tenant read to miss, found=%t err=%v", found, err)
	}
	if _, found, err := store.ActionPlanByID(ctx, record.TenantID, inventory.InventoryID("inventory-other"), record.ID); err != nil || found {
		t.Fatalf("expected wrong inventory read to miss, found=%t err=%v", found, err)
	}

	transition := ports.ActionPlanStateTransition{
		PrincipalID: identity.PrincipalID("user-1"),
		From:        actionplan.StateProposed,
		To:          actionplan.StateApproved,
		At:          record.CreatedAt.Add(time.Second),
	}
	if _, found, err := store.UpdateActionPlanState(ctx, tenant.ID("tenant-other"), record.InventoryID, record.ID, transition); err != nil || found {
		t.Fatalf("expected wrong tenant transition to miss, found=%t err=%v", found, err)
	}
	if _, _, err := store.UpdateActionPlanState(ctx, record.TenantID, record.InventoryID, record.ID, ports.ActionPlanStateTransition{PrincipalID: identity.PrincipalID("user-2"), From: actionplan.StateProposed, To: actionplan.StateApproved, At: transition.At}); err == nil {
		t.Fatalf("expected wrong principal transition to fail")
	}

	approved, found, err := store.UpdateActionPlanState(ctx, record.TenantID, record.InventoryID, record.ID, transition)
	if err != nil {
		t.Fatalf("approve action plan: %v", err)
	}
	if !found || approved.State != actionplan.StateApproved || approved.ApprovedAt.IsZero() {
		t.Fatalf("unexpected approved plan found=%t record=%+v", found, approved)
	}
	if _, _, err := store.UpdateActionPlanState(ctx, record.TenantID, record.InventoryID, record.ID, ports.ActionPlanStateTransition{PrincipalID: identity.PrincipalID("user-1"), From: actionplan.StateProposed, To: actionplan.StateCancelled, At: record.CreatedAt.Add(2 * time.Second)}); err == nil {
		t.Fatalf("expected stale proposed transition to fail")
	}
	executed, found, err := store.UpdateActionPlanState(ctx, record.TenantID, record.InventoryID, record.ID, ports.ActionPlanStateTransition{PrincipalID: identity.PrincipalID("user-1"), From: actionplan.StateApproved, To: actionplan.StateExecuted, At: record.CreatedAt.Add(3 * time.Second)})
	if err != nil {
		t.Fatalf("execute approved action plan: %v", err)
	}
	if !found || executed.State != actionplan.StateExecuted || executed.ExecutedAt.IsZero() {
		t.Fatalf("unexpected executed plan found=%t record=%+v", found, executed)
	}
	if _, _, err := store.UpdateActionPlanState(ctx, record.TenantID, record.InventoryID, record.ID, ports.ActionPlanStateTransition{PrincipalID: identity.PrincipalID("user-1"), From: actionplan.StateExecuted, To: actionplan.StateFailed, At: record.CreatedAt.Add(4 * time.Second)}); err == nil {
		t.Fatalf("expected terminal transition to fail")
	}
}

func TestActionPlanRepositoryExecutesUpdateAssetAtomically(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewStore()
	tenantID := tenant.ID("tenant-home")
	inventoryID := inventory.InventoryID("inventory-home")
	saveMemoryTenant(t, ctx, store, tenantID)
	saveMemoryInventory(t, ctx, store, tenantID, inventoryID)
	record := memoryActionPlanRecord("plan-move", time.Date(2026, 6, 26, 18, 0, 0, 0, time.UTC))
	if err := store.SaveActionPlan(ctx, record); err != nil {
		t.Fatalf("save action plan: %v", err)
	}
	if _, _, err := store.UpdateActionPlanState(ctx, record.TenantID, record.InventoryID, record.ID, ports.ActionPlanStateTransition{PrincipalID: record.PrincipalID, From: actionplan.StateProposed, To: actionplan.StateApproved, At: record.CreatedAt.Add(time.Second)}); err != nil {
		t.Fatalf("approve action plan: %v", err)
	}
	item := memoryAsset(t, "asset-one", tenantID, inventoryID)
	location := memoryAsset(t, "location-one", tenantID, inventoryID)
	location.Kind = asset.KindLocation
	if err := store.CreateAsset(ctx, item, memoryAuditRecord(t, "audit-create-asset", tenantID), nil); err != nil {
		t.Fatalf("create asset: %v", err)
	}
	if err := store.CreateAsset(ctx, location, memoryAuditRecord(t, "audit-create-location", tenantID), nil); err != nil {
		t.Fatalf("create location: %v", err)
	}
	previous := item
	item.ParentAssetID = location.ID
	moveAudit := memoryAuditRecord(t, "audit-move", tenantID)
	moveAudit.Action = audit.ActionAssetMoved
	executed, found, err := store.ExecuteUpdateAssetActionPlan(ctx, record.TenantID, record.InventoryID, record.ID, ports.ActionPlanStateTransition{PrincipalID: record.PrincipalID, From: actionplan.StateApproved, To: actionplan.StateExecuted, At: record.CreatedAt.Add(2 * time.Second)}, previous, item, []audit.Record{moveAudit}, nil)
	if err != nil {
		t.Fatalf("execute update asset action plan: %v", err)
	}
	if !found || executed.State != actionplan.StateExecuted || executed.ExecutedAt.IsZero() {
		t.Fatalf("unexpected executed plan found=%t record=%+v", found, executed)
	}
	moved, found, err := store.AssetByID(ctx, tenantID, inventoryID, item.ID)
	if err != nil || !found || moved.ParentAssetID != location.ID {
		t.Fatalf("expected moved asset found=%t item=%+v err=%v", found, moved, err)
	}

	staleRecord := memoryActionPlanRecord("plan-stale-move", record.CreatedAt.Add(time.Hour))
	if err := store.SaveActionPlan(ctx, staleRecord); err != nil {
		t.Fatalf("save stale action plan: %v", err)
	}
	if _, _, err := store.UpdateActionPlanState(ctx, staleRecord.TenantID, staleRecord.InventoryID, staleRecord.ID, ports.ActionPlanStateTransition{PrincipalID: staleRecord.PrincipalID, From: actionplan.StateProposed, To: actionplan.StateApproved, At: staleRecord.CreatedAt.Add(time.Second)}); err != nil {
		t.Fatalf("approve stale action plan: %v", err)
	}
	bumped := moved
	bumped.UpdatedAt = moved.UpdatedAt.Add(time.Minute)
	if err := store.UpdateAsset(ctx, bumped, nil, nil); err != nil {
		t.Fatalf("bump asset timestamp: %v", err)
	}
	target := bumped
	target.ParentAssetID = asset.ID("")
	if _, _, err := store.ExecuteUpdateAssetActionPlan(ctx, staleRecord.TenantID, staleRecord.InventoryID, staleRecord.ID, ports.ActionPlanStateTransition{PrincipalID: staleRecord.PrincipalID, From: actionplan.StateApproved, To: actionplan.StateExecuted, At: staleRecord.CreatedAt.Add(2 * time.Second)}, moved, target, []audit.Record{moveAudit}, nil); err == nil {
		t.Fatalf("expected stale timestamp execution to fail")
	}
}

func memoryActionPlanRecord(id string, createdAt time.Time) ports.ActionPlanRecord {
	return ports.ActionPlanRecord{
		ID:                         id,
		TenantID:                   tenant.ID("tenant-home"),
		InventoryID:                inventory.InventoryID("inventory-home"),
		PrincipalID:                identity.PrincipalID("user-1"),
		Source:                     "mobile_voice",
		RealtimeSessionID:          "session-1",
		State:                      actionplan.StateProposed,
		IntentSummary:              "Create a water bottle asset",
		ModelInterpretationSummary: "The user wants to add an item.",
		ConfirmationSummary:        "Create item water bottle?",
		Commands: []ports.ActionPlanCommandRecord{{
			ID:            "command-1",
			Kind:          actionplan.CommandKindCreateAsset,
			Summary:       "Create item water bottle",
			ArgumentsJSON: []byte(`{"name":"water bottle"}`),
		}},
		Risks:     []string{"Creates a new inventory item."},
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}
}
