package app

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestApproveActionPlanReturnsNotFoundForMissingPlan(t *testing.T) {
	t.Parallel()

	application := newActionPlanTestApp(&fakeActionPlanRepository{records: map[string]ports.ActionPlanRecord{}}, &fakeIDGenerator{}, nil)
	_, err := application.ApproveActionPlan(context.Background(), ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "missing-plan",
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestApproveActionPlanAtomicallyAppliesReviewedCreateEdits(t *testing.T) {
	t.Parallel()
	repository := &fakeActionPlanRepository{records: map[string]ports.ActionPlanRecord{
		"plan-1": actionPlanRecordWithCommand("plan-1", actionplan.StateProposed, actionplan.CommandKindCreateAsset, `{"name":"water bottle","parentAssetId":"asset-kitchen"}`),
	}}
	application := newActionPlanTestApp(repository, &fakeIDGenerator{}, nil)
	title := "Holiday hand towels"
	approved, err := application.ApproveActionPlan(context.Background(), ActionPlanDecisionInput{
		Principal: identity.Principal{ID: identity.PrincipalID("user-1")}, TenantID: tenant.ID("tenant-home"), InventoryID: inventory.InventoryID("inventory-home"), PlanID: "plan-1",
		CommandEdits: []ActionPlanCommandEditInput{{CommandID: "command-1", Title: &title, ParentSelection: &ActionPlanParentSelectionInput{Kind: "root"}}},
	})
	if err != nil {
		t.Fatalf("approve edited plan: %v", err)
	}
	var arguments map[string]any
	if err := json.Unmarshal(approved.Commands[0].ArgumentsJSON, &arguments); err != nil {
		t.Fatalf("decode edited arguments: %v", err)
	}
	if approved.State != actionplan.StateApproved || arguments["title"] != title || arguments["name"] != nil || arguments["parentAssetId"] != nil {
		t.Fatalf("unexpected edited approval: state=%s arguments=%v", approved.State, arguments)
	}
}

func TestApproveActionPlanRejectsEditsOutsideCreateCommandsWithoutChangingPlan(t *testing.T) {
	t.Parallel()
	record := actionPlanRecordWithCommand("plan-1", actionplan.StateProposed, actionplan.CommandKindArchiveAsset, `{"assetId":"asset-1"}`)
	repository := &fakeActionPlanRepository{records: map[string]ports.ActionPlanRecord{"plan-1": record}}
	application := newActionPlanTestApp(repository, &fakeIDGenerator{}, nil)
	title := "Forged title"
	_, err := application.ApproveActionPlan(context.Background(), ActionPlanDecisionInput{
		Principal: identity.Principal{ID: identity.PrincipalID("user-1")}, TenantID: tenant.ID("tenant-home"), InventoryID: inventory.InventoryID("inventory-home"), PlanID: "plan-1",
		CommandEdits: []ActionPlanCommandEditInput{{CommandID: "command-1", Title: &title}},
	})
	if !errors.Is(err, ErrValidation) || repository.records["plan-1"].State != actionplan.StateProposed {
		t.Fatalf("expected unchanged proposed plan after invalid edit, got err=%v record=%+v", err, repository.records["plan-1"])
	}
}
