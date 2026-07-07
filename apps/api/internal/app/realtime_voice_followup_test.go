package app

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceDerivesEffectiveTranscriptForClarificationFollowUp(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindAnswer,
				SpokenResponse:  "I understand the follow-up.",
				DisplayResponse: "I understand the follow-up.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Kitchen."}
	resolver.providers.LanguageInference = language
	application, _ := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{
		Session:                    session,
		AudioChunks:                [][]byte{[]byte("audio")},
		ContinueAfterClarification: true,
		ConversationTurns: []ports.AgentConversationTurn{
			{Role: ports.AgentConversationRoleUser, Text: "Move my water bottle."},
			{Role: ports.AgentConversationRoleAssistant, Kind: string(ports.StructuredAgentResponseKindClarification), Text: "Where should I move it?"},
		},
	}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}
	if len(language.seenTranscripts) == 0 || language.seenTranscripts[0] != "Move my water bottle. Follow-up answer: Kitchen." {
		t.Fatalf("expected model to receive effective follow-up transcript, got %+v", language.seenTranscripts)
	}
	if len(events) == 0 || events[0].Type != RealtimeVoiceEventTranscriptFinal || events[0].Text != "Kitchen." {
		t.Fatalf("expected client transcript event to keep literal follow-up transcript, got %+v", events)
	}
}

func TestRealtimeVoiceDerivesEffectiveTranscriptForReadClarificationFollowUp(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindAnswer,
				SpokenResponse:  "I understand the read follow-up.",
				DisplayResponse: "I understand the read follow-up.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Water bottle."}
	resolver.providers.LanguageInference = language
	application, _ := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{
		Session:                    session,
		AudioChunks:                [][]byte{[]byte("audio")},
		ContinueAfterClarification: true,
		ConversationTurns: []ports.AgentConversationTurn{
			{Role: ports.AgentConversationRoleUser, Text: "Where is it?"},
			{Role: ports.AgentConversationRoleAssistant, Kind: string(ports.StructuredAgentResponseKindClarification), Text: "Which item should I find?"},
		},
	}, func(RealtimeVoiceEvent) error {
		return nil
	})
	if err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}
	if len(language.seenTranscripts) == 0 || language.seenTranscripts[0] != "Where is it? Follow-up answer: Water bottle." {
		t.Fatalf("expected model to receive effective read follow-up transcript, got %+v", language.seenTranscripts)
	}
}

func TestRealtimeVoiceDerivesEffectiveTranscriptForReturnClarificationFollowUp(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindAnswer,
				SpokenResponse:  "I understand the return follow-up.",
				DisplayResponse: "I understand the return follow-up.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Drill."}
	resolver.providers.LanguageInference = language
	application, _ := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{
		Session:                    session,
		AudioChunks:                [][]byte{[]byte("audio")},
		ContinueAfterClarification: true,
		ConversationTurns: []ports.AgentConversationTurn{
			{Role: ports.AgentConversationRoleUser, Text: "Return it."},
			{Role: ports.AgentConversationRoleAssistant, Kind: string(ports.StructuredAgentResponseKindClarification), Text: "Which item should I mark as returned?"},
		},
	}, func(RealtimeVoiceEvent) error {
		return nil
	})
	if err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}
	if len(language.seenTranscripts) == 0 || language.seenTranscripts[0] != "Return it. Follow-up answer: Drill." {
		t.Fatalf("expected model to receive effective return follow-up transcript, got %+v", language.seenTranscripts)
	}
}
