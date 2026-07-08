package app

import (
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceFailedReadResultsDoNotSatisfyContextGates(t *testing.T) {
	t.Parallel()

	failedSearch := []ports.AgentToolResult{{
		Name:    RealtimeVoiceToolSearchAuthorizedAssets,
		Content: `{"tool":"search_authorized_assets","status":"error","code":"invalid_tool_request","retryable":true}`,
	}}
	if realtimeVoiceReadToolResultCount(failedSearch) != 0 {
		t.Fatalf("expected failed search result not to count as successful read context")
	}
	if realtimeVoiceShouldFinalizeReadOnlyAfterToolTurn("Where is the water bottle?", failedSearch) {
		t.Fatalf("expected failed search result not to allow read-only finalization")
	}
	if realtimeVoiceShouldUseConstrainedPlanner("Move my water bottle to the kitchen.", 2, failedSearch) {
		t.Fatalf("expected failed search result not to allow planner readiness")
	}

	failedList := []ports.AgentToolResult{{
		Name:    RealtimeVoiceToolListAuthorizedAssets,
		Content: `{"tool":"list_authorized_assets","status":"error","code":"invalid_tool_request","retryable":true}`,
	}}
	if realtimeVoiceHasListResult(failedList) {
		t.Fatalf("expected failed list result not to satisfy contents-list gate")
	}

	failedAuditHistory := []ports.AgentToolResult{
		{
			Name:    RealtimeVoiceToolSearchAuthorizedAssets,
			Content: `{"tool":"search_authorized_assets","items":[{"assetId":"water-bottle-1","title":"Water bottle","kind":"item"}]}`,
		},
		{
			Name:    RealtimeVoiceToolListAssetAuditHistory,
			Content: `{"tool":"list_asset_audit_history","status":"error","code":"invalid_tool_request","retryable":true}`,
		},
	}
	if realtimeVoiceHasAssetAuditHistoryResult(failedAuditHistory) {
		t.Fatalf("expected failed audit history result not to satisfy history gate")
	}
	if realtimeVoiceShouldFinalizeReadOnlyAfterToolTurn("When did I move the water bottle?", failedAuditHistory) {
		t.Fatalf("expected failed audit history result not to allow history finalization")
	}
}
