package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type moveConversationModel struct {
	calls  int
	forge  bool
	nested bool
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
	commands := []any{map[string]any{"id": "move-drill", "kind": "move_asset", "summary": "Move the Drill into the Garage.", "arguments": map[string]any{"assetId": ids["Drill"], "parentAssetId": ids["Garage"]}}}
	if m.nested {
		commands = []any{
			map[string]any{"id": "create-box", "kind": "create_asset", "summary": "Create Blue Box in Garage.", "arguments": map[string]any{"title": "Blue Box", "kind": "container", "parentAssetId": ids["Garage"]}},
			map[string]any{"id": "move-drill", "kind": "move_asset", "summary": "Move Drill into Blue Box.", "arguments": map[string]any{"assetId": ids["Drill"], "parentCommandId": "create-box"}},
		}
	}
	return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{
		{ID: "propose-move", Name: "propose_inventory_change", Arguments: map[string]any{"summary": "Move the Drill into the Garage.", "commands": commands}},
		{ID: "must-not-run", Name: RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "Unrelated"}},
	}}, nil
}
func TestModelLedMoveProposesExistingItemWithoutExecuting(t *testing.T) {
	for _, tc := range []struct {
		name           string
		forged, nested bool
	}{{name: "authorized"}, {name: "forged reference", forged: true}, {name: "dependent create and move", nested: true}} {
		t.Run(tc.name, func(t *testing.T) {
			forged := tc.forged
			resolver := successfulRealtimeVoiceResolver()
			model := &moveConversationModel{forge: forged, nested: tc.nested}
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
			expectedCommands := 1
			if tc.nested {
				expectedCommands = 2
			}
			if proposal == nil || len(proposal.Commands) != expectedCommands || model.calls != 2 {
				t.Fatalf("existing item move not proposed or loop continued: %+v calls=%d", proposal, model.calls)
			}
			last := proposal.Commands[len(proposal.Commands)-1]
			if tc.nested && last.ParentCommandID != "create-box" {
				t.Fatal("dependent destination lost")
			}
			if last.Title != drill.Title.String() || (!tc.nested && last.ParentAssetID != garage.ID.String()) {
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

// Domain reads retain the working service repository; the review projection
// encounters a storage outage while resolving its display metadata.
type unavailableProposalReviewAssets struct {
	ports.AssetRepository
	calls int
}

func (r *unavailableProposalReviewAssets) AssetByID(context.Context, tenant.ID, inventory.InventoryID, asset.ID) (asset.Asset, bool, error) {
	r.calls++
	return asset.Asset{}, false, errors.New("review read unavailable")
}
func TestModelLedProposalReviewFailureDoesNotLeaveRetryableSavedDraft(t *testing.T) {
	resolver := successfulRealtimeVoiceResolver()
	model := &moveConversationModel{}
	resolver.providers.ConversationModel = model
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	seedRealtimeVoiceLoopAsset(t, store, realtimeVoiceInvestigationAsset("existing-drill", "Drill", asset.KindItem, ""), "audit-drill")
	seedRealtimeVoiceLoopAsset(t, store, realtimeVoiceInvestigationAsset("existing-garage", "Garage", asset.KindLocation, ""), "audit-garage")
	plans := &fakeActionPlanRepository{}
	application.actionPlans = plans
	review := &unavailableProposalReviewAssets{AssetRepository: application.assets}
	application.assets = review
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatal(err)
	}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error { return nil })
	if len(plans.saved) != 0 {
		t.Fatalf("model retried after saving an undisplayed draft: saved=%d calls=%d", len(plans.saved), model.calls)
	}
	if err != nil || review.calls == 0 || model.calls != 3 {
		t.Fatalf("expected review outage followed by a model explanation: reads=%d calls=%d err=%v", review.calls, model.calls, err)
	}
}
