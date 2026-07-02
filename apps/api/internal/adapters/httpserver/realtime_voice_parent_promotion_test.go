package httpserver

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceActionPlanApprovalPromotesItemParentForCreate(t *testing.T) {
	t.Parallel()

	var application app.App
	ctx, connection, sessionID, planID := openRealtimeVoiceReviewSessionWithSetupAndIDsAndTranscript(t, createUnderItemActionPlanProposalLanguageModel{}, []string{
		"baby-id", "baby-undo-id", "baby-audit-id",
		"voice-session-id", "read-tool-id", "plan-id", "command-id",
		"milk-id", "audit-baby-promotion", "milk-undo-id", "milk-audit-id",
	}, "Add milk to Baby.", func(seedApplication app.App) {
		application = seedApplication
		seedVoiceAsset(t, seedApplication, "user-1", "tenant-home", "inventory-home", "item", "Baby", "")
	})

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":      "action.plan.approve",
		"seq":       4,
		"sessionId": sessionID,
		"planId":    planID,
	})
	approved := readRealtimeMessage(t, ctx, connection)
	if approved["type"] != "action.plan.approved" || approved["planId"] != planID || approved["status"] != "approved" {
		t.Fatalf("expected safe approval event, got %+v", approved)
	}
	executed := readRealtimeMessage(t, ctx, connection)
	if executed["type"] != "action.plan.executed" || executed["planId"] != planID || executed["status"] != "executed" {
		t.Fatalf("expected safe create execution event, got %+v", executed)
	}
	commandResults, ok := executed["commandResults"].([]any)
	if !ok || len(commandResults) != 1 {
		t.Fatalf("expected one create command result, got %+v", executed["commandResults"])
	}
	commandResult, ok := commandResults[0].(map[string]any)
	if !ok || commandResult["commandId"] != "command-id" || commandResult["assetId"] != "milk-id" || commandResult["operation"] != "create" || commandResult["assetKind"] != "item" || commandResult["title"] != nil {
		t.Fatalf("unexpected create command result: %+v", commandResult)
	}

	parent, err := application.GetAsset(context.Background(), app.GetAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		Source:      audit.SourceAPI,
		RequestID:   "assert-promoted-parent",
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		AssetID:     asset.ID("baby-id"),
	})
	if err != nil {
		t.Fatalf("read promoted parent: %v", err)
	}
	if parent.Kind != asset.KindContainer {
		t.Fatalf("expected realtime-approved create to promote item parent, got %+v", parent)
	}
	child, err := application.GetAsset(context.Background(), app.GetAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		Source:      audit.SourceAPI,
		RequestID:   "assert-created-child",
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		AssetID:     asset.ID("milk-id"),
	})
	if err != nil {
		t.Fatalf("read created child: %v", err)
	}
	if child.ParentAssetID != parent.ID {
		t.Fatalf("expected child under promoted parent, got %+v", child)
	}
	assertSafeRealtimeEvents(t, []map[string]any{approved, executed}, []string{"Baby", "Milk", "apiKey", "Bearer", "provider_session_id"})
}

type createUnderItemActionPlanProposalLanguageModel struct{}

func (m createUnderItemActionPlanProposalLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	return ports.LanguageInferenceTurn{
		ToolCalls: []ports.AgentToolCall{{
			ID:   "plan-tool-call",
			Name: "propose_action_plan",
			Arguments: map[string]any{
				"commandKind":                "create_asset",
				"intentSummary":              "Create milk under Baby.",
				"modelInterpretationSummary": "The user wants to add a milk item inside the existing Baby item.",
				"confirmationSummary":        "Create milk in Baby?",
				"commandSummary":             "Create milk in Baby",
				"arguments": map[string]any{
					"kind":          "item",
					"parentAssetId": "baby-id",
					"title":         "Milk",
				},
				"riskSummary": "Adds a new item to this inventory.",
			},
		}},
	}, nil
}
