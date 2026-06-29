package httpserver

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceQueryFinalizesAfterDuplicateToolCallWithoutReexecuting(t *testing.T) {
	t.Parallel()

	language := &duplicateToolCallLanguageModel{}
	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"tools-id", "voice-session-id", "response-id"},
	}, fakeSpeechToText{transcript: "Where are my tools?"}, language, fakeTextToSpeech{
		chunks: [][]byte{[]byte("spoken-audio")},
	})
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "container", "Tools", "")

	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	events := runRealtimeVoiceQuestion(t, server.URL, "tenant-home", "inventory-home", "user-1")
	if started := countRealtimeEvents(events, "tool.call.started"); started != 1 {
		t.Fatalf("expected duplicate tool call not to execute twice, got %d starts in %+v", started, events)
	}
	failed := findRealtimeEvent(t, events, "tool.call.failed")
	if failed["code"] != "duplicate_tool_request" {
		t.Fatalf("expected duplicate tool request event, got %+v", failed)
	}
	final := findRealtimeEvent(t, events, "assistant.response.completed")
	response, ok := final["response"].(map[string]any)
	if !ok {
		t.Fatalf("expected structured response, got %+v", final)
	}
	if response["spokenResponse"] != "I found Tools." {
		t.Fatalf("unexpected spoken response: %+v", response)
	}
	if !language.finalizationWithoutTools {
		t.Fatalf("expected duplicate handling to request a finalization-only turn without tool catalog")
	}
}

func TestRealtimeVoiceQueryAllowsMultipleDistinctToolCalls(t *testing.T) {
	t.Parallel()

	language := &multipleDistinctToolCallsLanguageModel{}
	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"tools-id", "bottle-id", "voice-session-id", "response-id"},
	}, fakeSpeechToText{transcript: "Where are my tools and water bottle?"}, language, fakeTextToSpeech{
		chunks: [][]byte{[]byte("spoken-audio")},
	})
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "container", "Tools", "")
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "item", "Water bottle", "")

	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	events := runRealtimeVoiceQuestion(t, server.URL, "tenant-home", "inventory-home", "user-1")
	if started := countRealtimeEvents(events, "tool.call.started"); started != 2 {
		t.Fatalf("expected two distinct tool calls to execute, got %d starts in %+v", started, events)
	}
	if len(language.lastToolResults) != 2 || !strings.Contains(language.lastToolResults[0], "Tools") || !strings.Contains(language.lastToolResults[1], "Water bottle") {
		t.Fatalf("expected both distinct tool results, got %+v", language.lastToolResults)
	}
}

func TestRealtimeVoiceQueryAllowsDistinctToolCallAfterToolResult(t *testing.T) {
	t.Parallel()

	language := &acrossTurnDistinctToolCallsLanguageModel{}
	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"tools-id", "bottle-id", "voice-session-id", "response-id"},
	}, fakeSpeechToText{transcript: "Where are my tools and water bottle?"}, language, fakeTextToSpeech{
		chunks: [][]byte{[]byte("spoken-audio")},
	})
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "container", "Tools", "")
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "item", "Water bottle", "")

	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	events := runRealtimeVoiceQuestion(t, server.URL, "tenant-home", "inventory-home", "user-1")
	if started := countRealtimeEvents(events, "tool.call.started"); started != 2 {
		t.Fatalf("expected two across-turn tool calls to execute, got %d starts in %+v", started, events)
	}
	if !language.secondTurnHadTools {
		t.Fatalf("expected tool catalog to remain available for a distinct across-turn tool call")
	}
}

func TestRealtimeVoiceQuerySkipsDuplicateButExecutesDistinctCallInSameTurn(t *testing.T) {
	t.Parallel()

	language := &duplicateAndDistinctToolCallLanguageModel{}
	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"tools-id", "bottle-id", "voice-session-id", "response-id"},
	}, fakeSpeechToText{transcript: "Where are my tools and water bottle?"}, language, fakeTextToSpeech{
		chunks: [][]byte{[]byte("spoken-audio")},
	})
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "container", "Tools", "")
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "item", "Water bottle", "")

	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	events := runRealtimeVoiceQuestion(t, server.URL, "tenant-home", "inventory-home", "user-1")
	if started := countRealtimeEvents(events, "tool.call.started"); started != 2 {
		t.Fatalf("expected duplicate skipped and distinct call executed, got %d starts in %+v", started, events)
	}
	failed := findRealtimeEvent(t, events, "tool.call.failed")
	if failed["code"] != "duplicate_tool_request" {
		t.Fatalf("expected duplicate tool request event, got %+v", failed)
	}
	if !language.finalizationOnly {
		t.Fatalf("expected duplicate handling to request explicit finalization-only turn")
	}
}

type duplicateToolCallLanguageModel struct {
	turns                    int
	finalizationWithoutTools bool
}

func (m *duplicateToolCallLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	m.turns++
	if m.turns == 1 {
		return ports.LanguageInferenceTurn{ToolCalls: []ports.AgentToolCall{duplicateSearchToolCall()}}, nil
	}
	if input.FinalOnly && len(input.Tools) == 0 && len(input.ToolResults) >= 1 {
		m.finalizationWithoutTools = true
		return ports.LanguageInferenceTurn{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindAnswer,
				SpokenResponse:  "I found Tools.",
				DisplayResponse: "I found Tools.",
			},
		}, nil
	}
	return ports.LanguageInferenceTurn{ToolCalls: []ports.AgentToolCall{duplicateSearchToolCall()}}, nil
}

func duplicateSearchToolCall() ports.AgentToolCall {
	return ports.AgentToolCall{
		ID:        "repeat-search-tools",
		Name:      "search_authorized_assets",
		Arguments: map[string]any{"query": "tools"},
	}
}

type acrossTurnDistinctToolCallsLanguageModel struct {
	secondTurnHadTools bool
}

func (m *acrossTurnDistinctToolCallsLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	switch len(input.ToolResults) {
	case 0:
		return ports.LanguageInferenceTurn{ToolCalls: []ports.AgentToolCall{duplicateSearchToolCall()}}, nil
	case 1:
		m.secondTurnHadTools = len(input.Tools) > 0 && !input.FinalOnly
		return ports.LanguageInferenceTurn{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-bottle",
				Name:      "search_authorized_assets",
				Arguments: map[string]any{"query": "water bottle"},
			}},
		}, nil
	default:
		return ports.LanguageInferenceTurn{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindAnswer,
				SpokenResponse:  "I found your tools and water bottle.",
				DisplayResponse: "I found your tools and water bottle.",
			},
		}, nil
	}
}

type duplicateAndDistinctToolCallLanguageModel struct {
	turns            int
	finalizationOnly bool
}

func (m *duplicateAndDistinctToolCallLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	m.turns++
	if m.turns == 1 {
		return ports.LanguageInferenceTurn{ToolCalls: []ports.AgentToolCall{duplicateSearchToolCall()}}, nil
	}
	if input.FinalOnly {
		m.finalizationOnly = true
		return ports.LanguageInferenceTurn{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindAnswer,
				SpokenResponse:  "I found your tools and water bottle.",
				DisplayResponse: "I found your tools and water bottle.",
			},
		}, nil
	}
	return ports.LanguageInferenceTurn{
		ToolCalls: []ports.AgentToolCall{
			duplicateSearchToolCall(),
			{ID: "search-bottle", Name: "search_authorized_assets", Arguments: map[string]any{"query": "water bottle"}},
		},
	}, nil
}

type multipleDistinctToolCallsLanguageModel struct {
	lastToolResults []string
}

func (m *multipleDistinctToolCallsLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	if len(input.ToolResults) == 0 {
		return ports.LanguageInferenceTurn{
			ToolCalls: []ports.AgentToolCall{
				{ID: "search-tools", Name: "search_authorized_assets", Arguments: map[string]any{"query": "tools"}},
				{ID: "search-bottle", Name: "search_authorized_assets", Arguments: map[string]any{"query": "water bottle"}},
			},
		}, nil
	}
	m.lastToolResults = []string{}
	for _, result := range input.ToolResults {
		m.lastToolResults = append(m.lastToolResults, result.Content)
	}
	return ports.LanguageInferenceTurn{
		Final: &ports.StructuredAgentResponse{
			Kind:            ports.StructuredAgentResponseKindAnswer,
			SpokenResponse:  "I found your tools and water bottle.",
			DisplayResponse: "I found your tools and water bottle.",
		},
	}, nil
}
