package httpserver

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceQueryStreamsSanitizedLanguageFailureDiagnosticBeforeFailure(t *testing.T) {
	t.Parallel()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"water-bottle-id", "voice-session-id", "tool-call-id"},
	}, fakeSpeechToText{transcript: "Move my water bottle to the kitchen."}, lateFailingLanguageModel{}, fakeTextToSpeech{chunks: [][]byte{[]byte("spoken-audio")}})
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "item", "Water bottle", "")
	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	events := runRealtimeVoiceQuestionUntilWithStart(t, server.URL, map[string]any{
		"type":                 "session.start",
		"seq":                  1,
		"tenantId":             "tenant-home",
		"inventoryId":          "inventory-home",
		"source":               "mobile_voice",
		"developerDiagnostics": true,
		"inputAudio":           map[string]any{"mimeType": "audio/mp4", "sampleRate": 44100, "channels": 1},
		"outputAudio":          map[string]any{"mimeTypes": []string{"audio/mpeg"}},
	}, "user-1", "session.failed")

	diagnosticIndex := -1
	failedIndex := -1
	for index, event := range events {
		switch event["type"] {
		case "agent.diagnostic":
			if event["message"] == "Language provider failed" {
				diagnosticIndex = index
			}
		case "session.failed":
			failedIndex = index
		}
	}
	if diagnosticIndex < 0 || failedIndex < 0 || diagnosticIndex > failedIndex {
		t.Fatalf("expected language failure diagnostic before session.failed, got %+v", events)
	}
	detail, _ := events[diagnosticIndex]["detail"].(string)
	for _, required := range []string{`"turn": 2`, `"previousTurns": 1`, `"toolResultCount": 1`, "search_authorized_assets", "provider_http_status_429"} {
		if !strings.Contains(detail, required) {
			t.Fatalf("expected diagnostic detail to include %q, got %s", required, detail)
		}
	}
	if strings.Contains(detail, "provider.invalid") || strings.Contains(detail, "should-not-leak") || strings.Contains(strings.ToLower(detail), "bearer ") {
		t.Fatalf("diagnostic leaked unsafe provider detail: %s", detail)
	}
}

type lateFailingLanguageModel struct{}

func (lateFailingLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	if len(input.ToolResults) == 0 {
		return ports.LanguageInferenceTurn{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "tool-call-id",
				Name:      "search_authorized_assets",
				Arguments: map[string]any{"query": "water bottle"},
			}},
		}, nil
	}
	return ports.LanguageInferenceTurn{}, safeHTTPStatusLanguageError{}
}

type safeHTTPStatusLanguageError struct{}

func (e safeHTTPStatusLanguageError) Error() string {
	return "raw provider response from https://provider.invalid bearer should-not-leak"
}

func (e safeHTTPStatusLanguageError) SafeRealtimeVoiceDiagnostic() string {
	return "provider_http_status_429"
}
