package httpserver

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceQueryRejectsMalformedInvestigationFromProvider(t *testing.T) {
	t.Parallel()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"voice-session-id", "response-id"},
	}, fakeSpeechToText{transcript: "Where is my water bottle?"}, finalResponseLanguageModel{
		final: ports.StructuredAgentResponse{
			Kind:            ports.StructuredAgentResponseKind("raw_provider_dump"),
			SpokenResponse:  strings.Repeat("x", 501),
			DisplayResponse: strings.Repeat("x", 1001),
		},
	}, fakeTextToSpeech{chunks: [][]byte{[]byte("spoken-audio")}})

	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	events := runRealtimeVoiceQuestionUntil(t, server.URL, "tenant-home", "inventory-home", "user-1", "session.failed")
	failed := findRealtimeEvent(t, events, "session.failed")
	if failed["code"] != "invalid_request" {
		t.Fatalf("expected malformed typed investigation to fail safely, got %+v", failed)
	}
}
