package app

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceToolCompletionMarksNoVisibleMatchBlandly(t *testing.T) {
	t.Parallel()

	tts := &resolvedTextToSpeech{}
	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-missing-drill",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "drill"},
			}},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindClarification,
				SpokenResponse:  "I could not find a visible drill. Can you describe it another way?",
				DisplayResponse: "I could not find a visible drill. Can you describe it another way?",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Where is my drill?"}
	resolver.providers.LanguageInference = language
	resolver.providers.TextToSpeech = tts
	application, _ := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	var events []RealtimeVoiceEvent
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %T %[1]v", err)
	}

	var completed *RealtimeVoiceEvent
	for index := range events {
		if events[index].Type == RealtimeVoiceEventToolCallCompleted {
			completed = &events[index]
			break
		}
	}
	if completed == nil {
		t.Fatalf("expected tool completion event, got %+v", events)
	}
	if completed.Status != "no_visible_match" || completed.Detail != "" || completed.Message != "" {
		t.Fatalf("expected bland no-match completion status, got %+v", completed)
	}
}

func TestRealtimeVoiceToolCompletionStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result ports.AgentToolResult
		want   string
	}{
		{
			name: "empty search",
			result: ports.AgentToolResult{
				Name:    RealtimeVoiceToolSearchAuthorizedAssets,
				Content: `{"tool":"search_authorized_assets","query":"drill","count":0,"items":[]}`,
			},
			want: "no_visible_match",
		},
		{
			name: "empty list",
			result: ports.AgentToolResult{
				Name:    RealtimeVoiceToolListAuthorizedAssets,
				Content: `{"tool":"list_authorized_assets","count":0,"items":[]}`,
			},
			want: "no_visible_match",
		},
		{
			name: "nonempty search",
			result: ports.AgentToolResult{
				Name:    RealtimeVoiceToolSearchAuthorizedAssets,
				Content: `{"tool":"search_authorized_assets","query":"drill","count":1,"items":[{"assetId":"asset-1","title":"Drill","kind":"item"}]}`,
			},
			want: "completed",
		},
		{
			name: "malformed asset result",
			result: ports.AgentToolResult{
				Name:    RealtimeVoiceToolSearchAuthorizedAssets,
				Content: `{`,
			},
			want: "completed",
		},
		{
			name: "non asset result",
			result: ports.AgentToolResult{
				Name:    RealtimeVoiceToolListAssetAuditHistory,
				Content: `{"tool":"list_asset_audit_history","count":0,"items":[]}`,
			},
			want: "completed",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := realtimeVoiceToolCompletionStatus(test.result); got != test.want {
				t.Fatalf("expected %q, got %q", test.want, got)
			}
		})
	}
}
