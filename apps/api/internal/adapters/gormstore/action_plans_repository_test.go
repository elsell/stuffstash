package gormstore

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

func TestActionPlanRepositorySavesSafeStructuredPlan(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	saveInventory(t, ctx, store, "inventory-home", tenant.ID("tenant-home"), "Home")
	record := gormActionPlanRecord("plan-1", time.Date(2026, 6, 26, 18, 0, 0, 0, time.UTC))

	if err := store.SaveActionPlan(ctx, record); err != nil {
		t.Fatalf("save action plan: %v", err)
	}
	got, found, err := store.ActionPlanByID(ctx, record.TenantID, record.InventoryID, record.ID)
	if err != nil {
		t.Fatalf("read action plan: %v", err)
	}
	if !found {
		t.Fatalf("expected action plan to be found")
	}
	if got.State != actionplan.StateProposed || got.ConfirmationSummary != record.ConfirmationSummary || len(got.Commands) != 1 {
		t.Fatalf("unexpected persisted action plan: %+v", got)
	}
	if string(got.Commands[0].ArgumentsJSON) != `{"name":"water bottle"}` {
		t.Fatalf("unexpected command arguments: %s", got.Commands[0].ArgumentsJSON)
	}
}

func TestActionPlanRepositoryScopesAndFreezesTransitions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	saveTenant(t, ctx, store, tenant.ID("tenant-other"), "Other")
	saveInventory(t, ctx, store, "inventory-home", tenant.ID("tenant-home"), "Home")
	record := gormActionPlanRecord("plan-1", time.Date(2026, 6, 26, 18, 0, 0, 0, time.UTC))
	if err := store.SaveActionPlan(ctx, record); err != nil {
		t.Fatalf("save action plan: %v", err)
	}

	if _, found, err := store.ActionPlanByID(ctx, tenant.ID("tenant-other"), record.InventoryID, record.ID); err != nil || found {
		t.Fatalf("expected wrong tenant read to miss, found=%t err=%v", found, err)
	}
	transition := ports.ActionPlanStateTransition{
		PrincipalID: identity.PrincipalID("user-1"),
		From:        actionplan.StateProposed,
		To:          actionplan.StateCancelled,
		At:          record.CreatedAt.Add(time.Second),
	}
	if _, found, err := store.UpdateActionPlanState(ctx, tenant.ID("tenant-other"), record.InventoryID, record.ID, transition); err != nil || found {
		t.Fatalf("expected wrong tenant transition to miss, found=%t err=%v", found, err)
	}
	if _, _, err := store.UpdateActionPlanState(ctx, record.TenantID, record.InventoryID, record.ID, ports.ActionPlanStateTransition{PrincipalID: identity.PrincipalID("user-2"), From: actionplan.StateProposed, To: actionplan.StateCancelled, At: transition.At}); err == nil {
		t.Fatalf("expected wrong principal transition to fail")
	}

	cancelled, found, err := store.UpdateActionPlanState(ctx, record.TenantID, record.InventoryID, record.ID, transition)
	if err != nil {
		t.Fatalf("cancel action plan: %v", err)
	}
	if !found || cancelled.State != actionplan.StateCancelled || cancelled.CancelledAt.IsZero() {
		t.Fatalf("unexpected cancelled plan found=%t record=%+v", found, cancelled)
	}
	if _, _, err := store.UpdateActionPlanState(ctx, record.TenantID, record.InventoryID, record.ID, ports.ActionPlanStateTransition{PrincipalID: identity.PrincipalID("user-1"), From: actionplan.StateProposed, To: actionplan.StateApproved, At: record.CreatedAt.Add(2 * time.Second)}); err == nil {
		t.Fatalf("expected terminal transition to fail")
	}
}

func TestActionPlanRepositoryAllowsApprovedPlanToExecuteOrFail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	saveInventory(t, ctx, store, "inventory-home", tenant.ID("tenant-home"), "Home")
	executedRecord := gormActionPlanRecord("plan-executed", time.Date(2026, 6, 26, 18, 0, 0, 0, time.UTC))
	failedRecord := gormActionPlanRecord("plan-failed", time.Date(2026, 6, 26, 18, 1, 0, 0, time.UTC))
	if err := store.SaveActionPlan(ctx, executedRecord); err != nil {
		t.Fatalf("save executable action plan: %v", err)
	}
	if err := store.SaveActionPlan(ctx, failedRecord); err != nil {
		t.Fatalf("save failable action plan: %v", err)
	}
	approve := func(record ports.ActionPlanRecord) {
		t.Helper()
		if _, _, err := store.UpdateActionPlanState(ctx, record.TenantID, record.InventoryID, record.ID, ports.ActionPlanStateTransition{PrincipalID: record.PrincipalID, From: actionplan.StateProposed, To: actionplan.StateApproved, At: record.CreatedAt.Add(time.Second)}); err != nil {
			t.Fatalf("approve %s: %v", record.ID, err)
		}
	}
	approve(executedRecord)
	approve(failedRecord)

	executed, found, err := store.UpdateActionPlanState(ctx, executedRecord.TenantID, executedRecord.InventoryID, executedRecord.ID, ports.ActionPlanStateTransition{PrincipalID: executedRecord.PrincipalID, From: actionplan.StateApproved, To: actionplan.StateExecuted, At: executedRecord.CreatedAt.Add(2 * time.Second)})
	if err != nil {
		t.Fatalf("execute approved action plan: %v", err)
	}
	if !found || executed.State != actionplan.StateExecuted || executed.ExecutedAt.IsZero() {
		t.Fatalf("unexpected executed plan found=%t record=%+v", found, executed)
	}

	failed, found, err := store.UpdateActionPlanState(ctx, failedRecord.TenantID, failedRecord.InventoryID, failedRecord.ID, ports.ActionPlanStateTransition{PrincipalID: failedRecord.PrincipalID, From: actionplan.StateApproved, To: actionplan.StateFailed, At: failedRecord.CreatedAt.Add(2 * time.Second)})
	if err != nil {
		t.Fatalf("fail approved action plan: %v", err)
	}
	if !found || failed.State != actionplan.StateFailed || failed.FailedAt.IsZero() {
		t.Fatalf("unexpected failed plan found=%t record=%+v", found, failed)
	}
}

func TestActionPlanRepositoryExecutesCreateAssetAtomically(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	saveInventory(t, ctx, store, "inventory-home", tenant.ID("tenant-home"), "Home")
	successRecord := gormActionPlanRecord("plan-execute-create", time.Date(2026, 6, 26, 18, 0, 0, 0, time.UTC))
	rollbackRecord := gormActionPlanRecord("plan-rollback-create", time.Date(2026, 6, 26, 18, 1, 0, 0, time.UTC))
	if err := store.SaveActionPlan(ctx, successRecord); err != nil {
		t.Fatalf("save successful execution plan: %v", err)
	}
	if err := store.SaveActionPlan(ctx, rollbackRecord); err != nil {
		t.Fatalf("save rollback execution plan: %v", err)
	}
	approveActionPlanForGormTest(t, ctx, store, successRecord)
	approveActionPlanForGormTest(t, ctx, store, rollbackRecord)

	item := assetItem("asset-from-plan", successRecord.TenantID.String(), successRecord.InventoryID.String(), asset.KindItem, "")
	executed, found, err := store.ExecuteCreateAssetActionPlan(ctx, successRecord.TenantID, successRecord.InventoryID, successRecord.ID, ports.ActionPlanStateTransition{
		PrincipalID: successRecord.PrincipalID,
		From:        actionplan.StateApproved,
		To:          actionplan.StateExecuted,
		At:          successRecord.CreatedAt.Add(2 * time.Second),
	}, item, auditRecord(t, "audit-plan-create", successRecord.TenantID, successRecord.InventoryID, audit.ActionAssetCreated), nil)
	if err != nil {
		t.Fatalf("execute create asset action plan: %v", err)
	}
	if !found || executed.State != actionplan.StateExecuted || executed.ExecutedAt.IsZero() {
		t.Fatalf("unexpected executed plan found=%t record=%+v", found, executed)
	}
	if _, found, err := store.AssetByID(ctx, successRecord.TenantID, successRecord.InventoryID, item.ID); err != nil || !found {
		t.Fatalf("expected atomically created asset found=%t err=%v", found, err)
	}

	duplicate := assetItem("asset-duplicate", rollbackRecord.TenantID.String(), rollbackRecord.InventoryID.String(), asset.KindItem, "")
	if err := createAsset(t, ctx, store, duplicate); err != nil {
		t.Fatalf("seed duplicate asset: %v", err)
	}
	if _, _, err := store.ExecuteCreateAssetActionPlan(ctx, rollbackRecord.TenantID, rollbackRecord.InventoryID, rollbackRecord.ID, ports.ActionPlanStateTransition{
		PrincipalID: rollbackRecord.PrincipalID,
		From:        actionplan.StateApproved,
		To:          actionplan.StateExecuted,
		At:          rollbackRecord.CreatedAt.Add(2 * time.Second),
	}, duplicate, auditRecord(t, "audit-plan-rollback", rollbackRecord.TenantID, rollbackRecord.InventoryID, audit.ActionAssetCreated), nil); err == nil {
		t.Fatalf("expected duplicate asset execution to fail")
	}
	rolledBack, found, err := store.ActionPlanByID(ctx, rollbackRecord.TenantID, rollbackRecord.InventoryID, rollbackRecord.ID)
	if err != nil {
		t.Fatalf("read rolled back action plan: %v", err)
	}
	if !found || rolledBack.State != actionplan.StateApproved || !rolledBack.ExecutedAt.IsZero() {
		t.Fatalf("expected plan to remain approved after rolled back execution found=%t record=%+v", found, rolledBack)
	}
}

func TestActionPlanRepositoryExecutesUpdateAssetAtomically(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	saveInventory(t, ctx, store, "inventory-home", tenant.ID("tenant-home"), "Home")
	successRecord := gormActionPlanRecord("plan-execute-move", time.Date(2026, 6, 26, 18, 0, 0, 0, time.UTC))
	rollbackRecord := gormActionPlanRecord("plan-rollback-move", time.Date(2026, 6, 26, 18, 1, 0, 0, time.UTC))
	if err := store.SaveActionPlan(ctx, successRecord); err != nil {
		t.Fatalf("save successful execution plan: %v", err)
	}
	if err := store.SaveActionPlan(ctx, rollbackRecord); err != nil {
		t.Fatalf("save rollback execution plan: %v", err)
	}
	approveActionPlanForGormTest(t, ctx, store, successRecord)
	approveActionPlanForGormTest(t, ctx, store, rollbackRecord)

	location := assetItem("location-one", successRecord.TenantID.String(), successRecord.InventoryID.String(), asset.KindLocation, "")
	item := assetItem("asset-one", successRecord.TenantID.String(), successRecord.InventoryID.String(), asset.KindItem, "")
	location.CreatedAt = successRecord.CreatedAt
	location.UpdatedAt = successRecord.CreatedAt
	item.CreatedAt = successRecord.CreatedAt
	item.UpdatedAt = successRecord.CreatedAt
	if err := createAsset(t, ctx, store, location); err != nil {
		t.Fatalf("seed location: %v", err)
	}
	if err := createAsset(t, ctx, store, item); err != nil {
		t.Fatalf("seed asset: %v", err)
	}
	previous := item
	item.ParentAssetID = location.ID
	executed, found, err := store.ExecuteUpdateAssetActionPlan(ctx, successRecord.TenantID, successRecord.InventoryID, successRecord.ID, ports.ActionPlanStateTransition{
		PrincipalID: successRecord.PrincipalID,
		From:        actionplan.StateApproved,
		To:          actionplan.StateExecuted,
		At:          successRecord.CreatedAt.Add(2 * time.Second),
	}, previous, item, []audit.Record{auditRecord(t, "audit-plan-move", successRecord.TenantID, successRecord.InventoryID, audit.ActionAssetMoved)}, nil)
	if err != nil {
		t.Fatalf("execute update asset action plan: %v", err)
	}
	if !found || executed.State != actionplan.StateExecuted || executed.ExecutedAt.IsZero() {
		t.Fatalf("unexpected executed plan found=%t record=%+v", found, executed)
	}
	moved, found, err := store.AssetByID(ctx, successRecord.TenantID, successRecord.InventoryID, item.ID)
	if err != nil || !found || moved.ParentAssetID != location.ID {
		t.Fatalf("expected atomically moved asset found=%t item=%+v err=%v", found, moved, err)
	}

	duplicateAudit := auditRecord(t, "audit-plan-rollback", rollbackRecord.TenantID, rollbackRecord.InventoryID, audit.ActionAssetMoved)
	if err := store.SaveAuditRecord(ctx, duplicateAudit); err != nil {
		t.Fatalf("seed duplicate audit: %v", err)
	}
	item.ParentAssetID = asset.ID("")
	if _, _, err := store.ExecuteUpdateAssetActionPlan(ctx, rollbackRecord.TenantID, rollbackRecord.InventoryID, rollbackRecord.ID, ports.ActionPlanStateTransition{
		PrincipalID: rollbackRecord.PrincipalID,
		From:        actionplan.StateApproved,
		To:          actionplan.StateExecuted,
		At:          rollbackRecord.CreatedAt.Add(2 * time.Second),
	}, moved, item, []audit.Record{duplicateAudit}, nil); err == nil {
		t.Fatalf("expected duplicate audit execution to fail")
	}
	rolledBack, found, err := store.ActionPlanByID(ctx, rollbackRecord.TenantID, rollbackRecord.InventoryID, rollbackRecord.ID)
	if err != nil {
		t.Fatalf("read rolled back action plan: %v", err)
	}
	if !found || rolledBack.State != actionplan.StateApproved || !rolledBack.ExecutedAt.IsZero() {
		t.Fatalf("expected plan to remain approved after rolled back execution found=%t record=%+v", found, rolledBack)
	}
	stillMoved, found, err := store.AssetByID(ctx, successRecord.TenantID, successRecord.InventoryID, item.ID)
	if err != nil || !found || stillMoved.ParentAssetID != location.ID {
		t.Fatalf("expected asset move to roll back found=%t item=%+v err=%v", found, stillMoved, err)
	}

	staleRecord := gormActionPlanRecord("plan-stale-move", time.Date(2026, 6, 26, 18, 2, 0, 0, time.UTC))
	if err := store.SaveActionPlan(ctx, staleRecord); err != nil {
		t.Fatalf("save stale execution plan: %v", err)
	}
	approveActionPlanForGormTest(t, ctx, store, staleRecord)
	bumped := stillMoved
	bumped.UpdatedAt = stillMoved.UpdatedAt.Add(time.Minute)
	if err := updateAsset(t, ctx, store, bumped); err != nil {
		t.Fatalf("bump asset timestamp: %v", err)
	}
	staleTarget := bumped
	staleTarget.ParentAssetID = asset.ID("")
	if _, _, err := store.ExecuteUpdateAssetActionPlan(ctx, staleRecord.TenantID, staleRecord.InventoryID, staleRecord.ID, ports.ActionPlanStateTransition{
		PrincipalID: staleRecord.PrincipalID,
		From:        actionplan.StateApproved,
		To:          actionplan.StateExecuted,
		At:          staleRecord.CreatedAt.Add(2 * time.Second),
	}, stillMoved, staleTarget, []audit.Record{auditRecord(t, "audit-plan-stale", staleRecord.TenantID, staleRecord.InventoryID, audit.ActionAssetMoved)}, nil); err == nil {
		t.Fatalf("expected stale asset snapshot execution to fail")
	}
	stalePlan, found, err := store.ActionPlanByID(ctx, staleRecord.TenantID, staleRecord.InventoryID, staleRecord.ID)
	if err != nil {
		t.Fatalf("read stale action plan: %v", err)
	}
	if !found || stalePlan.State != actionplan.StateApproved || !stalePlan.ExecutedAt.IsZero() {
		t.Fatalf("expected stale plan to remain approved found=%t record=%+v", found, stalePlan)
	}
}

func TestActionPlanRepositoryStoresOnlySafeColumns(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	disallowed := []string{
		"audio",
		"audio_chunks",
		"transcript",
		"prompt",
		"provider_prompt",
		"provider_response",
		"model_response",
		"generated_speech",
		"credential",
		"bearer_token",
		"provider_session_id",
	}
	for _, column := range disallowed {
		if store.db.WithContext(ctx).Migrator().HasColumn(&actionPlanModel{}, column) {
			t.Fatalf("action plan model must not persist unsafe column %q", column)
		}
	}
}

func approveActionPlanForGormTest(t *testing.T, ctx context.Context, store Store, record ports.ActionPlanRecord) {
	t.Helper()

	if _, _, err := store.UpdateActionPlanState(ctx, record.TenantID, record.InventoryID, record.ID, ports.ActionPlanStateTransition{
		PrincipalID: record.PrincipalID,
		From:        actionplan.StateProposed,
		To:          actionplan.StateApproved,
		At:          record.CreatedAt.Add(time.Second),
	}); err != nil {
		t.Fatalf("approve %s: %v", record.ID, err)
	}
}

func gormActionPlanRecord(id string, createdAt time.Time) ports.ActionPlanRecord {
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
