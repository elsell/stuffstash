package app

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type inventoryConversationModel struct {
	calls     int
	seenScope bool
}

func (m *inventoryConversationModel) Converse(_ context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	m.calls++
	m.seenScope = input.TenantID == "tenant-home" && input.InventoryID == "inventory-home" && input.Principal.ID == "user-1"
	last := input.Messages[len(input.Messages)-1]
	if last.Role == ports.ConversationRoleUser {
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "category-search", Name: RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "Acetone"}}}}, nil
	}
	var result realtimeVoiceAssetToolOutput
	if len(last.ToolResults) > 0 {
		_ = json.Unmarshal([]byte(last.ToolResults[0].Content), &result)
	}
	if len(result.Items) == 0 {
		return ports.ConversationModelTurn{Text: "I couldn't find any matching chemicals."}, nil
	}
	return ports.ConversationModelTurn{Answer: &ports.ConversationAnswer{Spoken: "Yes, you have chemicals, including Acetone.", Display: "Acetone is recorded in your inventory.", AssetIDs: []string{result.Items[0].AssetID}}}, nil
}
func TestRealtimeVoiceUsesModelLedToolsAndNaturalAnswer(t *testing.T) {
	resolver := successfulRealtimeVoiceResolver()
	model := &inventoryConversationModel{}
	resolver.providers.ConversationModel = model
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Do I have any chemicals?"}
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	item := realtimeVoiceInvestigationAsset("chemical-acetone", "Acetone", asset.KindItem, "")
	seedRealtimeVoiceLoopAsset(t, store, item, "audit-chemical")
	events := runRealtimeVoiceProductionEntrypoint(t, application)
	response := realtimeVoiceInvestigationCompletedResponse(events)
	if response == nil || response.SpokenResponse != "Yes, you have chemicals, including Acetone." {
		t.Fatalf("natural model answer lost: %+v", response)
	}
	if model.calls != 2 || !model.seenScope {
		t.Fatalf("model not invoked with scoped context: %+v", model)
	}
	if resolver.providers.TextToSpeech.(*resolvedTextToSpeech).lastText != response.SpokenResponse {
		t.Fatal("TTS did not receive model wording")
	}
	if len(response.Artifacts) != 1 || response.Artifacts[0].AssetID != item.ID {
		t.Fatal("authorized reference lost")
	}
}

type forgedReferenceConversation struct{ calls int }

func (m *forgedReferenceConversation) Converse(context.Context, ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	m.calls++
	return ports.ConversationModelTurn{Answer: &ports.ConversationAnswer{Spoken: "I found an item.", Display: "An item.", AssetIDs: []string{"unobserved-private-record"}}}, nil
}
func TestRealtimeVoiceModelCannotPublishUnobservedReference(t *testing.T) {
	resolver := successfulRealtimeVoiceResolver()
	model := &forgedReferenceConversation{}
	resolver.providers.ConversationModel = model
	application := newRealtimeVoiceResolutionTestApp(t, resolver)
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatal(err)
	}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error { return nil })
	if model.calls != 1 || err == nil {
		t.Fatalf("unobserved reference not rejected at conversation boundary: calls=%d err=%v", model.calls, err)
	}
	if resolver.providers.TextToSpeech.(*resolvedTextToSpeech).lastText != "" {
		t.Fatal("unvalidated reference reached speech")
	}
}
