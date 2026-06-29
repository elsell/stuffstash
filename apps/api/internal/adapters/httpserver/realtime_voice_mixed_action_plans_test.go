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

func TestRealtimeVoiceActionPlanApprovalCreatesMissingLocationThenMovesAsset(t *testing.T) {
	t.Parallel()

	var application app.App
	ctx, connection, sessionID, planID, proposedPlan := openRealtimeVoiceReviewSessionWithSetupAndIDsAndProposal(t, moveToMissingLocationActionPlanProposalLanguageModel{}, []string{
		"office-id", "office-undo-id", "office-audit-id",
		"asset-id", "asset-undo-id", "asset-audit-id",
		"voice-session-id", "plan-id",
		"kitchen-id", "kitchen-undo-id", "kitchen-audit-id", "move-undo-id", "move-audit-id",
	}, func(seedApplication app.App) {
		application = seedApplication
		seedVoiceAsset(t, seedApplication, "user-1", "tenant-home", "inventory-home", "location", "Office", "")
		seedVoiceAsset(t, seedApplication, "user-1", "tenant-home", "inventory-home", "item", "Water bottle", "office-id")
	})
	assertMixedMoveProposalForKitchen(t, proposedPlan)

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
		t.Fatalf("expected safe mixed-plan execution event, got %+v", executed)
	}
	assets, err := application.ListAssets(context.Background(), app.ListAssetsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		Source:      audit.SourceAPI,
		RequestID:   "assert-created-kitchen",
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		Limit:       20,
	})
	if err != nil {
		t.Fatalf("list assets after mixed plan: %v", err)
	}
	kitchen := asset.Asset{}
	for _, item := range assets.Items {
		if item.Title.String() == "Kitchen" {
			kitchen = item
		}
	}
	if kitchen.Kind != asset.KindLocation || kitchen.Title.String() != "Kitchen" {
		t.Fatalf("expected created Kitchen location in %+v", assets.Items)
	}
	moved, err := application.GetAsset(context.Background(), app.GetAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		Source:      audit.SourceAPI,
		RequestID:   "assert-moved-water-bottle",
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		AssetID:     asset.ID("asset-id"),
	})
	if err != nil {
		t.Fatalf("read moved water bottle: %v", err)
	}
	if moved.ParentAssetID != kitchen.ID {
		t.Fatalf("expected water bottle moved into Kitchen, got %+v", moved)
	}
	assertSafeRealtimeEvents(t, []map[string]any{approved, executed}, []string{"Water bottle", "Kitchen", "apiKey", "Bearer", "provider_session_id"})
}

func assertMixedMoveProposalForKitchen(t *testing.T, actionPlan map[string]any) {
	t.Helper()

	planID, _ := actionPlan["planId"].(string)
	if planID == "" || actionPlan["confirmationSummary"] != "Move Water bottle to Kitchen?" {
		t.Fatalf("unexpected mixed action plan summary: %+v", actionPlan)
	}
	commands, ok := actionPlan["commands"].([]any)
	if !ok || len(commands) != 2 {
		t.Fatalf("expected create and move commands, got %+v", actionPlan["commands"])
	}
	createCommand, ok := commands[0].(map[string]any)
	if !ok || createCommand["id"] != "cmd-kitchen" || createCommand["kind"] != "create_location" || createCommand["operation"] != "create" || createCommand["title"] != "Kitchen" || createCommand["assetKind"] != "location" {
		t.Fatalf("unexpected create command proposal: %+v", commands[0])
	}
	moveCommand, ok := commands[1].(map[string]any)
	if !ok || moveCommand["id"] != "cmd-move-water-bottle" || moveCommand["kind"] != "move_asset" || moveCommand["operation"] != "move" || moveCommand["parentCommandId"] != "cmd-kitchen" {
		t.Fatalf("unexpected move command proposal: %+v", commands[1])
	}
	if _, leaked := moveCommand["parentAssetId"]; leaked {
		t.Fatalf("expected move proposal to target command dependency instead of raw parent asset id: %+v", moveCommand)
	}
	assertSafeRealtimeEvents(t, []map[string]any{actionPlan}, []string{"apiKey", "Bearer", "provider_session_id"})
}

type moveToMissingLocationActionPlanProposalLanguageModel struct{}

func (m moveToMissingLocationActionPlanProposalLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	if len(input.ToolResults) == 0 {
		return ports.LanguageInferenceTurn{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "list-active-assets",
				Name: "list_authorized_assets",
				Arguments: map[string]any{
					"lifecycleState": "active",
					"limit":          float64(10),
				},
			}},
		}, nil
	}
	if len(input.ToolResults) == 1 {
		assetID, err := firstToolResultAssetID(input.ToolResults[0].Content, "Water bottle", "active")
		if err != nil {
			return ports.LanguageInferenceTurn{}, err
		}
		return ports.LanguageInferenceTurn{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "plan-tool-call",
				Name: "propose_action_plan",
				Arguments: map[string]any{
					"intentSummary":              "Move the water bottle to the kitchen.",
					"modelInterpretationSummary": "The user wants the existing visible water bottle moved into a new Kitchen location.",
					"confirmationSummary":        "Move Water bottle to Kitchen?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-kitchen",
							"kind":    "create_location",
							"summary": "Create Kitchen",
							"arguments": map[string]any{
								"title": "Kitchen",
								"kind":  "location",
							},
						},
						map[string]any{
							"id":      "cmd-move-water-bottle",
							"kind":    "move_asset",
							"summary": "Move Water bottle to Kitchen",
							"arguments": map[string]any{
								"assetId":         assetID,
								"parentCommandId": "cmd-kitchen",
							},
						},
					},
					"risks": []any{"Creates one new location and moves an item in this inventory."},
				},
			}},
		}, nil
	}
	return ports.LanguageInferenceTurn{
		Final: &ports.StructuredAgentResponse{
			Kind:            ports.StructuredAgentResponseKindClarification,
			SpokenResponse:  "I prepared that move for review.",
			DisplayResponse: "I prepared that move for review.",
		},
	}, nil
}
