package httpserver

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceQueryRecoversMalformedFinalResponseFromProvider(t *testing.T) {
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

	events := runRealtimeVoiceQuestionUntil(t, server.URL, "tenant-home", "inventory-home", "user-1", "session.completed")
	final := findRealtimeEvent(t, events, "assistant.response.completed")
	response, ok := final["response"].(map[string]any)
	if !ok {
		t.Fatalf("expected structured response, got %+v", final)
	}
	if response["kind"] != "safe_failure" || response["spokenResponse"] != "I could not finish that voice request safely. Please try again with a little more detail." {
		t.Fatalf("expected safe recovery response for malformed final response, got %+v", response)
	}
}

func TestRealtimeVoiceQueryReturnsSafeRecoveryForUnexpectedToolArguments(t *testing.T) {
	t.Parallel()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"voice-session-id", "tool-call-1", "tool-call-2", "tool-call-3", "tool-call-4", "response-id"},
	}, fakeSpeechToText{transcript: "What items do I have?"}, unexpectedToolArgumentLanguageModel{}, fakeTextToSpeech{chunks: [][]byte{[]byte("spoken-audio")}})

	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	events := runRealtimeVoiceQuestionUntil(t, server.URL, "tenant-home", "inventory-home", "user-1", "session.completed")
	failedTool := findRealtimeEvent(t, events, "tool.call.failed")
	if failedTool["code"] != "invalid_tool_request" {
		t.Fatalf("expected invalid tool request event, got %+v", failedTool)
	}
	final := findRealtimeEvent(t, events, "assistant.response.completed")
	response, ok := final["response"].(map[string]any)
	if !ok {
		t.Fatalf("expected structured response, got %+v", final)
	}
	if response["kind"] != "safe_failure" {
		t.Fatalf("expected safe recovery response, got %+v", response)
	}
}
