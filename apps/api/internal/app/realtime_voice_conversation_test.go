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

type inventoryConversationModel struct {
	query, answer string
	calls         int
	seenScope     bool
}

func (m *inventoryConversationModel) Converse(_ context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	m.calls++
	m.seenScope = input.TenantID == "tenant-home" && input.InventoryID == "inventory-home" && input.Principal.ID == "user-1"
	last := input.Messages[len(input.Messages)-1]
	if last.Role == ports.ConversationRoleUser {
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "category-search", Name: RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": m.searchQuery()}}}}, nil
	}
	if len(last.ToolResults) > 0 {
		var failure map[string]any
		_ = json.Unmarshal([]byte(last.ToolResults[0].Content), &failure)
		if failure["error"] != nil && m.calls == 2 {
			return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "category-retry", Name: RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": m.searchQuery()}}}}, nil
		}
	}
	var result realtimeVoiceAssetToolOutput
	if len(last.ToolResults) > 0 {
		_ = json.Unmarshal([]byte(last.ToolResults[0].Content), &result)
	}
	if len(result.Items) == 0 {
		return ports.ConversationModelTurn{Text: "I couldn't find any matching chemicals."}, nil
	}
	spoken := m.answer
	if spoken == "" {
		spoken = "Yes, you have chemicals, including Acetone."
	}
	return ports.ConversationModelTurn{Answer: &ports.ConversationAnswer{Spoken: spoken, Display: "Matching belongings are shown below.", AssetIDs: []string{result.Items[0].AssetID}}}, nil
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

func (m *inventoryConversationModel) searchQuery() string {
	if m.query != "" {
		return m.query
	}
	return "Acetone"
}
func TestRealtimeConversationAcceptsNaturalTitleWithoutPhraseVeto(t *testing.T) {
	resolver := successfulRealtimeVoiceResolver()
	model := &inventoryConversationModel{query: "Chain of Thought", answer: "I found your Chain of Thought book."}
	resolver.providers.ConversationModel = model
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	seedRealtimeVoiceLoopAsset(t, store, realtimeVoiceInvestigationAsset("book-id", "Chain of Thought", asset.KindItem, ""), "audit-book")
	response := realtimeVoiceInvestigationCompletedResponse(runRealtimeVoiceProductionEntrypoint(t, application))
	if response == nil || response.SpokenResponse != model.answer || len(response.Artifacts) != 1 {
		t.Fatalf("natural title or independent card rejected: %+v", response)
	}
}

type interruptedConversationSearch struct {
	ports.AssetSearchRepository
	failed bool
}

func (s *interruptedConversationSearch) SearchAssets(ctx context.Context, t tenant.ID, i []inventory.InventoryID, p ports.AssetSearchPageRequest) ([]ports.AssetSearchResult, error) {
	if !s.failed {
		s.failed = true
		return nil, errors.New("temporary database failure with private diagnostic details")
	}
	return s.AssetSearchRepository.SearchAssets(ctx, t, i, p)
}
func TestRealtimeConversationRecoversFromReadFailure(t *testing.T) {
	resolver := successfulRealtimeVoiceResolver()
	model := &inventoryConversationModel{}
	resolver.providers.ConversationModel = model
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	seedRealtimeVoiceLoopAsset(t, store, realtimeVoiceInvestigationAsset("chemical-id", "Acetone", asset.KindItem, ""), "audit-chemical-retry")
	application.search = &interruptedConversationSearch{AssetSearchRepository: application.search}
	response := realtimeVoiceInvestigationCompletedResponse(runRealtimeVoiceProductionEntrypoint(t, application))
	if response == nil || model.calls != 3 || len(response.Artifacts) != 1 {
		t.Fatalf("model could not recover after read error: calls=%d response=%+v", model.calls, response)
	}
}

func TestDirectlyInjectedNativeProviderUsesConversationLoop(t *testing.T) {
	resolver := successfulRealtimeVoiceResolver()
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	seedRealtimeVoiceLoopAsset(t, store, realtimeVoiceInvestigationAsset("acetone", "Acetone", asset.KindItem, ""), "audit-acetone")
	native := &inventoryConversationModel{}
	application = application.WithRealtimeVoiceProviders(resolver.providers.SpeechToText, native, resolver.providers.TextToSpeech)
	response := realtimeVoiceInvestigationCompletedResponse(runRealtimeVoiceProductionEntrypoint(t, application))
	if response == nil || native.calls != 2 || response.SpokenResponse != "Yes, you have chemicals, including Acetone." {
		t.Fatalf("direct injection used obsolete inference: %+v calls=%d", response, native.calls)
	}
}

type presentingConversationModel struct{ inventoryConversationModel }

func (m *presentingConversationModel) Converse(ctx context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	canAnswer := false
	for _, tool := range input.Tools {
		if tool.Name == "present_answer" && tool.ResponseTool {
			canAnswer = true
		}
	}
	if !canAnswer {
		return ports.ConversationModelTurn{}, errors.New("no declared response tool for structured provider output")
	}
	turn, err := m.inventoryConversationModel.Converse(ctx, input)
	if err != nil || turn.Answer == nil {
		return turn, err
	}
	return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{
		{ID: "present", Name: "present_answer", Arguments: map[string]any{"spoken": turn.Answer.Spoken, "display": turn.Answer.Display, "assetIds": turn.Answer.AssetIDs}},
		{ID: "must-not-follow-answer", Name: RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "unrelated"}},
	}}, nil
}
func TestModelCanFinishWithNaturalSpeechAndIndependentCardsThroughTool(t *testing.T) {
	resolver := successfulRealtimeVoiceResolver()
	model := &presentingConversationModel{}
	resolver.providers.ConversationModel = model
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	item := realtimeVoiceInvestigationAsset("acetone", "Acetone", asset.KindItem, "")
	seedRealtimeVoiceLoopAsset(t, store, item, "audit-acetone")
	events := runRealtimeVoiceProductionEntrypoint(t, application)
	response := realtimeVoiceInvestigationCompletedResponse(events)
	if response == nil || response.SpokenResponse != "Yes, you have chemicals, including Acetone." || model.calls != 2 || len(response.Artifacts) != 1 || response.Artifacts[0].AssetID != item.ID {
		t.Fatalf("native presentation did not finish with evidence: %+v calls=%d", response, model.calls)
	}
	for _, event := range events {
		if event.ToolCallID == "must-not-follow-answer" {
			t.Fatal("answer did not stop remaining tool work")
		}
	}
}
