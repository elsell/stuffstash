package httpserver

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"nhooyr.io/websocket"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
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

	events := runRealtimeVoiceQuestionUntil(t, server.URL, "tenant-home", "inventory-home", "user-1", "action.plan.proposed")
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
	if hasRealtimeEvent(events, "assistant.response.completed") || hasRealtimeEvent(events, "session.completed") {
		t.Fatalf("expected action plan proposal to pause review before final response, got %+v", events)
	}
}

func TestRealtimeVoiceActionPlanApprovalUsesOpenReviewSession(t *testing.T) {
	t.Parallel()

	ctx, connection, sessionID, planID := openRealtimeVoiceReviewSession(t)

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
		t.Fatalf("expected safe execution event, got %+v", executed)
	}
	commandResults, ok := executed["commandResults"].([]any)
	if !ok || len(commandResults) != 1 {
		t.Fatalf("expected one command result for photo attachment, got %+v", executed["commandResults"])
	}
	commandResult, ok := commandResults[0].(map[string]any)
	if !ok || commandResult["commandId"] == "" || commandResult["assetId"] == "" || commandResult["operation"] != "create" || commandResult["assetKind"] != "item" || commandResult["title"] != nil {
		t.Fatalf("unexpected command result: %+v", commandResult)
	}
	assertSafeRealtimeEvents(t, []map[string]any{approved, executed}, []string{"fake-audio", "apiKey", "Bearer", "provider_session_id"})
}

func TestRealtimeVoiceActionPlanApprovalAcceptsBoundedCreateEdits(t *testing.T) {
	t.Parallel()
	ctx, connection, sessionID, planID := openRealtimeVoiceReviewSession(t)
	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type": "action.plan.approve", "seq": 4, "sessionId": sessionID, "planId": planID,
		"commandEdits": []map[string]any{{"commandId": "command-id", "title": "Insulated water bottle", "parent": map[string]any{"kind": "root"}}},
	})
	approved := readRealtimeMessage(t, ctx, connection)
	if approved["type"] != "action.plan.approved" || approved["status"] != "approved" {
		t.Fatalf("expected edited approval, got %+v", approved)
	}
	executed := readRealtimeMessage(t, ctx, connection)
	if executed["type"] != "action.plan.executed" || executed["status"] != "executed" {
		t.Fatalf("expected edited plan execution, got %+v", executed)
	}
}

func TestRealtimeVoiceActionPlanApprovalExecutesMoveAsset(t *testing.T) {
	t.Parallel()

	var application app.App
	ctx, connection, sessionID, planID := openRealtimeVoiceReviewSessionWithSetupAndTranscript(t, moveActionPlanProposalLanguageModel{}, "Move the water bottle to the office.", func(seedApplication app.App) {
		application = seedApplication
		seedVoiceAsset(t, seedApplication, "user-1", "tenant-home", "inventory-home", "location", "Office", "")
		seedVoiceAsset(t, seedApplication, "user-1", "tenant-home", "inventory-home", "item", "Water bottle", "")
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
		t.Fatalf("expected safe move execution event, got %+v", executed)
	}
	commandResults, ok := executed["commandResults"].([]any)
	if !ok || len(commandResults) != 1 {
		t.Fatalf("expected one move command result for photo attachment, got %+v", executed["commandResults"])
	}
	commandResult, ok := commandResults[0].(map[string]any)
	if !ok || commandResult["commandId"] == "" || commandResult["assetId"] == "" || commandResult["operation"] != "move" || commandResult["assetKind"] != "item" || commandResult["title"] != nil {
		t.Fatalf("unexpected move command result: %+v", commandResult)
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
	ctx, connection, sessionID, planID := openRealtimeVoiceReviewSessionWithSetup(t, archiveActionPlanProposalLanguageModel{}, func(seedApplication app.App) {
		application = seedApplication
		seedVoiceAsset(t, seedApplication, "user-1", "tenant-home", "inventory-home", "location", "Office", "")
		seedVoiceAsset(t, seedApplication, "user-1", "tenant-home", "inventory-home", "item", "Water bottle", "")
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
	ctx, connection, sessionID, planID := openRealtimeVoiceReviewSessionWithSetupAndIDs(t, archiveActionPlanProposalLanguageModel{}, []string{
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
		"planId":    planID,
	})
	approved := readRealtimeMessage(t, ctx, connection)
	if approved["type"] != "action.plan.approved" || approved["planId"] != planID || approved["status"] != "approved" {
		t.Fatalf("expected safe approval event, got %+v", approved)
	}
	failed := readRealtimeMessage(t, ctx, connection)
	if failed["type"] != "action.plan.failed" || failed["planId"] != planID || failed["status"] != "failed" {
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

func TestRealtimeVoiceActionPlanApprovalExecutesRestoreAsset(t *testing.T) {
	t.Parallel()

	var application app.App
	ctx, connection, sessionID, planID := openRealtimeVoiceReviewSessionWithSetupAndIDs(t, restoreActionPlanProposalLanguageModel{}, []string{
		"location-id", "location-undo-id", "location-audit-id",
		"asset-id", "asset-undo-id", "asset-audit-id",
		"seed-archive-undo-id", "seed-archive-audit-id",
		"voice-session-id", "plan-id", "command-id", "response-id", "restore-undo-id", "restore-audit-id",
	}, func(seedApplication app.App) {
		application = seedApplication
		seedVoiceAsset(t, seedApplication, "user-1", "tenant-home", "inventory-home", "location", "Office", "")
		seedVoiceAsset(t, seedApplication, "user-1", "tenant-home", "inventory-home", "item", "Water bottle", "")
		_, err := seedApplication.ArchiveAssetWithOperation(context.Background(), app.UpdateAssetLifecycleInput{
			Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
			Source:      audit.SourceAPI,
			RequestID:   "seed-archive-water-bottle",
			TenantID:    tenant.ID("tenant-home"),
			InventoryID: inventory.InventoryID("inventory-home"),
			AssetID:     asset.ID("asset-id"),
		})
		if err != nil {
			t.Fatalf("seed archived asset: %v", err)
		}
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
		t.Fatalf("expected safe restore execution event, got %+v", executed)
	}
	restored, err := application.GetAsset(context.Background(), app.GetAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		Source:      audit.SourceAPI,
		RequestID:   "assert-restored-asset",
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		AssetID:     asset.ID("asset-id"),
	})
	if err != nil {
		t.Fatalf("read restored asset: %v", err)
	}
	if restored.LifecycleState != asset.LifecycleStateActive {
		t.Fatalf("expected realtime-approved restore to update asset lifecycle, got %+v", restored)
	}
	assertSafeRealtimeEvents(t, []map[string]any{approved, executed}, []string{"Water bottle", "Office", "apiKey", "Bearer", "provider_session_id"})
}

func TestRealtimeVoiceActionPlanRestoreApprovalDeniedSafelyWithoutMutation(t *testing.T) {
	t.Parallel()

	authorizer := &denyEditAfterProposalAuthorizer{delegate: memory.NewAuthorizer()}
	application := newSeededTestAppWithAuthorizer(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids: []string{
			"location-id", "location-undo-id", "location-audit-id",
			"asset-id", "asset-undo-id", "asset-audit-id",
			"seed-archive-undo-id", "seed-archive-audit-id",
			"voice-session-id", "plan-id", "command-id", "response-id", "restore-undo-id", "restore-audit-id",
		},
	}, authorizer).WithRealtimeVoiceProviders(fakeSpeechToText{transcript: "Restore the water bottle."}, restoreActionPlanProposalLanguageModel{}, fakeTextToSpeech{
		chunks: [][]byte{[]byte("spoken-audio")},
	})
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "location", "Office", "")
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "item", "Water bottle", "")
	if _, err := application.ArchiveAssetWithOperation(context.Background(), app.UpdateAssetLifecycleInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		Source:      audit.SourceAPI,
		RequestID:   "seed-archive-water-bottle",
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		AssetID:     asset.ID("asset-id"),
	}); err != nil {
		t.Fatalf("seed archived asset: %v", err)
	}

	ctx, connection, sessionID, planID := openRealtimeVoiceReviewSessionForApplication(t, application)
	authorizer.denyEdit.Store(true)
	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":      "action.plan.approve",
		"seq":       4,
		"sessionId": sessionID,
		"planId":    planID,
	})
	failed := readRealtimeMessage(t, ctx, connection)
	if failed["type"] != "session.failed" {
		t.Fatalf("expected safe approval denial, got %+v", failed)
	}
	archived, err := application.GetAsset(context.Background(), app.GetAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		Source:      audit.SourceAPI,
		RequestID:   "assert-restore-denied-asset",
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		AssetID:     asset.ID("asset-id"),
	})
	if err != nil {
		t.Fatalf("read denied restore asset: %v", err)
	}
	if archived.LifecycleState != asset.LifecycleStateArchived {
		t.Fatalf("expected denied restore to leave asset archived, got %+v", archived)
	}
	assertSafeRealtimeEvents(t, []map[string]any{failed}, []string{"asset-id", "tenant-home", "inventory-home", "apiKey", "Bearer", "provider_session_id"})
}

func TestRealtimeVoiceActionPlanCancellationUsesOpenReviewSession(t *testing.T) {
	t.Parallel()

	ctx, connection, sessionID, planID := openRealtimeVoiceReviewSession(t)
	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":      "action.plan.cancel",
		"seq":       4,
		"sessionId": sessionID,
		"planId":    planID,
	})
	cancelled := readRealtimeMessage(t, ctx, connection)
	if cancelled["type"] != "action.plan.cancelled" || cancelled["planId"] != planID || cancelled["status"] != "cancelled" {
		t.Fatalf("expected safe cancellation event, got %+v", cancelled)
	}
}

func openRealtimeVoiceReviewSession(t *testing.T) (context.Context, *websocket.Conn, string, string) {
	t.Helper()

	return openRealtimeVoiceReviewSessionWithModel(t, actionPlanProposalLanguageModel{})
}

func openRealtimeVoiceReviewSessionWithModel(t *testing.T, languageInference ports.LanguageInferenceProvider) (context.Context, *websocket.Conn, string, string) {
	t.Helper()

	return openRealtimeVoiceReviewSessionWithSetup(t, languageInference, nil)
}

func openRealtimeVoiceReviewSessionWithSetup(t *testing.T, languageInference ports.LanguageInferenceProvider, setup func(app.App)) (context.Context, *websocket.Conn, string, string) {
	t.Helper()

	return openRealtimeVoiceReviewSessionWithSetupAndTranscript(t, languageInference, "Add a water bottle.", setup)
}

func openRealtimeVoiceReviewSessionWithSetupAndTranscript(t *testing.T, languageInference ports.LanguageInferenceProvider, transcript string, setup func(app.App)) (context.Context, *websocket.Conn, string, string) {
	t.Helper()

	ids := []string{"voice-session-id", "plan-id", "command-id", "response-id", "asset-id", "undo-id", "audit-id"}
	if setup != nil {
		ids = []string{
			"location-id", "location-undo-id", "location-audit-id",
			"asset-id", "asset-undo-id", "asset-audit-id",
			"voice-session-id", "plan-id", "command-id", "response-id", "move-undo-id", "move-audit-id",
		}
	}
	return openRealtimeVoiceReviewSessionWithSetupAndIDsAndTranscript(t, languageInference, ids, transcript, setup)
}

func openRealtimeVoiceReviewSessionWithSetupAndIDs(t *testing.T, languageInference ports.LanguageInferenceProvider, ids []string, setup func(app.App)) (context.Context, *websocket.Conn, string, string) {
	t.Helper()

	ctx, connection, sessionID, planID, _ := openRealtimeVoiceReviewSessionWithSetupAndIDsAndProposal(t, languageInference, ids, setup)
	return ctx, connection, sessionID, planID
}

func openRealtimeVoiceReviewSessionWithSetupAndIDsAndProposal(t *testing.T, languageInference ports.LanguageInferenceProvider, ids []string, setup func(app.App)) (context.Context, *websocket.Conn, string, string, map[string]any) {
	t.Helper()

	return openRealtimeVoiceReviewSessionWithSetupAndIDsAndTranscriptAndProposal(t, languageInference, ids, "Add a water bottle.", setup)
}

func openRealtimeVoiceReviewSessionWithSetupAndIDsAndTranscript(t *testing.T, languageInference ports.LanguageInferenceProvider, ids []string, transcript string, setup func(app.App)) (context.Context, *websocket.Conn, string, string) {
	t.Helper()

	ctx, connection, sessionID, planID, _ := openRealtimeVoiceReviewSessionWithSetupAndIDsAndTranscriptAndProposal(t, languageInference, ids, transcript, setup)
	return ctx, connection, sessionID, planID
}

func openRealtimeVoiceReviewSessionWithSetupAndIDsAndTranscriptAndProposal(t *testing.T, languageInference ports.LanguageInferenceProvider, ids []string, transcript string, setup func(app.App)) (context.Context, *websocket.Conn, string, string, map[string]any) {
	t.Helper()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         ids,
	}, fakeSpeechToText{transcript: transcript}, languageInference, fakeTextToSpeech{
		chunks: [][]byte{[]byte("spoken-audio")},
	})
	if setup != nil {
		setup(application)
	}
	return openRealtimeVoiceReviewSessionForApplicationWithProposal(t, application)
}

func openRealtimeVoiceReviewSessionForApplication(t *testing.T, application app.App) (context.Context, *websocket.Conn, string, string) {
	t.Helper()

	ctx, connection, sessionID, planID, _ := openRealtimeVoiceReviewSessionForApplicationWithProposal(t, application)
	return ctx, connection, sessionID, planID
}

func openRealtimeVoiceReviewSessionForApplicationWithProposal(t *testing.T, application app.App) (context.Context, *websocket.Conn, string, string, map[string]any) {
	t.Helper()

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

	writeRealtimeMessage(t, ctx, connection, realtimeVoiceStartMessage("tenant-home", "inventory-home"))
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

	events := readRealtimeMessagesUntil(t, ctx, connection, "action.plan.proposed")
	proposed := findRealtimeEvent(t, events, "action.plan.proposed")
	actionPlan, ok := proposed["actionPlan"].(map[string]any)
	if !ok {
		t.Fatalf("expected action plan payload, got %+v", proposed)
	}
	planID, _ := actionPlan["planId"].(string)
	if planID == "" {
		t.Fatalf("expected proposed action plan id, got %+v", actionPlan)
	}
	return ctx, connection, sessionID, planID, actionPlan
}

func hasRealtimeEvent(events []map[string]any, eventType string) bool {
	for _, event := range events {
		if event["type"] == eventType {
			return true
		}
	}
	return false
}

type denyEditAfterProposalAuthorizer struct {
	delegate ports.Authorizer
	denyEdit atomic.Bool
}

func (d *denyEditAfterProposalAuthorizer) CheckTenant(ctx context.Context, principal identity.Principal, permission ports.TenantPermission, tenantID tenant.ID) error {
	return d.delegate.CheckTenant(ctx, principal, permission, tenantID)
}

func (d *denyEditAfterProposalAuthorizer) CheckInventory(ctx context.Context, principal identity.Principal, permission ports.InventoryPermission, inventoryID inventory.InventoryID) error {
	if d.denyEdit.Load() && permission == ports.InventoryPermissionEditAsset {
		return ports.ErrForbidden
	}
	return d.delegate.CheckInventory(ctx, principal, permission, inventoryID)
}

func (d *denyEditAfterProposalAuthorizer) ListViewableInventoryIDs(ctx context.Context, principal identity.Principal, tenantID tenant.ID, candidates []inventory.InventoryID) ([]inventory.InventoryID, error) {
	return d.delegate.ListViewableInventoryIDs(ctx, principal, tenantID, candidates)
}

func (d *denyEditAfterProposalAuthorizer) GrantTenantOwner(ctx context.Context, principal identity.Principal, tenantID tenant.ID) error {
	return d.delegate.GrantTenantOwner(ctx, principal, tenantID)
}

func (d *denyEditAfterProposalAuthorizer) GrantInventoryOwner(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return d.delegate.GrantInventoryOwner(ctx, principal, tenantID, inventoryID)
}

func (d *denyEditAfterProposalAuthorizer) GrantInventoryViewer(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return d.delegate.GrantInventoryViewer(ctx, principal, tenantID, inventoryID)
}

func (d *denyEditAfterProposalAuthorizer) GrantInventoryEditor(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return d.delegate.GrantInventoryEditor(ctx, principal, tenantID, inventoryID)
}

func (d *denyEditAfterProposalAuthorizer) RevokeInventoryViewer(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return d.delegate.RevokeInventoryViewer(ctx, principal, tenantID, inventoryID)
}

func (d *denyEditAfterProposalAuthorizer) RevokeInventoryEditor(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return d.delegate.RevokeInventoryEditor(ctx, principal, tenantID, inventoryID)
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

type moveActionPlanProposalLanguageModel struct{}

func (m moveActionPlanProposalLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
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
		parentID, err := firstToolResultAssetID(input.ToolResults[0].Content, "Office", "active")
		if err != nil {
			return ports.LanguageInferenceTurn{}, err
		}
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
					"argumentsJson":              `{"assetId":"` + assetID + `","parentAssetId":"` + parentID + `"}`,
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
		assetID, err := firstToolResultAnyAssetID(input.ToolResults[0].Content, "active", "Water bottle", "Toolbox")
		if err != nil {
			return ports.LanguageInferenceTurn{}, err
		}
		return ports.LanguageInferenceTurn{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "plan-tool-call",
				Name: "propose_action_plan",
				Arguments: map[string]any{
					"intentSummary":              "Archive the selected asset.",
					"modelInterpretationSummary": "The user wants the visible selected asset archived.",
					"confirmationSummary":        "Archive selected asset?",
					"commands": []any{map[string]any{
						"id":      "cmd-archive-asset",
						"kind":    "archive_asset",
						"summary": "Archive selected asset",
						"arguments": map[string]any{
							"assetId": assetID,
						},
					}},
					"riskSummary": "Archives an asset in this inventory.",
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

type restoreActionPlanProposalLanguageModel struct{}

func (m restoreActionPlanProposalLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	if len(input.ToolResults) == 0 {
		return ports.LanguageInferenceTurn{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "list-archived-water-bottle",
				Name: "list_authorized_assets",
				Arguments: map[string]any{
					"kind":           "item",
					"lifecycleState": "archived",
					"limit":          float64(10),
				},
			}},
		}, nil
	}
	if len(input.ToolResults) == 1 {
		assetID, err := firstToolResultAssetID(input.ToolResults[0].Content, "Water bottle", "archived")
		if err != nil {
			return ports.LanguageInferenceTurn{}, err
		}
		return ports.LanguageInferenceTurn{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "plan-tool-call",
				Name: "propose_action_plan",
				Arguments: map[string]any{
					"intentSummary":              "Restore the water bottle.",
					"modelInterpretationSummary": "The user wants the visible archived water bottle restored.",
					"confirmationSummary":        "Restore water bottle?",
					"commands": []any{map[string]any{
						"id":      "cmd-restore-water-bottle",
						"kind":    "restore_asset",
						"summary": "Restore water bottle",
						"arguments": map[string]any{
							"assetId": assetID,
						},
					}},
					"riskSummary": "Restores an item in this inventory.",
				},
			}},
		}, nil
	}
	return ports.LanguageInferenceTurn{
		Final: &ports.StructuredAgentResponse{
			Kind:            ports.StructuredAgentResponseKindClarification,
			SpokenResponse:  "I prepared that restore for review.",
			DisplayResponse: "I prepared that restore for review.",
		},
	}, nil
}
