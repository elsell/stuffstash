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
)

func TestRealtimeVoiceActionPlanApprovalEmitsSafeExecutionFailure(t *testing.T) {
	t.Parallel()

	var application app.App
	ctx, connection, sessionID, planID := openRealtimeVoiceReviewSessionWithSetupAndIDsAndTranscript(t, archiveActionPlanProposalLanguageModel{}, []string{
		"location-id", "location-undo-id", "location-audit-id",
		"asset-id", "asset-undo-id", "asset-audit-id",
		"child-id", "child-undo-id", "child-audit-id",
		"voice-session-id", "plan-id", "command-id", "response-id", "archive-undo-id", "archive-audit-id",
	}, "Archive Toolbox.", func(seedApplication app.App) {
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
		t.Fatalf("expected safe execution failure event, got %+v", failed)
	}
	item, err := application.GetAsset(context.Background(), app.GetAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		Source:      audit.SourceAPI,
		RequestID:   "assert-safe-failed-asset",
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		AssetID:     asset.ID("asset-id"),
	})
	if err != nil {
		t.Fatalf("read failed asset: %v", err)
	}
	if item.LifecycleState != asset.LifecycleStateActive {
		t.Fatalf("expected failed execution to leave asset active, got %+v", item)
	}
	assertSafeRealtimeEvents(t, []map[string]any{approved, failed}, []string{"Toolbox", "Wrench", "apiKey", "Bearer", "provider_session_id"})
}
