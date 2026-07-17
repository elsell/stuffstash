package app

import (
	"context"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceDerivesEffectiveTranscriptForClarificationFollowUp(t *testing.T) {
	t.Parallel()

	language := realtimeVoiceUnsupportedScript()
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

	language := realtimeVoiceUnsupportedScript()
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
			{Role: ports.AgentConversationRoleUser, Text: "Move my water bottle. bearer abc/def== https://provider.example.test/raw raw model response: {\"error\":\"provider internals\"}", Kind: "apiKey: should-redact"},
			{Role: ports.AgentConversationRoleAssistant, Kind: string(ports.StructuredAgentResponseKindClarification), Text: "Where should I move it? providerSessionId: live-1 endpoint URL wss://provider.example.test/session"},
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
	expected := []struct {
		role ports.AgentConversationRole
		kind string
		text string
	}{
		{role: ports.AgentConversationRoleUser, text: "Older two."},
		{role: ports.AgentConversationRoleAssistant, kind: string(ports.StructuredAgentResponseKindClarification), text: "Older clarification."},
		{role: ports.AgentConversationRoleUser, kind: "[redacted-key][redacted]", text: "Move my water bottle. [redacted-bearer] [redacted] [redacted-url] [redacted]"},
		{role: ports.AgentConversationRoleAssistant, kind: string(ports.StructuredAgentResponseKindClarification), text: "Where should I move it? [redacted-provider-session][redacted] endpoint URL [redacted-url]"},
	}
	for index, expectedTurn := range expected {
		if turns[index].Role != expectedTurn.role || turns[index].Kind != expectedTurn.kind || turns[index].Text != expectedTurn.text {
			t.Fatalf("unexpected safe context turn %d: got %+v want %+v", index, turns[index], expectedTurn)
		}
	}
	if turns[4].Role != ports.AgentConversationRoleUser || !strings.HasPrefix(turns[4].Text, "garage garage") || !strings.HasSuffix(turns[4].Text, " ...") {
		t.Fatalf("expected newest long user turn to be retained and bounded, got %+v", turns[4])
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
	for _, unsafe := range []string{"abc/def", "should-redact", "providerSessionId", "live-1", "https://provider.example.test", "wss://provider.example.test", "raw model response", "provider internals", "ignored bad role", "Old one", "Old clarification"} {
		if strings.Contains(joined, unsafe) {
			t.Fatalf("unsafe or out-of-window conversation context reached provider: %q", joined)
		}
	}
	for _, expected := range []string{"[redacted-bearer]", "[redacted-key]", "[redacted-provider-session]", "[redacted-url]"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected redacted marker %q in provider context, got %q", expected, joined)
		}
	}
	if len(language.seenTranscripts) == 0 {
		t.Fatalf("expected language provider to receive effective transcript")
	}
	effectiveTranscript := language.seenTranscripts[0]
	for _, unsafe := range []string{"abc/def", "should-redact", "providerSessionId", "live-1", "https://provider.example.test", "wss://provider.example.test", "raw model response", "provider internals"} {
		if strings.Contains(effectiveTranscript, unsafe) {
			t.Fatalf("unsafe conversation text reached effective transcript: %q", effectiveTranscript)
		}
	}
	if !strings.Contains(effectiveTranscript, "Move my water bottle. [redacted-bearer] [redacted] [redacted-url] [redacted] Follow-up answer: Kitchen.") {
		t.Fatalf("expected safe prior intent in effective transcript, got %q", effectiveTranscript)
	}
}

func TestRealtimeVoiceDerivesEffectiveTranscriptForReadClarificationFollowUp(t *testing.T) {
	t.Parallel()

	language := realtimeVoiceUnsupportedScript()
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

	language := realtimeVoiceUnsupportedScript()
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

func realtimeVoiceUnsupportedScript() *scriptedRealtimeLanguageInference {
	step := agentmodel.InvestigationStep{
		Decision: agentmodel.InvestigationDecisionFinish,
		Intent: agentmodel.Intent{
			Kind:      agentmodel.IntentKindUnsupported,
			Operation: agentmodel.OperationUnsupported,
		},
		Resolutions: []agentmodel.Resolution{{
			ReferenceKey: agentmodel.SemanticReferenceSubject,
			Status:       agentmodel.ResolutionUnsupported,
		}},
	}
	return &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{{Investigation: &step}}}
}
