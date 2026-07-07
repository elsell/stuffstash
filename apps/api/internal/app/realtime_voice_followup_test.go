package app

import (
	"context"
	"strings"
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

func TestRealtimeVoiceSendsOnlyBoundedSafeConversationContextToLanguagePort(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindAnswer,
				SpokenResponse:  "I understand the safe follow-up context.",
				DisplayResponse: "I understand the safe follow-up context.",
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
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{
		Session:                    session,
		AudioChunks:                [][]byte{[]byte("audio")},
		ContinueAfterClarification: true,
		ConversationTurns: []ports.AgentConversationTurn{
			{Role: ports.AgentConversationRoleUser, Text: "Old one."},
			{Role: ports.AgentConversationRoleAssistant, Kind: string(ports.StructuredAgentResponseKindClarification), Text: "Old clarification."},
			{Role: ports.AgentConversationRoleUser, Text: "Older two."},
			{Role: ports.AgentConversationRoleAssistant, Kind: string(ports.StructuredAgentResponseKindClarification), Text: "Older clarification."},
			{Role: ports.AgentConversationRoleUser, Text: "Move my water bottle. bearer abc/def==", Kind: "apiKey: should-redact"},
			{Role: ports.AgentConversationRoleAssistant, Kind: string(ports.StructuredAgentResponseKindClarification), Text: "Where should I move it? providerSessionId: live-1"},
			{Role: ports.AgentConversationRole("tool"), Text: "ignored bad role", Kind: "clarification"},
			{Role: ports.AgentConversationRoleUser, Text: strings.Repeat("garage ", 120)},
		},
	}, func(RealtimeVoiceEvent) error {
		return nil
	})
	if err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}
	if len(language.seenConversationTurns) == 0 {
		t.Fatalf("expected language provider to receive conversation turns")
	}
	turns := language.seenConversationTurns[0]
	if len(turns) != 5 {
		t.Fatalf("expected last six input turns minus invalid role to be sent, got %+v", turns)
	}
	joined := ""
	for _, turn := range turns {
		if turn.Role != ports.AgentConversationRoleUser && turn.Role != ports.AgentConversationRoleAssistant {
			t.Fatalf("unexpected unsafe role sent to language provider: %+v", turn)
		}
		if len(turn.Text) > 505 || len(turn.Kind) > 90 {
			t.Fatalf("expected bounded context turn, got %+v", turn)
		}
		joined += string(turn.Role) + " " + turn.Kind + " " + turn.Text + "\n"
	}
	for _, unsafe := range []string{"abc/def", "should-redact", "providerSessionId", "live-1", "ignored bad role", "Old one"} {
		if strings.Contains(joined, unsafe) {
			t.Fatalf("unsafe or out-of-window conversation context reached provider: %q", joined)
		}
	}
	for _, expected := range []string{"[redacted-bearer]", "[redacted-key]", "[redacted-provider-session]"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected redacted marker %q in provider context, got %q", expected, joined)
		}
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
