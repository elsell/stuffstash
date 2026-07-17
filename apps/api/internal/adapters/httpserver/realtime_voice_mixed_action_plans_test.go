package httpserver

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
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
		"voice-session-id", "read-subject-id", "read-destination-id", "plan-id",
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
	assertMixedMoveCommandResults(t, executed)
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

func TestRealtimeVoiceActionPlanApprovalReturnsAttachmentUploadIntentForDependentCreate(t *testing.T) {
	t.Parallel()

	ctx, connection, sessionID, planID, proposedPlan := openRealtimeVoiceReviewSessionWithSetupAndIDsAndTranscriptAndProposal(t, createNestedItemActionPlanProposalLanguageModel{}, []string{
		"voice-session-id", "read-subject-id", "read-room-id", "read-closet-id", "plan-id",
		"room-id", "room-undo-id", "room-audit-id",
		"closet-id", "closet-undo-id", "closet-audit-id",
		"item-id", "item-undo-id", "item-audit-id",
		"upload-id", "attachment-id",
	}, "Add diaper genie refills to the closet in Henry's room.", nil)
	assertNestedCreateProposal(t, proposedPlan)

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":      "action.plan.approve",
		"seq":       4,
		"sessionId": sessionID,
		"planId":    planID,
		"photoAttachments": []map[string]any{{
			"commandId":   "create-subject",
			"photoIndex":  float64(0),
			"fileName":    "diaper-genie-refills.jpg",
			"contentType": "image/jpeg",
			"sizeBytes":   float64(1841050),
		}},
	})
	approved := readRealtimeMessage(t, ctx, connection)
	if approved["type"] != "action.plan.approved" || approved["planId"] != planID || approved["status"] != "approved" {
		t.Fatalf("expected safe approval event, got %+v", approved)
	}
	executed := readRealtimeMessage(t, ctx, connection)
	if executed["type"] != "action.plan.executed" || executed["planId"] != planID || executed["status"] != "executed" {
		t.Fatalf("expected safe execution event, got %+v", executed)
	}
	assertNestedCreateUploadIntent(t, executed)
	assertSafeRealtimeEvents(t, []map[string]any{approved, executed}, []string{"apiKey", "Bearer", "provider_session_id", "file://", "base64", "raw-photo"})
}

func TestRealtimeVoiceActionPlanApprovalRejectsInvalidPhotoMetadataBeforeApproval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		photo map[string]any
	}{
		{
			name: "unsupported mime",
			photo: map[string]any{
				"commandId":   "create-subject",
				"photoIndex":  float64(0),
				"fileName":    "manual.pdf",
				"contentType": "application/pdf",
				"sizeBytes":   float64(1024),
			},
		},
		{
			name: "unknown command",
			photo: map[string]any{
				"commandId":   "cmd-missing",
				"photoIndex":  float64(0),
				"fileName":    "refills.jpg",
				"contentType": "image/jpeg",
				"sizeBytes":   float64(1024),
			},
		},
		{
			name: "oversized photo",
			photo: map[string]any{
				"commandId":   "create-subject",
				"photoIndex":  float64(0),
				"fileName":    "refills.jpg",
				"contentType": "image/jpeg",
				"sizeBytes":   float64(26 * 1024 * 1024),
			},
		},
		{
			name: "local uri and bytes are not metadata",
			photo: map[string]any{
				"commandId":     "create-subject",
				"photoIndex":    float64(0),
				"fileName":      "refills.jpg",
				"contentType":   "image/jpeg",
				"sizeBytes":     float64(1024),
				"localUri":      "file:///private/mobile/photo.jpg",
				"contentBase64": "cGhvdG8=",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, connection, sessionID, planID, _ := openRealtimeVoiceReviewSessionWithSetupAndIDsAndTranscriptAndProposal(t, createNestedItemActionPlanProposalLanguageModel{}, []string{
				"voice-session-id",
				"plan-nested-item",
			}, "Add diaper genie refills to the closet in Henry's room.", nil)

			writeRealtimeMessage(t, ctx, connection, map[string]any{
				"type":             "action.plan.approve",
				"seq":              4,
				"sessionId":        sessionID,
				"planId":           planID,
				"photoAttachments": []map[string]any{tt.photo},
			})
			events := readRealtimeMessagesUntil(t, ctx, connection, "session.failed")
			if hasRealtimeEvent(events, "action.plan.approved") || hasRealtimeEvent(events, "action.plan.executed") {
				t.Fatalf("expected invalid photo metadata to fail before approval, got %+v", events)
			}
			failed := findRealtimeEvent(t, events, "session.failed")
			if failed["code"] != "invalid_request" {
				t.Fatalf("expected invalid request failure, got %+v", failed)
			}
		})
	}
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
	if !ok || createCommand["id"] != "create-destination-0" || createCommand["kind"] != "create_location" || createCommand["operation"] != "create" || createCommand["title"] != "Kitchen" || createCommand["assetKind"] != "location" {
		t.Fatalf("unexpected create command proposal: %+v", commands[0])
	}
	moveCommand, ok := commands[1].(map[string]any)
	if !ok || moveCommand["id"] != "move-subject" || moveCommand["kind"] != "move_asset" || moveCommand["operation"] != "move" || moveCommand["assetKind"] != "item" || moveCommand["parentCommandId"] != "create-destination-0" {
		t.Fatalf("unexpected move command proposal: %+v", commands[1])
	}
	if _, leaked := moveCommand["parentAssetId"]; leaked {
		t.Fatalf("expected move proposal to target command dependency instead of raw parent asset id: %+v", moveCommand)
	}
	assertSafeRealtimeEvents(t, []map[string]any{actionPlan}, []string{"apiKey", "Bearer", "provider_session_id"})
}

func assertNestedCreateProposal(t *testing.T, actionPlan map[string]any) {
	t.Helper()

	commands, ok := actionPlan["commands"].([]any)
	if !ok || len(commands) != 3 {
		t.Fatalf("expected nested create commands, got %+v", actionPlan["commands"])
	}
	itemCommand, ok := commands[2].(map[string]any)
	if !ok || itemCommand["id"] != "create-subject" || itemCommand["operation"] != "create" || itemCommand["assetKind"] != "item" || itemCommand["parentCommandId"] != "create-destination-1" {
		t.Fatalf("unexpected item command: %+v", commands[2])
	}
}

func assertNestedCreateUploadIntent(t *testing.T, executed map[string]any) {
	t.Helper()

	commandResults, ok := executed["commandResults"].([]any)
	if !ok || len(commandResults) != 3 {
		t.Fatalf("expected three command results, got %+v", executed["commandResults"])
	}
	intents, ok := executed["attachmentUploadIntents"].([]any)
	if !ok || len(intents) != 1 {
		t.Fatalf("expected one upload intent, got %+v", executed["attachmentUploadIntents"])
	}
	intent, ok := intents[0].(map[string]any)
	itemResult := map[string]any{}
	for _, raw := range commandResults {
		result, ok := raw.(map[string]any)
		if ok && result["commandId"] == "create-subject" {
			itemResult = result
		}
	}
	if itemResult["assetId"] == "" {
		t.Fatalf("expected item command result, got %+v", commandResults)
	}
	if !ok || intent["commandId"] != "create-subject" || intent["photoIndex"] != float64(0) || intent["assetId"] != itemResult["assetId"] || intent["fileName"] != "diaper-genie-refills.jpg" || intent["contentType"] != "image/jpeg" {
		t.Fatalf("unexpected upload intent: %+v", intent)
	}
	directUpload, ok := intent["directUpload"].(map[string]any)
	if !ok || directUpload["uploadId"] == "" || directUpload["attachmentId"] == "" || directUpload["method"] != "PUT" {
		t.Fatalf("unexpected direct upload intent: %+v", intent)
	}
	if _, leakedTitle := intent["title"]; leakedTitle {
		t.Fatalf("upload intent must not include asset title: %+v", intent)
	}
}

func assertMixedMoveCommandResults(t *testing.T, executed map[string]any) {
	t.Helper()

	commandResults, ok := executed["commandResults"].([]any)
	if !ok || len(commandResults) != 2 {
		t.Fatalf("expected create and move command results, got %+v", executed["commandResults"])
	}
	resultsByCommandID := map[string]map[string]any{}
	for _, raw := range commandResults {
		result, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("unexpected command result shape: %+v", raw)
		}
		commandID, ok := result["commandId"].(string)
		if !ok || commandID == "" {
			t.Fatalf("expected safe command result id, got %+v", result)
		}
		if result["title"] != nil {
			t.Fatalf("expected command result to omit titles, got %+v", result)
		}
		resultsByCommandID[commandID] = result
	}
	if result := resultsByCommandID["create-destination-0"]; result["assetId"] == "" || result["operation"] != "create" || result["assetKind"] != "location" {
		t.Fatalf("unexpected kitchen command result: %+v", result)
	}
	if result := resultsByCommandID["move-subject"]; result["assetId"] == "" || result["operation"] != "move" || result["assetKind"] != "item" {
		t.Fatalf("unexpected move command result: %+v", result)
	}
	assertSafeRealtimeEvents(t, []map[string]any{executed}, []string{"Water bottle", "Kitchen", "apiKey", "Bearer", "provider_session_id"})
}

type createNestedItemActionPlanProposalLanguageModel struct{}

func (m createNestedItemActionPlanProposalLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	intent := agentmodel.Intent{
		RequestShape: agentmodel.RequestShapeSingleTarget,
		Kind:         agentmodel.IntentKindChange, Operation: agentmodel.OperationCreate, SubjectMention: "diaper genie refills", NewAssetKind: "item",
		DestinationPath:  []string{"Henry's room", "closet"},
		DestinationKinds: []agentmodel.DestinationKind{agentmodel.DestinationKindLocation, agentmodel.DestinationKindContainer},
	}
	return typedVoiceInvestigationTurn(input, intent, nil)
}

type moveToMissingLocationActionPlanProposalLanguageModel struct{}

func (m moveToMissingLocationActionPlanProposalLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	intent := agentmodel.Intent{
		RequestShape: agentmodel.RequestShapeSingleTarget,
		Kind:         agentmodel.IntentKindChange, Operation: agentmodel.OperationMove, SubjectMention: "water bottle",
		DestinationPath: []string{"Kitchen"}, DestinationKinds: []agentmodel.DestinationKind{agentmodel.DestinationKindLocation},
	}
	return typedVoiceInvestigationTurn(input, intent, nil)
}
