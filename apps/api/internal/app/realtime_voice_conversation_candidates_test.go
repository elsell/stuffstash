package app

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

func TestConversationSearchRetainsCompetingCandidates(t *testing.T) {
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, successfulRealtimeVoiceResolver())
	for _, id := range []string{"counter-one", "counter-two"} {
		seedRealtimeVoiceLoopAsset(t, store, realtimeVoiceInvestigationAsset(id, "Coffee Counter", asset.KindContainer, ""), "audit-"+id)
	}
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatal(err)
	}
	executor := &realtimeConversationTools{application: application, session: session, emit: func(RealtimeVoiceEvent) error { return nil }, visible: map[string]struct{}{}, items: map[string]realtimeVoiceAssetToolItem{}}
	for _, args := range []map[string]any{{"query": "Coffee Counter", "limit": "one"}, {"query": ""}, {"query": "Coffee Counter", "unsupported": true}} {
		failed, err := executor.ExecuteConversationTool(context.Background(), ports.AgentToolCall{ID: "malformed", Name: RealtimeVoiceToolSearchAuthorizedAssets, Arguments: args})
		var feedback map[string]any
		if err != nil || json.Unmarshal([]byte(failed.Result.Content), &feedback) != nil || feedback["error"] == nil {
			t.Fatalf("malformed search prevented correction: result=%+v err=%v", failed, err)
		}
	}
	result, err := executor.ExecuteConversationTool(context.Background(), ports.AgentToolCall{ID: "search", Name: RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "Coffee Counter", "limit": 1}})
	if err != nil {
		t.Fatal(err)
	}
	var output realtimeVoiceAssetToolOutput
	if err := json.Unmarshal([]byte(result.Result.Content), &output); err != nil {
		t.Fatal(err)
	}
	if len(output.Items) != 2 {
		t.Fatalf("model-selected single result hid competing destination: %+v", output)
	}
}
