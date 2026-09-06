package app

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type conversationFollowUpSpeech struct{ calls int }

func (s *conversationFollowUpSpeech) Transcribe(context.Context, ports.SpeechToTextInput) (ports.SpeechToTextResult, error) {
	s.calls++
	text := "Do I have a Drill?"
	if s.calls == 2 {
		text = "Where was that?"
	}
	return ports.SpeechToTextResult{Transcript: text}, nil
}

type rememberingConversationModel struct {
	calls               int
	evidence, signature bool
}

func (m *rememberingConversationModel) Converse(_ context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	m.calls++
	if m.calls == 1 {
		return ports.ConversationModelTurn{ProviderState: []byte("signed-search"), ToolCalls: []ports.AgentToolCall{{ID: "find", Name: RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "Drill"}}}}, nil
	}
	if m.calls == 2 {
		return ports.ConversationModelTurn{Text: "You have a Drill."}, nil
	}
	for _, message := range input.Messages {
		if string(message.ProviderState) == "signed-search" {
			m.signature = true
		}
		for _, result := range message.ToolResults {
			var output realtimeVoiceAssetToolOutput
			_ = json.Unmarshal([]byte(result.Content), &output)
			for _, item := range output.Items {
				if item.AssetID == "existing-drill" {
					m.evidence = true
				}
			}
		}
	}
	return ports.ConversationModelTurn{Text: "That Drill has no recorded location."}, nil
}
func TestModelLedVoiceFollowUpRetainsNativeToolHistory(t *testing.T) {
	resolver := successfulRealtimeVoiceResolver()
	model := &rememberingConversationModel{}
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
	if model.calls != 3 || !model.evidence || !model.signature {
		t.Fatalf("follow-up lost prior evidence or native continuation: %+v", model)
	}
	model.evidence, model.signature = false, false
	fresh, err := application.StartRealtimeVoiceSession(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: fresh, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if model.evidence || model.signature {
		t.Fatal("new session inherited private conversation state")
	}
}

func TestModelLedHistoryCannotBeReboundToAnotherSession(t *testing.T) {
	resolver := successfulRealtimeVoiceResolver()
	model := &rememberingConversationModel{}
	resolver.providers.ConversationModel = model
	sessions := newFakeRealtimeSessionRepository()
	application := newRealtimeVoiceResolutionTestAppWithSessions(t, resolver, sessions)
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatal(err)
	}
	other, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatal(err)
	}
	session.ID = other.ID
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error { return nil })
	if err == nil || model.calls != 0 {
		t.Fatalf("session state was rebound: calls=%d err=%v", model.calls, err)
	}
	record, found, err := sessions.RealtimeSessionByID(context.Background(), other.TenantID, other.InventoryID, other.ID)
	if err != nil || !found || record.State != ports.RealtimeSessionStateStarted {
		t.Fatalf("rebound request altered the other session: %+v err=%v", record, err)
	}
}
