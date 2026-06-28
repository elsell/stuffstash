package httpserver

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceQueryStreamsActionPlanProposalForReview(t *testing.T) {
	t.Parallel()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"voice-session-id", "plan-id", "command-id", "response-id"},
	}, fakeSpeechToText{transcript: "Add a water bottle."}, actionPlanProposalLanguageModel{}, fakeTextToSpeech{
		chunks: [][]byte{[]byte("spoken-audio")},
	})
	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	events := runRealtimeVoiceQuestion(t, server.URL, "tenant-home", "inventory-home", "user-1")
	proposed := findRealtimeEvent(t, events, "action.plan.proposed")
	actionPlan, ok := proposed["actionPlan"].(map[string]any)
	if !ok {
		t.Fatalf("expected action plan payload, got %+v", proposed)
	}
	if actionPlan["planId"] != "plan-id" || actionPlan["confirmationSummary"] != "Create item water bottle?" {
		t.Fatalf("unexpected action plan payload: %+v", actionPlan)
	}
	commands, ok := actionPlan["commands"].([]any)
	if !ok || len(commands) != 1 {
		t.Fatalf("expected one safe command, got %+v", actionPlan["commands"])
	}
	command, ok := commands[0].(map[string]any)
	if !ok || command["kind"] != "create_asset" || command["summary"] != "Create item water bottle" {
		t.Fatalf("unexpected command payload: %+v", commands[0])
	}
	assertSafeRealtimeEvents(t, events, []string{"fake-audio", "apiKey", "Bearer", "provider_session_id"})
}

func TestRealtimeVoiceActionPlanApprovalUsesOpenReviewSession(t *testing.T) {
	t.Parallel()

	ctx, connection, sessionID := openRealtimeVoiceReviewSession(t)

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":      "action.plan.approve",
		"seq":       4,
		"sessionId": sessionID,
		"planId":    "plan-id",
	})
	approved := readRealtimeMessage(t, ctx, connection)
	if approved["type"] != "action.plan.approved" || approved["planId"] != "plan-id" || approved["status"] != "approved" {
		t.Fatalf("expected safe approval event, got %+v", approved)
	}
	executed := readRealtimeMessage(t, ctx, connection)
	if executed["type"] != "action.plan.executed" || executed["planId"] != "plan-id" || executed["status"] != "executed" {
		t.Fatalf("expected safe execution event, got %+v", executed)
	}
	assertSafeRealtimeEvents(t, []map[string]any{approved, executed}, []string{"fake-audio", "apiKey", "Bearer", "provider_session_id"})
}

func TestRealtimeVoiceActionPlanApprovalExecutesMoveAsset(t *testing.T) {
	t.Parallel()

	var application app.App
	ctx, connection, sessionID := openRealtimeVoiceReviewSessionWithSetup(t, moveActionPlanProposalLanguageModel{}, func(seedApplication app.App) {
		application = seedApplication
		seedVoiceAsset(t, seedApplication, "user-1", "tenant-home", "inventory-home", "location", "Office", "")
		seedVoiceAsset(t, seedApplication, "user-1", "tenant-home", "inventory-home", "item", "Water bottle", "")
	})

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":      "action.plan.approve",
		"seq":       4,
		"sessionId": sessionID,
		"planId":    "plan-id",
	})
	approved := readRealtimeMessage(t, ctx, connection)
	if approved["type"] != "action.plan.approved" || approved["planId"] != "plan-id" || approved["status"] != "approved" {
		t.Fatalf("expected safe approval event, got %+v", approved)
	}
	executed := readRealtimeMessage(t, ctx, connection)
	if executed["type"] != "action.plan.executed" || executed["planId"] != "plan-id" || executed["status"] != "executed" {
		t.Fatalf("expected safe move execution event, got %+v", executed)
	}
	moved, err := application.GetAsset(context.Background(), app.GetAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		Source:      audit.SourceAPI,
		RequestID:   "assert-moved-asset",
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		AssetID:     asset.ID("asset-id"),
	})
	if err != nil {
		t.Fatalf("read moved asset: %v", err)
	}
	if moved.ParentAssetID != asset.ID("location-id") {
		t.Fatalf("expected realtime-approved move to update asset parent, got %+v", moved)
	}
	assertSafeRealtimeEvents(t, []map[string]any{approved, executed}, []string{"Water bottle", "Office", "apiKey", "Bearer", "provider_session_id"})
}

func TestRealtimeVoiceActionPlanApprovalExecutesArchiveAsset(t *testing.T) {
	t.Parallel()

	var application app.App
	ctx, connection, sessionID := openRealtimeVoiceReviewSessionWithSetup(t, archiveActionPlanProposalLanguageModel{}, func(seedApplication app.App) {
		application = seedApplication
		seedVoiceAsset(t, seedApplication, "user-1", "tenant-home", "inventory-home", "location", "Office", "")
		seedVoiceAsset(t, seedApplication, "user-1", "tenant-home", "inventory-home", "item", "Water bottle", "")
	})

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":      "action.plan.approve",
		"seq":       4,
		"sessionId": sessionID,
		"planId":    "plan-id",
	})
	approved := readRealtimeMessage(t, ctx, connection)
	if approved["type"] != "action.plan.approved" || approved["planId"] != "plan-id" || approved["status"] != "approved" {
		t.Fatalf("expected safe approval event, got %+v", approved)
	}
	executed := readRealtimeMessage(t, ctx, connection)
	if executed["type"] != "action.plan.executed" || executed["planId"] != "plan-id" || executed["status"] != "executed" {
		t.Fatalf("expected safe archive execution event, got %+v", executed)
	}
	archived, err := application.GetAsset(context.Background(), app.GetAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		Source:      audit.SourceAPI,
		RequestID:   "assert-archived-asset",
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		AssetID:     asset.ID("asset-id"),
	})
	if err != nil {
		t.Fatalf("read archived asset: %v", err)
	}
	if archived.LifecycleState != asset.LifecycleStateArchived {
		t.Fatalf("expected realtime-approved archive to update asset lifecycle, got %+v", archived)
	}
	assertSafeRealtimeEvents(t, []map[string]any{approved, executed}, []string{"Water bottle", "Office", "apiKey", "Bearer", "provider_session_id"})
}

func TestRealtimeVoiceActionPlanArchiveFailureLeavesAssetActive(t *testing.T) {
	t.Parallel()

	var application app.App
	ctx, connection, sessionID := openRealtimeVoiceReviewSessionWithSetupAndIDs(t, archiveActionPlanProposalLanguageModel{}, []string{
		"location-id", "location-undo-id", "location-audit-id",
		"asset-id", "asset-undo-id", "asset-audit-id",
		"child-id", "child-undo-id", "child-audit-id",
		"voice-session-id", "plan-id", "command-id", "response-id", "archive-undo-id", "archive-audit-id",
	}, func(seedApplication app.App) {
		application = seedApplication
		seedVoiceAsset(t, seedApplication, "user-1", "tenant-home", "inventory-home", "location", "Office", "")
		seedVoiceAsset(t, seedApplication, "user-1", "tenant-home", "inventory-home", "container", "Toolbox", "")
		seedVoiceAsset(t, seedApplication, "user-1", "tenant-home", "inventory-home", "item", "Wrench", "asset-id")
	})

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":      "action.plan.approve",
		"seq":       4,
		"sessionId": sessionID,
		"planId":    "plan-id",
	})
	approved := readRealtimeMessage(t, ctx, connection)
	if approved["type"] != "action.plan.approved" || approved["planId"] != "plan-id" || approved["status"] != "approved" {
		t.Fatalf("expected safe approval event, got %+v", approved)
	}
	failed := readRealtimeMessage(t, ctx, connection)
	if failed["type"] != "action.plan.failed" || failed["planId"] != "plan-id" || failed["status"] != "failed" {
		t.Fatalf("expected safe archive failure event, got %+v", failed)
	}
	item, err := application.GetAsset(context.Background(), app.GetAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		Source:      audit.SourceAPI,
		RequestID:   "assert-active-asset",
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		AssetID:     asset.ID("asset-id"),
	})
	if err != nil {
		t.Fatalf("read active asset: %v", err)
	}
	if item.LifecycleState != asset.LifecycleStateActive {
		t.Fatalf("expected failed realtime archive to leave asset active, got %+v", item)
	}
	assertSafeRealtimeEvents(t, []map[string]any{approved, failed}, []string{"Toolbox", "Wrench", "apiKey", "Bearer", "provider_session_id"})
}

func TestRealtimeVoiceActionPlanApprovalEmitsSafeExecutionFailure(t *testing.T) {
	t.Parallel()

	ctx, connection, sessionID := openRealtimeVoiceReviewSessionWithModel(t, unsupportedActionPlanProposalLanguageModel{})

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":      "action.plan.approve",
		"seq":       4,
		"sessionId": sessionID,
		"planId":    "plan-id",
	})
	approved := readRealtimeMessage(t, ctx, connection)
	if approved["type"] != "action.plan.approved" || approved["planId"] != "plan-id" || approved["status"] != "approved" {
		t.Fatalf("expected safe approval event, got %+v", approved)
	}
	failed := readRealtimeMessage(t, ctx, connection)
	if failed["type"] != "action.plan.failed" || failed["planId"] != "plan-id" || failed["status"] != "failed" {
		t.Fatalf("expected safe execution failure event, got %+v", failed)
	}
	assertSafeRealtimeEvents(t, []map[string]any{approved, failed}, []string{"asset-1", "location-1", "apiKey", "Bearer", "provider_session_id"})
}

func TestRealtimeVoiceActionPlanCancellationUsesOpenReviewSession(t *testing.T) {
	t.Parallel()

	ctx, connection, sessionID := openRealtimeVoiceReviewSession(t)
	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":      "action.plan.cancel",
		"seq":       4,
		"sessionId": sessionID,
		"planId":    "plan-id",
	})
	cancelled := readRealtimeMessage(t, ctx, connection)
	if cancelled["type"] != "action.plan.cancelled" || cancelled["planId"] != "plan-id" || cancelled["status"] != "cancelled" {
		t.Fatalf("expected safe cancellation event, got %+v", cancelled)
	}
}

func TestRealtimeVoiceActionPlanDecisionRejectsUnsafeMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message map[string]any
	}{
		{
			name: "stale sequence",
			message: map[string]any{
				"type":      "action.plan.approve",
				"seq":       3,
				"sessionId": "voice-session-id",
				"planId":    "plan-id",
			},
		},
		{
			name: "forged session",
			message: map[string]any{
				"type":      "action.plan.approve",
				"seq":       4,
				"sessionId": "voice-session-other",
				"planId":    "plan-id",
			},
		},
		{
			name: "wrong plan",
			message: map[string]any{
				"type":      "action.plan.approve",
				"seq":       4,
				"sessionId": "voice-session-id",
				"planId":    "plan-other",
			},
		},
		{
			name: "forbidden fields",
			message: map[string]any{
				"type":        "action.plan.approve",
				"seq":         4,
				"sessionId":   "voice-session-id",
				"planId":      "plan-id",
				"tenantId":    "tenant-other",
				"inventoryId": "inventory-other",
				"arguments":   map[string]any{"apiKey": "secret"},
			},
		},
		{
			name: "malformed type",
			message: map[string]any{
				"type":      "action.plan.execute",
				"seq":       4,
				"sessionId": "voice-session-id",
				"planId":    "plan-id",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, connection, _ := openRealtimeVoiceReviewSession(t)
			writeRealtimeMessage(t, ctx, connection, tt.message)
			failed := readRealtimeMessage(t, ctx, connection)
			if failed["type"] != "session.failed" {
				t.Fatalf("expected safe failure, got %+v", failed)
			}
			assertSafeRealtimeEvents(t, []map[string]any{failed}, []string{"apiKey", "secret", "tenant-other", "inventory-other", "provider_session_id"})
		})
	}
}

func openRealtimeVoiceReviewSession(t *testing.T) (context.Context, *websocket.Conn, string) {
	t.Helper()

	return openRealtimeVoiceReviewSessionWithModel(t, actionPlanProposalLanguageModel{})
}

func openRealtimeVoiceReviewSessionWithModel(t *testing.T, languageInference ports.LanguageInferenceProvider) (context.Context, *websocket.Conn, string) {
	t.Helper()

	return openRealtimeVoiceReviewSessionWithSetup(t, languageInference, nil)
}

func openRealtimeVoiceReviewSessionWithSetup(t *testing.T, languageInference ports.LanguageInferenceProvider, setup func(app.App)) (context.Context, *websocket.Conn, string) {
	t.Helper()

	ids := []string{"voice-session-id", "plan-id", "command-id", "response-id", "asset-id", "undo-id", "audit-id"}
	if setup != nil {
		ids = []string{
			"location-id", "location-undo-id", "location-audit-id",
			"asset-id", "asset-undo-id", "asset-audit-id",
			"voice-session-id", "plan-id", "command-id", "response-id", "move-undo-id", "move-audit-id",
		}
	}
	return openRealtimeVoiceReviewSessionWithSetupAndIDs(t, languageInference, ids, setup)
}

func openRealtimeVoiceReviewSessionWithSetupAndIDs(t *testing.T, languageInference ports.LanguageInferenceProvider, ids []string, setup func(app.App)) (context.Context, *websocket.Conn, string) {
	t.Helper()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         ids,
	}, fakeSpeechToText{transcript: "Add a water bottle."}, languageInference, fakeTextToSpeech{
		chunks: [][]byte{[]byte("spoken-audio")},
	})
	if setup != nil {
		setup(application)
	}
	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{
		HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:user-1"}},
	})
	if err != nil {
		t.Fatalf("dial realtime voice websocket: %v", err)
	}
	t.Cleanup(func() { _ = connection.Close(websocket.StatusNormalClosure, "") })

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":        "session.start",
		"seq":         1,
		"tenantId":    "tenant-home",
		"inventoryId": "inventory-home",
		"source":      "mobile_voice",
		"inputAudio":  map[string]any{"mimeType": "audio/mp4", "sampleRate": 44100, "channels": 1},
		"outputAudio": map[string]any{"mimeTypes": []string{"audio/mpeg"}},
	})
	started := readRealtimeMessage(t, ctx, connection)
	sessionID, _ := started["sessionId"].(string)
	if sessionID != "voice-session-id" {
		t.Fatalf("expected deterministic session id, got %+v", started)
	}
	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":         "audio.chunk",
		"seq":          2,
		"sessionId":    sessionID,
		"chunkId":      "chunk-1",
		"audioBase64":  base64.StdEncoding.EncodeToString([]byte("fake-audio")),
		"isFinalChunk": true,
	})
	writeRealtimeMessage(t, ctx, connection, map[string]any{"type": "audio.end", "seq": 3, "sessionId": sessionID})

	events := readRealtimeMessagesUntil(t, ctx, connection, "session.completed")
	findRealtimeEvent(t, events, "action.plan.proposed")
	return ctx, connection, sessionID
}

type actionPlanProposalLanguageModel struct{}

func (m actionPlanProposalLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	if len(input.ToolResults) == 0 {
		return ports.LanguageInferenceTurn{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "plan-tool-call",
				Name: "propose_action_plan",
				Arguments: map[string]any{
					"commandKind":                "create_asset",
					"intentSummary":              "Create a water bottle item.",
					"modelInterpretationSummary": "The user wants to add a water bottle to this inventory.",
					"confirmationSummary":        "Create item water bottle?",
					"commandSummary":             "Create item water bottle",
					"argumentsJson":              `{"kind":"item","name":"water bottle"}`,
					"riskSummary":                "Adds a new item to this inventory.",
				},
			}},
		}, nil
	}
	return ports.LanguageInferenceTurn{
		Final: &ports.StructuredAgentResponse{
			Kind:            ports.StructuredAgentResponseKindClarification,
			SpokenResponse:  "I prepared that change for review.",
			DisplayResponse: "I prepared that change for review.",
		},
	}, nil
}

type unsupportedActionPlanProposalLanguageModel struct{}

func (m unsupportedActionPlanProposalLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	if len(input.ToolResults) == 0 {
		return ports.LanguageInferenceTurn{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "plan-tool-call",
				Name: "propose_action_plan",
				Arguments: map[string]any{
					"commandKind":                "move_asset",
					"intentSummary":              "Move a visible item.",
					"modelInterpretationSummary": "The user wants to move an item.",
					"confirmationSummary":        "Move item?",
					"commandSummary":             "Move item",
					"argumentsJson":              `{"assetId":"asset-1","parentAssetId":"location-1"}`,
					"riskSummary":                "Moves an item in this inventory.",
				},
			}},
		}, nil
	}
	return ports.LanguageInferenceTurn{
		Final: &ports.StructuredAgentResponse{
			Kind:            ports.StructuredAgentResponseKindClarification,
			SpokenResponse:  "I prepared that change for review.",
			DisplayResponse: "I prepared that change for review.",
		},
	}, nil
}

type moveActionPlanProposalLanguageModel struct{}

func (m moveActionPlanProposalLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	if len(input.ToolResults) == 0 {
		return ports.LanguageInferenceTurn{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "plan-tool-call",
				Name: "propose_action_plan",
				Arguments: map[string]any{
					"commandKind":                "move_asset",
					"intentSummary":              "Move the water bottle to the office.",
					"modelInterpretationSummary": "The user wants the visible water bottle moved into Office.",
					"confirmationSummary":        "Move water bottle to Office?",
					"commandSummary":             "Move water bottle to Office",
					"argumentsJson":              `{"assetId":"asset-id","parentAssetId":"location-id"}`,
					"riskSummary":                "Moves an item in this inventory.",
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

type archiveActionPlanProposalLanguageModel struct{}

func (m archiveActionPlanProposalLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	if len(input.ToolResults) == 0 {
		return ports.LanguageInferenceTurn{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "plan-tool-call",
				Name: "propose_action_plan",
				Arguments: map[string]any{
					"commandKind":                "archive_asset",
					"intentSummary":              "Archive the water bottle.",
					"modelInterpretationSummary": "The user wants the visible water bottle archived.",
					"confirmationSummary":        "Archive water bottle?",
					"commandSummary":             "Archive water bottle",
					"argumentsJson":              `{"assetId":"asset-id"}`,
					"riskSummary":                "Archives an item in this inventory.",
				},
			}},
		}, nil
	}
	return ports.LanguageInferenceTurn{
		Final: &ports.StructuredAgentResponse{
			Kind:            ports.StructuredAgentResponseKindClarification,
			SpokenResponse:  "I prepared that archive for review.",
			DisplayResponse: "I prepared that archive for review.",
		},
	}, nil
}
