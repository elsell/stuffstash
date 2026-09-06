package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type historyConversationModel struct {
	tool     string
	calls    int
	observed string
}

func (m *historyConversationModel) Converse(_ context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	m.calls++
	available := false
	for _, tool := range input.Tools {
		if tool.Name == m.tool {
			available = true
		}
	}
	if !available {
		return ports.ConversationModelTurn{}, errors.New("requested history read is unavailable to the model")
	}
	if m.calls == 1 {
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "find", Name: RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "Drill"}}}}, nil
	}
	last := input.Messages[len(input.Messages)-1]
	if len(last.ToolResults) != 1 {
		return ports.ConversationModelTurn{}, errors.New("missing read evidence")
	}
	if m.calls == 2 {
		var found realtimeVoiceAssetToolOutput
		if json.Unmarshal([]byte(last.ToolResults[0].Content), &found) != nil || len(found.Items) != 1 {
			return ports.ConversationModelTurn{}, errors.New("missing drill")
		}
		args := map[string]any{"limit": 10}
		if m.tool != RealtimeVoiceToolListCheckedOutAssets {
			args["assetId"] = found.Items[0].AssetID
		}
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "read", Name: m.tool, Arguments: args}}}, nil
	}
	m.observed = last.ToolResults[0].Content
	var evidence struct {
		Count int `json:"count"`
	}
	if json.Unmarshal([]byte(m.observed), &evidence) != nil || evidence.Count == 0 {
		return ports.ConversationModelTurn{}, errors.New("read did not return the seeded history")
	}
	return ports.ConversationModelTurn{Text: "Your drill has a recorded checkout."}, nil
}
func TestConversationModelCanReadHistoryAndCheckoutEvidence(t *testing.T) {
	for _, tool := range []string{RealtimeVoiceToolListAssetAuditHistory, RealtimeVoiceToolListCheckedOutAssets, RealtimeVoiceToolListAssetCheckoutHistory} {
		t.Run(tool, func(t *testing.T) {
			resolver := successfulRealtimeVoiceResolver()
			model := &historyConversationModel{tool: tool}
			resolver.providers.ConversationModel = model
			application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
			drill := realtimeVoiceInvestigationAsset("drill", "Drill", asset.KindItem, "")
			seedRealtimeVoiceLoopAsset(t, store, drill, "created-drill")
			sessionInput := defaultRealtimeVoiceSessionInput()
			_, err := application.CheckoutAsset(context.Background(), CheckoutAssetInput{Principal: sessionInput.Principal, Source: audit.SourceAPI, TenantID: sessionInput.TenantID, InventoryID: sessionInput.InventoryID, AssetID: drill.ID, Details: "Borrowed for shelves"})
			if err != nil {
				t.Fatal(err)
			}
			response := realtimeVoiceInvestigationCompletedResponse(runRealtimeVoiceProductionEntrypoint(t, application))
			if response == nil || response.SpokenResponse != "Your drill has a recorded checkout." || model.calls != 3 {
				t.Fatalf("history answer lost: response=%+v calls=%d", response, model.calls)
			}
			if !strings.Contains(model.observed, "Drill") {
				t.Fatal("history did not identify the authorized drill")
			}
			if tool == RealtimeVoiceToolListAssetCheckoutHistory && !strings.Contains(model.observed, "Borrowed for shelves") {
				t.Fatal("checkout details did not reach the model")
			}
		})
	}
}
