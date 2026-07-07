package httpserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceQueryWebSocketStreamsTranscriptToolResultAndSpeech(t *testing.T) {
	t.Parallel()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"garage-id", "tools-id", "voice-session-id", "tool-call-id", "response-id"},
	}, fakeSpeechToText{transcript: "Where are my tools?"}, scriptedLanguageModel{}, fakeTextToSpeech{
		chunks: [][]byte{[]byte("spoken-audio-1"), []byte("spoken-audio-2")},
	})
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "location", "Garage", "")
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "container", "Tools", "garage-id")

	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{
		HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:user-1"}},
	})
	if err != nil {
		t.Fatalf("dial realtime voice websocket: %v", err)
	}
	t.Cleanup(func() { _ = connection.Close(websocket.StatusNormalClosure, "") })

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":                  "session.start",
		"seq":                   1,
		"tenantId":              "tenant-home",
		"inventoryId":           "inventory-home",
		"source":                "mobile_voice",
		"requestedCapabilities": []string{"speech_to_text", "language_inference", "text_to_speech"},
		"inputAudio":            map[string]any{"mimeType": "audio/mp4", "sampleRate": 44100, "channels": 1},
		"outputAudio":           map[string]any{"mimeTypes": []string{"audio/mpeg"}},
	})
	started := readRealtimeMessage(t, ctx, connection)
	if started["type"] != "session.started" {
		t.Fatalf("expected session.started, got %+v", started)
	}
	sessionID, _ := started["sessionId"].(string)
	if sessionID == "" {
		t.Fatalf("expected server-created session id, got %q", sessionID)
	}

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":         "audio.chunk",
		"seq":          2,
		"sessionId":    sessionID,
		"chunkId":      "chunk-1",
		"audioBase64":  base64.StdEncoding.EncodeToString([]byte("fake-audio")),
		"isFinalChunk": true,
	})
	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":      "audio.end",
		"seq":       3,
		"sessionId": sessionID,
	})

	events := readRealtimeMessagesUntil(t, ctx, connection, "session.completed")
	assertRealtimeEventTypes(t, events,
		"transcript.final",
		"agent.progress",
		"tool.call.started",
		"tool.call.completed",
		"assistant.response.started",
		"assistant.response.completed",
		"tts.audio.started",
		"tts.audio.chunk",
		"tts.audio.completed",
		"session.completed",
	)
	assertNoRealtimeEventType(t, events, "assistant.response.delta")
	assertSafeRealtimeEvents(t, events, []string{"fake-audio", "garage-id", "tools-id"})

	final := findRealtimeEvent(t, events, "assistant.response.completed")
	response, ok := final["response"].(map[string]any)
	if !ok {
		t.Fatalf("expected structured response, got %+v", final)
	}
	if response["spokenResponse"] != "Your tools are in Garage." {
		t.Fatalf("unexpected spoken response: %+v", response)
	}
	if response["kind"] != "answer" {
		t.Fatalf("unexpected response kind: %+v", response)
	}

	var speechChunks int
	for _, event := range events {
		if event["type"] == "tts.audio.chunk" {
			speechChunks++
			if event["audioBase64"] == "" {
				t.Fatalf("expected speech chunk payload, got %+v", event)
			}
		}
	}
	if speechChunks != 2 {
		t.Fatalf("expected two streamed speech chunks, got %d", speechChunks)
	}
}

func TestRealtimeVoiceWebSocketAcceptsFollowUpAudioAfterClarification(t *testing.T) {
	t.Parallel()

	language := &scriptedFinalLanguageModel{finals: []ports.StructuredAgentResponse{
		{
			Kind:            ports.StructuredAgentResponseKindClarification,
			SpokenResponse:  "Which item should I update?",
			DisplayResponse: "Which item should I update?",
		},
		{
			Kind:            ports.StructuredAgentResponseKindAnswer,
			SpokenResponse:  "Got it. I will use the office context.",
			DisplayResponse: "Got it. I will use the office context.",
		},
	}}
	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"voice-session-id", "clarification-response-id", "answer-response-id"},
	}, &scriptedSpeechToText{transcripts: []string{"Where should I put it?", "Put it in the office."}}, language, fakeTextToSpeech{chunks: [][]byte{[]byte("spoken-audio")}})

	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{
		HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:user-1"}},
	})
	if err != nil {
		t.Fatalf("dial realtime voice websocket: %v", err)
	}
	t.Cleanup(func() { _ = connection.Close(websocket.StatusNormalClosure, "") })

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":                  "session.start",
		"seq":                   1,
		"tenantId":              "tenant-home",
		"inventoryId":           "inventory-home",
		"source":                "mobile_voice",
		"requestedCapabilities": []string{"speech_to_text", "language_inference", "text_to_speech"},
		"inputAudio":            map[string]any{"mimeType": "audio/mp4", "sampleRate": 44100, "channels": 1},
		"outputAudio":           map[string]any{"mimeTypes": []string{"audio/mpeg"}},
	})
	started := readRealtimeMessage(t, ctx, connection)
	if started["type"] != "session.started" {
		t.Fatalf("expected session.started, got %+v", started)
	}
	sessionID, _ := started["sessionId"].(string)
	writeRealtimeAudioTurn(t, ctx, connection, sessionID, 2, "turn-1")
	firstTurn := readRealtimeMessagesUntil(t, ctx, connection, "session.completed")
	firstResponse := findRealtimeEvent(t, firstTurn, "assistant.response.completed")
	firstPayload, _ := firstResponse["response"].(map[string]any)
	if firstPayload["kind"] != "clarification" {
		t.Fatalf("expected first turn to request clarification, got %+v", firstPayload)
	}

	writeRealtimeAudioTurn(t, ctx, connection, sessionID, 4, "turn-2")
	secondTurn := readRealtimeMessagesUntil(t, ctx, connection, "session.completed")
	secondResponse := findRealtimeEvent(t, secondTurn, "assistant.response.completed")
	secondPayload, _ := secondResponse["response"].(map[string]any)
	if secondPayload["kind"] != "answer" || secondPayload["spokenResponse"] != "Got it. I will use the office context." {
		t.Fatalf("expected follow-up answer on same session, got %+v", secondPayload)
	}
	if len(language.inputs) < 2 || len(language.inputs[1].ConversationTurns) != 2 {
		t.Fatalf("expected follow-up language turn to include prior user and assistant context, got %+v", language.inputs)
	}
	if language.inputs[1].ConversationTurns[0].Role != ports.AgentConversationRoleUser || language.inputs[1].ConversationTurns[0].Text != "Where should I put it?" {
		t.Fatalf("unexpected prior user context: %+v", language.inputs[1].ConversationTurns)
	}
	if language.inputs[1].ConversationTurns[1].Role != ports.AgentConversationRoleAssistant || language.inputs[1].ConversationTurns[1].Kind != string(ports.StructuredAgentResponseKindClarification) || language.inputs[1].ConversationTurns[1].Text != "Which item should I update?" {
		t.Fatalf("unexpected prior assistant context: %+v", language.inputs[1].ConversationTurns)
	}
}

func TestRealtimeVoiceWebSocketFailsSafelyAtClarificationTurnLimit(t *testing.T) {
	t.Parallel()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"voice-session-id", "clarification-1", "clarification-2", "clarification-3"},
	}, &scriptedSpeechToText{transcripts: []string{"Move it.", "That one.", "Over there."}}, &scriptedFinalLanguageModel{finals: []ports.StructuredAgentResponse{
		{Kind: ports.StructuredAgentResponseKindClarification, SpokenResponse: "Which item?", DisplayResponse: "Which item?"},
		{Kind: ports.StructuredAgentResponseKindClarification, SpokenResponse: "Where should it go?", DisplayResponse: "Where should it go?"},
		{Kind: ports.StructuredAgentResponseKindClarification, SpokenResponse: "Please name the destination.", DisplayResponse: "Please name the destination."},
	}}, fakeTextToSpeech{chunks: [][]byte{[]byte("spoken-audio")}})

	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{
		HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:user-1"}},
	})
	if err != nil {
		t.Fatalf("dial realtime voice websocket: %v", err)
	}
	t.Cleanup(func() { _ = connection.Close(websocket.StatusNormalClosure, "") })

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":        "session.start",
		"seq":         1,
		"tenantId":    "tenant-home",
		"inventoryId": "inventory-home",
		"source":      "mobile_voice",
		"inputAudio":  map[string]any{"mimeType": "audio/mp4", "sampleRate": 44100, "channels": 1},
		"outputAudio": map[string]any{"mimeTypes": []string{"audio/mpeg"}},
	})
	started := readRealtimeMessage(t, ctx, connection)
	sessionID, _ := started["sessionId"].(string)
	for turn := 0; turn < maxRealtimeVoiceTurnsPerSession; turn++ {
		writeRealtimeAudioTurn(t, ctx, connection, sessionID, 2+(turn*2), "turn-limit-"+strconv.Itoa(turn+1))
		events := readRealtimeMessagesUntil(t, ctx, connection, "session.completed")
		response := findRealtimeEvent(t, events, "assistant.response.completed")
		payload, _ := response["response"].(map[string]any)
		if payload["kind"] != "clarification" {
			t.Fatalf("expected clarification turn %d, got %+v", turn+1, payload)
		}
	}
	failed := readRealtimeMessage(t, ctx, connection)
	if failed["type"] != "session.failed" || failed["code"] != "clarification_turn_limit" {
		t.Fatalf("expected clarification turn limit failure, got %+v", failed)
	}
}

func TestRealtimeVoiceQueryRejectsUnauthenticatedWebSocket(t *testing.T) {
	t.Parallel()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
	}, fakeSpeechToText{}, scriptedLanguageModel{}, fakeTextToSpeech{})
	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	_, response, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", nil)
	if err == nil {
		t.Fatal("expected unauthenticated websocket dial to fail")
	}
	if response == nil || response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected HTTP 401, got response=%v err=%v", response, err)
	}
}

func TestRealtimeVoiceQueryRejectsWrongInventory(t *testing.T) {
	t.Parallel()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants: []seedTenant{
			{id: "tenant-home", name: "Home", owner: "user-1"},
			{id: "tenant-other", name: "Other", owner: "user-2"},
		},
		inventories: []seedInventory{
			{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"},
			{id: "inventory-other", tenantID: "tenant-other", name: "Other inventory", owner: "user-2"},
		},
		ids: []string{"voice-session-id"},
	}, fakeSpeechToText{}, scriptedLanguageModel{}, fakeTextToSpeech{})
	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{
		HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:user-1"}},
	})
	if err != nil {
		t.Fatalf("dial realtime voice websocket: %v", err)
	}
	t.Cleanup(func() { _ = connection.Close(websocket.StatusNormalClosure, "") })

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":        "session.start",
		"seq":         1,
		"tenantId":    "tenant-other",
		"inventoryId": "inventory-other",
		"source":      "mobile_voice",
		"inputAudio":  map[string]any{"mimeType": "audio/mp4", "sampleRate": 44100, "channels": 1},
		"outputAudio": map[string]any{"mimeTypes": []string{"audio/mpeg"}},
	})
	failed := readRealtimeMessage(t, ctx, connection)
	if failed["type"] != "session.failed" {
		t.Fatalf("expected session.failed, got %+v", failed)
	}
	if failed["code"] != "forbidden" {
		t.Fatalf("expected forbidden, got %+v", failed)
	}
}

func TestRealtimeVoiceQueryRejectsInventoryFromDifferentTenantEvenWhenPrincipalCanViewBoth(t *testing.T) {
	t.Parallel()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants: []seedTenant{
			{id: "tenant-home", name: "Home", owner: "user-1"},
			{id: "tenant-shop", name: "Shop", owner: "user-1"},
		},
		inventories: []seedInventory{
			{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"},
			{id: "inventory-shop", tenantID: "tenant-shop", name: "Shop inventory", owner: "user-1"},
		},
		ids: []string{"voice-session-id"},
	}, fakeSpeechToText{}, scriptedLanguageModel{}, fakeTextToSpeech{})
	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{
		HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:user-1"}},
	})
	if err != nil {
		t.Fatalf("dial realtime voice websocket: %v", err)
	}
	t.Cleanup(func() { _ = connection.Close(websocket.StatusNormalClosure, "") })

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":        "session.start",
		"seq":         1,
		"tenantId":    "tenant-home",
		"inventoryId": "inventory-shop",
		"source":      "mobile_voice",
		"inputAudio":  map[string]any{"mimeType": "audio/mp4", "sampleRate": 44100, "channels": 1},
		"outputAudio": map[string]any{"mimeTypes": []string{"audio/mpeg"}},
	})
	failed := readRealtimeMessage(t, ctx, connection)
	if failed["type"] != "session.failed" {
		t.Fatalf("expected session.failed, got %+v", failed)
	}
	if failed["code"] != "forbidden" {
		t.Fatalf("expected forbidden, got %+v", failed)
	}
}

func TestRealtimeVoiceQueryRejectsDecodedAudioChunkLargerThanLimit(t *testing.T) {
	t.Parallel()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"voice-session-id"},
	}, fakeSpeechToText{}, scriptedLanguageModel{}, fakeTextToSpeech{})
	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{
		HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:user-1"}},
	})
	if err != nil {
		t.Fatalf("dial realtime voice websocket: %v", err)
	}
	t.Cleanup(func() { _ = connection.Close(websocket.StatusNormalClosure, "") })

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":        "session.start",
		"seq":         1,
		"tenantId":    "tenant-home",
		"inventoryId": "inventory-home",
		"source":      "mobile_voice",
		"inputAudio":  map[string]any{"mimeType": "audio/mp4", "sampleRate": 44100, "channels": 1},
		"outputAudio": map[string]any{"mimeTypes": []string{"audio/mpeg"}},
	})
	started := readRealtimeMessage(t, ctx, connection)
	sessionID, _ := started["sessionId"].(string)

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":        "audio.chunk",
		"seq":         2,
		"sessionId":   sessionID,
		"chunkId":     "oversized",
		"audioBase64": base64.StdEncoding.EncodeToString(make([]byte, 512*1024+1)),
	})
	failed := readRealtimeMessage(t, ctx, connection)
	if failed["type"] != "session.failed" {
		t.Fatalf("expected session.failed, got %+v", failed)
	}
	if failed["code"] != "invalid_request" {
		t.Fatalf("expected invalid request, got %+v", failed)
	}
}

func TestRealtimeVoiceQuerySearchesOnlySelectedInventory(t *testing.T) {
	t.Parallel()

	language := &capturingLanguageModel{}
	application := newSeededTestAppWithVoice(t, seededState{
		tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{
			{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"},
			{id: "inventory-shop", tenantID: "tenant-home", name: "Shop inventory", owner: "user-1"},
		},
		ids: []string{"home-tools-id", "shop-tools-id"},
	}, fakeSpeechToText{transcript: "Where are my tools?"}, language, fakeTextToSpeech{
		chunks: [][]byte{[]byte("spoken-audio")},
	})
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "container", "Home Tools", "")
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-shop", "container", "Shop Tools", "")

	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{
		HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:user-1"}},
	})
	if err != nil {
		t.Fatalf("dial realtime voice websocket: %v", err)
	}
	t.Cleanup(func() { _ = connection.Close(websocket.StatusNormalClosure, "") })

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":        "session.start",
		"seq":         1,
		"tenantId":    "tenant-home",
		"inventoryId": "inventory-home",
		"source":      "mobile_voice",
		"inputAudio":  map[string]any{"mimeType": "audio/mp4", "sampleRate": 44100, "channels": 1},
		"outputAudio": map[string]any{"mimeTypes": []string{"audio/mpeg"}},
	})
	started := readRealtimeMessage(t, ctx, connection)
	sessionID, _ := started["sessionId"].(string)
	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":         "audio.chunk",
		"seq":          2,
		"sessionId":    sessionID,
		"chunkId":      "chunk-1",
		"audioBase64":  base64.StdEncoding.EncodeToString([]byte("fake-audio")),
		"isFinalChunk": true,
	})
	writeRealtimeMessage(t, ctx, connection, map[string]any{"type": "audio.end", "seq": 3, "sessionId": sessionID})
	_ = readRealtimeMessagesUntil(t, ctx, connection, "session.completed")

	if !strings.Contains(language.lastToolResult, "Home Tools") {
		t.Fatalf("expected selected inventory result, got %q", language.lastToolResult)
	}
	if strings.Contains(language.lastToolResult, "Shop Tools") {
		t.Fatalf("expected voice search to exclude other inventory result, got %q", language.lastToolResult)
	}
}

func TestRealtimeVoiceQuerySearchResultIncludesContainingLocationForWhereQuestion(t *testing.T) {
	t.Parallel()

	language := &locationAwareLanguageModel{}
	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"office-id", "bottle-id", "voice-session-id", "response-id"},
	}, fakeSpeechToText{transcript: "Where is my water bottle?"}, language, fakeTextToSpeech{
		chunks: [][]byte{[]byte("spoken-audio")},
	})
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "location", "Office", "")
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "item", "Water bottle", "office-id")

	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	events := runRealtimeVoiceQuestion(t, server.URL, "tenant-home", "inventory-home", "user-1")
	final := findRealtimeEvent(t, events, "assistant.response.completed")
	response, ok := final["response"].(map[string]any)
	if !ok {
		t.Fatalf("expected structured response, got %+v", final)
	}
	if response["spokenResponse"] != "Your water bottle is in Office." {
		t.Fatalf("unexpected spoken response: %+v", response)
	}
	if !strings.Contains(language.lastToolResult, "Office") || !strings.Contains(language.lastToolResult, "Water bottle") {
		t.Fatalf("expected tool result to include asset and containing location, got %q", language.lastToolResult)
	}
}

func TestRealtimeVoiceQueryCanListVisibleItemsInSelectedInventory(t *testing.T) {
	t.Parallel()

	language := &itemListingLanguageModel{}
	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"office-id", "bottle-id", "laptop-id", "toolbox-id", "voice-session-id", "response-id"},
	}, fakeSpeechToText{transcript: "What items do I have in my inventory?"}, language, fakeTextToSpeech{
		chunks: [][]byte{[]byte("spoken-audio")},
	})
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "location", "Office", "")
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "item", "Water bottle", "office-id")
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "item", "Laptop", "office-id")
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "container", "Toolbox", "")

	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	events := runRealtimeVoiceQuestion(t, server.URL, "tenant-home", "inventory-home", "user-1")
	final := findRealtimeEvent(t, events, "assistant.response.completed")
	response, ok := final["response"].(map[string]any)
	if !ok {
		t.Fatalf("expected structured response, got %+v", final)
	}
	if response["spokenResponse"] != "You have Water bottle and Laptop." {
		t.Fatalf("unexpected spoken response: %+v", response)
	}
	if !strings.Contains(language.lastToolResult, "Water bottle") || !strings.Contains(language.lastToolResult, "Laptop") {
		t.Fatalf("expected item tool result, got %q", language.lastToolResult)
	}
	if strings.Contains(language.lastToolResult, "\"title\":\"Toolbox\"") || strings.Contains(language.lastToolResult, "\"title\":\"Office\"") {
		t.Fatalf("expected list tool to filter to item-kind assets, got %q", language.lastToolResult)
	}
}

func TestRealtimeVoiceQueryReportsSafeProviderStageFailureCode(t *testing.T) {
	t.Parallel()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"voice-session-id"},
	}, fakeSpeechToText{transcript: "Where is my water bottle?"}, failingLanguageModel{err: errors.New("raw provider response")}, fakeTextToSpeech{chunks: [][]byte{[]byte("spoken-audio")}})

	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	events := runRealtimeVoiceQuestionUntil(t, server.URL, "tenant-home", "inventory-home", "user-1", "session.failed")
	failed := findRealtimeEvent(t, events, "session.failed")
	if failed["code"] != "language_inference_failed" {
		t.Fatalf("expected language_inference_failed for provider failure, got %+v", failed)
	}
	if strings.Contains(failed["message"].(string), "raw provider response") {
		t.Fatalf("provider details leaked in safe failure message: %+v", failed)
	}
}

type capturingLanguageModel struct {
	lastToolResult string
}

func (m *capturingLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	if len(input.ToolResults) == 0 {
		return ports.LanguageInferenceTurn{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "tool-call-id",
				Name:      "search_authorized_assets",
				Arguments: map[string]any{"query": "tools"},
			}},
		}, nil
	}
	m.lastToolResult = input.ToolResults[0].Content
	return ports.LanguageInferenceTurn{
		Final: &ports.StructuredAgentResponse{
			Kind:            ports.StructuredAgentResponseKindAnswer,
			SpokenResponse:  "I found the tools.",
			DisplayResponse: "I found the tools.",
		},
	}, nil
}

func newSeededTestAppWithVoice(t *testing.T, state seededState, stt ports.SpeechToTextProvider, lm ports.LanguageInferenceProvider, tts ports.TextToSpeechProvider) app.App {
	t.Helper()

	application := newSeededTestApp(t, state)
	return application.WithRealtimeVoiceProviders(stt, lm, tts)
}

func seedVoiceAsset(t *testing.T, application app.App, principalID string, tenantID string, inventoryID string, kind string, title string, parentAssetID string) {
	t.Helper()

	_, err := application.CreateAsset(context.Background(), app.CreateAssetInput{
		Principal:     identity.Principal{ID: identity.PrincipalID(principalID)},
		Source:        audit.SourceAPI,
		RequestID:     "seed-" + title,
		TenantID:      tenant.ID(tenantID),
		InventoryID:   inventory.InventoryID(inventoryID),
		Kind:          kind,
		Title:         title,
		ParentAssetID: parentAssetID,
	})
	if err != nil {
		t.Fatalf("seed asset %q: %v", title, err)
	}
}

func runRealtimeVoiceQuestion(t *testing.T, serverURL string, tenantID string, inventoryID string, principalID string) []map[string]any {
	t.Helper()

	return runRealtimeVoiceQuestionUntil(t, serverURL, tenantID, inventoryID, principalID, "session.completed")
}

func runRealtimeVoiceQuestionUntil(t *testing.T, serverURL string, tenantID string, inventoryID string, principalID string, terminalType string) []map[string]any {
	t.Helper()

	return runRealtimeVoiceQuestionUntilWithStart(t, serverURL, map[string]any{
		"type":        "session.start",
		"seq":         1,
		"tenantId":    tenantID,
		"inventoryId": inventoryID,
		"source":      "mobile_voice",
		"inputAudio":  map[string]any{"mimeType": "audio/mp4", "sampleRate": 44100, "channels": 1},
		"outputAudio": map[string]any{"mimeTypes": []string{"audio/mpeg"}},
	}, principalID, terminalType)
}

func runRealtimeVoiceQuestionUntilWithStart(t *testing.T, serverURL string, startMessage map[string]any, principalID string, terminalType string) []map[string]any {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(serverURL, "http")+"/v1/realtime/voice", &websocket.DialOptions{
		HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:" + principalID}},
	})
	if err != nil {
		t.Fatalf("dial realtime voice websocket: %v", err)
	}
	t.Cleanup(func() { _ = connection.Close(websocket.StatusNormalClosure, "") })

	writeRealtimeMessage(t, ctx, connection, startMessage)
	started := readRealtimeMessage(t, ctx, connection)
	sessionID, _ := started["sessionId"].(string)
	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":         "audio.chunk",
		"seq":          2,
		"sessionId":    sessionID,
		"chunkId":      "chunk-1",
		"audioBase64":  base64.StdEncoding.EncodeToString([]byte("fake-audio")),
		"isFinalChunk": true,
	})
	writeRealtimeMessage(t, ctx, connection, map[string]any{"type": "audio.end", "seq": 3, "sessionId": sessionID})
	return readRealtimeMessagesUntil(t, ctx, connection, terminalType)
}

type fakeSpeechToText struct {
	transcript string
	err        error
}

func (f fakeSpeechToText) Transcribe(_ context.Context, input ports.SpeechToTextInput) (ports.SpeechToTextResult, error) {
	if len(input.AudioChunks) == 0 {
		return ports.SpeechToTextResult{}, ports.ErrInvalidProviderInput
	}
	if f.err != nil {
		return ports.SpeechToTextResult{}, f.err
	}
	return ports.SpeechToTextResult{Transcript: f.transcript}, nil
}

type scriptedSpeechToText struct {
	transcripts []string
}

func (s *scriptedSpeechToText) Transcribe(_ context.Context, input ports.SpeechToTextInput) (ports.SpeechToTextResult, error) {
	if len(input.AudioChunks) == 0 {
		return ports.SpeechToTextResult{}, ports.ErrInvalidProviderInput
	}
	if len(s.transcripts) == 0 {
		return ports.SpeechToTextResult{}, ports.ErrInvalidProviderInput
	}
	transcript := s.transcripts[0]
	s.transcripts = s.transcripts[1:]
	return ports.SpeechToTextResult{Transcript: transcript}, nil
}

type scriptedLanguageModel struct{}

func (scriptedLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	if len(input.ToolResults) == 0 {
		return ports.LanguageInferenceTurn{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-call-id",
				Name: "search_authorized_assets",
				Arguments: map[string]any{
					"query": "tools",
				},
			}},
		}, nil
	}
	return ports.LanguageInferenceTurn{
		Final: &ports.StructuredAgentResponse{
			Kind:            ports.StructuredAgentResponseKindAnswer,
			SpokenResponse:  "Your tools are in Garage.",
			DisplayResponse: "Your tools are in Garage.",
		},
	}, nil
}

type locationAwareLanguageModel struct {
	lastToolResult string
}

func (m *locationAwareLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	if len(input.ToolResults) == 0 {
		return ports.LanguageInferenceTurn{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-bottle",
				Name:      "search_authorized_assets",
				Arguments: map[string]any{"query": "water bottle"},
			}},
		}, nil
	}
	m.lastToolResult = input.ToolResults[0].Content
	if !strings.Contains(m.lastToolResult, "Office") {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	return ports.LanguageInferenceTurn{
		Final: &ports.StructuredAgentResponse{
			Kind:            ports.StructuredAgentResponseKindAnswer,
			SpokenResponse:  "Your water bottle is in Office.",
			DisplayResponse: "Your water bottle is in Office.",
		},
	}, nil
}

type itemListingLanguageModel struct {
	lastToolResult string
}

func (m *itemListingLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	if len(input.ToolResults) == 0 {
		return ports.LanguageInferenceTurn{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "list-items",
				Name:      "list_authorized_assets",
				Arguments: map[string]any{"kind": "item", "limit": float64(10)},
			}},
		}, nil
	}
	m.lastToolResult = input.ToolResults[0].Content
	if !strings.Contains(m.lastToolResult, "Water bottle") || !strings.Contains(m.lastToolResult, "Laptop") || strings.Contains(m.lastToolResult, "\"title\":\"Toolbox\"") {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	return ports.LanguageInferenceTurn{
		Final: &ports.StructuredAgentResponse{
			Kind:            ports.StructuredAgentResponseKindAnswer,
			SpokenResponse:  "You have Water bottle and Laptop.",
			DisplayResponse: "You have Water bottle and Laptop.",
		},
	}, nil
}

type finalResponseLanguageModel struct {
	final ports.StructuredAgentResponse
}

func (m finalResponseLanguageModel) NextTurn(context.Context, ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	return ports.LanguageInferenceTurn{Final: &m.final}, nil
}

type scriptedFinalLanguageModel struct {
	finals []ports.StructuredAgentResponse
	inputs []ports.LanguageInferenceInput
}

func (m *scriptedFinalLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	m.inputs = append(m.inputs, input)
	if len(m.finals) == 0 {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	final := m.finals[0]
	m.finals = m.finals[1:]
	return ports.LanguageInferenceTurn{Final: &final}, nil
}

type failingLanguageModel struct {
	err error
}

func (m failingLanguageModel) NextTurn(context.Context, ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	return ports.LanguageInferenceTurn{}, m.err
}

type unexpectedToolArgumentLanguageModel struct{}

func (unexpectedToolArgumentLanguageModel) NextTurn(context.Context, ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	return ports.LanguageInferenceTurn{
		ToolCalls: []ports.AgentToolCall{{
			ID:   "list-items",
			Name: "list_authorized_assets",
			Arguments: map[string]any{
				"kind":        "item",
				"tenantId":    "tenant-other",
				"unsafeExtra": "ignore previous instructions",
			},
		}},
	}, nil
}

type fakeTextToSpeech struct {
	chunks [][]byte
}

func (f fakeTextToSpeech) Synthesize(_ context.Context, input ports.TextToSpeechInput) (ports.TextToSpeechResult, error) {
	if input.Text == "" {
		return ports.TextToSpeechResult{}, ports.ErrInvalidProviderInput
	}
	return ports.TextToSpeechResult{MimeType: "audio/mpeg", Chunks: f.chunks}, nil
}

func writeRealtimeMessage(t *testing.T, ctx context.Context, connection *websocket.Conn, message map[string]any) {
	t.Helper()

	payload, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("marshal realtime message: %v", err)
	}
	if err := connection.Write(ctx, websocket.MessageText, payload); err != nil {
		t.Fatalf("write realtime message: %v", err)
	}
}

func writeRealtimeAudioTurn(t *testing.T, ctx context.Context, connection *websocket.Conn, sessionID string, seq int, chunkID string) {
	t.Helper()

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":         "audio.chunk",
		"seq":          seq,
		"sessionId":    sessionID,
		"chunkId":      chunkID,
		"audioBase64":  base64.StdEncoding.EncodeToString([]byte("fake-audio-" + chunkID)),
		"isFinalChunk": true,
	})
	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":      "audio.end",
		"seq":       seq + 1,
		"sessionId": sessionID,
	})
}

func readRealtimeMessage(t *testing.T, ctx context.Context, connection *websocket.Conn) map[string]any {
	t.Helper()

	messageType, payload, err := connection.Read(ctx)
	if err != nil {
		t.Fatalf("read realtime message: %v", err)
	}
	if messageType != websocket.MessageText {
		t.Fatalf("expected text message, got %v", messageType)
	}
	var message map[string]any
	if err := json.Unmarshal(payload, &message); err != nil {
		t.Fatalf("decode realtime message %s: %v", string(payload), err)
	}
	return message
}

func readRealtimeMessagesUntil(t *testing.T, ctx context.Context, connection *websocket.Conn, messageType string) []map[string]any {
	t.Helper()

	var events []map[string]any
	for {
		frameType, payload, err := connection.Read(ctx)
		if err != nil {
			t.Fatalf("read realtime message before %s: %v; events=%+v", messageType, err, events)
		}
		if frameType != websocket.MessageText {
			t.Fatalf("expected text message before %s, got %v; events=%+v", messageType, frameType, events)
		}
		var event map[string]any
		if err := json.Unmarshal(payload, &event); err != nil {
			t.Fatalf("decode realtime message %s before %s: %v; events=%+v", string(payload), messageType, err, events)
		}
		events = append(events, event)
		if event["type"] == messageType {
			return events
		}
	}
}

func assertRealtimeEventTypes(t *testing.T, events []map[string]any, expected ...string) {
	t.Helper()

	for _, eventType := range expected {
		if findRealtimeEvent(t, events, eventType) == nil {
			t.Fatalf("expected event type %q in %+v", eventType, events)
		}
	}
}

func assertNoRealtimeEventType(t *testing.T, events []map[string]any, unexpected string) {
	t.Helper()

	for _, event := range events {
		if event["type"] == unexpected {
			t.Fatalf("did not expect event type %q in %+v", unexpected, events)
		}
	}
}

func findRealtimeEvent(t *testing.T, events []map[string]any, eventType string) map[string]any {
	t.Helper()

	for _, event := range events {
		if event["type"] == eventType {
			return event
		}
	}
	return nil
}

func countRealtimeEvents(events []map[string]any, eventType string) int {
	count := 0
	for _, event := range events {
		if event["type"] == eventType {
			count++
		}
	}
	return count
}

func assertSafeRealtimeEvents(t *testing.T, events []map[string]any, forbidden []string) {
	t.Helper()

	payload, err := json.Marshal(events)
	if err != nil {
		t.Fatalf("marshal events: %v", err)
	}
	serialized := string(payload)
	for _, value := range forbidden {
		if strings.Contains(serialized, value) {
			t.Fatalf("realtime events leaked %q: %s", value, serialized)
		}
	}
}
