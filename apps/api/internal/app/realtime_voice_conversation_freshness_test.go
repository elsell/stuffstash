package app

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type refreshingConversationModel struct {
	calls            int
	refreshRequested bool
	direct           bool
}

func (m *refreshingConversationModel) Converse(_ context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	m.calls++
	if m.direct && m.calls == 3 {
		return ports.ConversationModelTurn{Answer: &ports.ConversationAnswer{Spoken: "Old Drill", Display: "Old Drill", AssetIDs: []string{"existing-drill"}}}, nil
	}
	if m.calls == 1 {
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "find", Name: RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "Drill"}}}}, nil
	}
	if m.calls == 4 {
		for _, message := range input.Messages {
			for _, result := range message.ToolResults {
				if strings.Contains(result.Content, "refreshAssetIds") && strings.Contains(result.Content, "existing-drill") {
					m.refreshRequested = true
				}
			}
		}
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "refresh", Name: RealtimeVoiceToolGetAssetDetail, Arguments: map[string]any{"assetId": "existing-drill"}}}}, nil
	}
	return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "answer-" + strconv.Itoa(m.calls), Name: realtimeConversationPresentTool, Arguments: map[string]any{"spoken": "You have a Drill.", "display": "You have a Drill.", "assetIds": []string{"existing-drill"}}}}}, nil
}
func TestConversationFollowUpRequiresFreshItemEvidence(t *testing.T) {
	resolver := successfulRealtimeVoiceResolver()
	model := &refreshingConversationModel{}
	resolver.providers.ConversationModel = model
	resolver.providers.SpeechToText = &conversationFollowUpSpeech{}
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	seedRealtimeVoiceLoopAsset(t, store, realtimeVoiceInvestigationAsset("existing-drill", "Drill", asset.KindItem, ""), "audit-drill")
	input := defaultRealtimeVoiceSessionInput()
	input.ConversationContinuity = true
	session, err := application.StartRealtimeVoiceSession(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error { return nil }); err != nil {
			t.Fatal(err)
		}
	}
	if model.calls != 5 || !model.refreshRequested {
		t.Fatalf("follow-up delivered prior facts without refreshing: calls=%d requested=%v", model.calls, model.refreshRequested)
	}
}

func TestConversationDirectAnswerCannotDeliverStaleCard(t *testing.T) {
	resolver := successfulRealtimeVoiceResolver()
	model := &refreshingConversationModel{direct: true}
	resolver.providers.ConversationModel = model
	resolver.providers.SpeechToText = &conversationFollowUpSpeech{}
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	seedRealtimeVoiceLoopAsset(t, store, realtimeVoiceInvestigationAsset("existing-drill", "Drill", asset.KindItem, ""), "audit-drill")
	input := defaultRealtimeVoiceSessionInput()
	input.ConversationContinuity = true
	session, err := application.StartRealtimeVoiceSession(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	query := RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}
	if err := application.RunRealtimeVoiceQuery(context.Background(), query, func(RealtimeVoiceEvent) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if err := application.RunRealtimeVoiceQuery(context.Background(), query, func(RealtimeVoiceEvent) error { return nil }); err == nil {
		t.Fatal("direct answer delivered stale card")
	}
}
