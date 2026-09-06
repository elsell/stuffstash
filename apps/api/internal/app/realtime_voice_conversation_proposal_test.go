package app

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type moveConversationModel struct {
	calls int
	forge bool
}

func (m *moveConversationModel) Converse(_ context.Context, in ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	m.calls++
	if m.calls == 1 {
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{
			{ID: "find-drill", Name: RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "Drill"}},
			{ID: "find-garage", Name: RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "Garage"}},
		}}, nil
	}
	if m.calls > 2 {
		return ports.ConversationModelTurn{Text: "I couldn't prepare that change."}, nil
	}
	ids := map[string]string{}
	for _, msg := range in.Messages {
		for _, result := range msg.ToolResults {
			var output realtimeVoiceAssetToolOutput
			_ = json.Unmarshal([]byte(result.Content), &output)
			for _, item := range output.Items {
				ids[item.Title] = item.AssetID
			}
		}
	}
	if m.forge {
		ids["Drill"] = "unobserved-drill"
	}
	return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{
		{ID: "propose-move", Name: "propose_inventory_change", Arguments: map[string]any{"summary": "Move the Drill into the Garage.", "commands": []any{map[string]any{"id": "move-drill", "kind": "move_asset", "summary": "Move the Drill into the Garage.", "arguments": map[string]any{"assetId": ids["Drill"], "parentAssetId": ids["Garage"]}}}}},
		{ID: "must-not-run", Name: RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "Unrelated"}},
	}}, nil
}
func TestModelLedMoveProposesExistingItemWithoutExecuting(t *testing.T) {
	for _, forged := range []bool{false, true} {
		t.Run(map[bool]string{false: "authorized", true: "forged reference"}[forged], func(t *testing.T) {
			resolver := successfulRealtimeVoiceResolver()
			model := &moveConversationModel{forge: forged}
			resolver.providers.ConversationModel = model
			resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Move my Drill into the Garage."}
			application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
			drill := realtimeVoiceInvestigationAsset("existing-drill", "Drill", asset.KindItem, "")
			garage := realtimeVoiceInvestigationAsset("existing-garage", "Garage", asset.KindLocation, "")
			seedRealtimeVoiceLoopAsset(t, store, drill, "audit-drill")
			seedRealtimeVoiceLoopAsset(t, store, garage, "audit-garage")
			events := runRealtimeVoiceProductionEntrypoint(t, application)
			proposal := realtimeVoiceInvestigationProposedPlan(events)
			current, found, err := store.AssetByID(context.Background(), "tenant-home", "inventory-home", drill.ID)
			if err != nil || !found || current.ParentAssetID != drill.ParentAssetID {
				t.Fatal("inventory changed without approval")
			}
			if forged {
				if proposal != nil {
					t.Fatal("unobserved reference became a proposal")
				}
				return
			}
			if proposal == nil || len(proposal.Commands) != 1 || model.calls != 2 {
				t.Fatalf("existing item move not proposed or loop continued: %+v calls=%d", proposal, model.calls)
			}
			if proposal.Commands[0].AssetID != drill.ID.String() || proposal.Commands[0].ParentAssetID != garage.ID.String() {
				t.Fatalf("wrong move references: %+v", proposal.Commands)
			}
			for _, event := range events {
				if event.ToolCallID == "must-not-run" {
					t.Fatal("loop executed work after approval pause")
				}
			}
		})
	}
}
