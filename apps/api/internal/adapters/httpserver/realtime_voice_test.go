package httpserver

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceQueryWebSocketStreamsTranscriptToolResultAndSpeech(t *testing.T) {
	t.Parallel()

	store := memory.NewStore()
	authorizer := memory.NewAuthorizer()
	application := newSeededTestAppWithStoreAndAuthorizer(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"garage-id", "tools-id", "voice-session-id", "tool-call-id", "response-id"},
	}, store, authorizer).WithRealtimeVoiceProviders(fakeSpeechToText{transcript: "Where are my tools?"}, scriptedLanguageModel{}, fakeTextToSpeech{
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
	if response["spokenResponse"] != "Tools. Its recorded path is Garage / Tools." {
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
	record, found, err := store.RealtimeSessionByID(ctx, tenant.ID("tenant-home"), inventory.InventoryID("inventory-home"), sessionID)
	if err != nil || !found {
		t.Fatalf("load realtime session record: found=%v err=%v", found, err)
	}
	if record.State != ports.RealtimeSessionStateCompleted || record.EndedAt.IsZero() || record.SafeFailureCode != "" {
		t.Fatalf("expected direct answer websocket session to be marked completed, got %+v", record)
	}
}

func TestRealtimeVoiceWebSocketAcceptsFollowUpAudioAfterClarification(t *testing.T) {
	t.Parallel()

	language := &scriptedFinalLanguageModel{}
	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids: []string{
			"first-item-id", "first-item-undo", "first-item-audit",
			"second-item-id", "second-item-undo", "second-item-audit",
			"office-id", "office-undo", "office-audit",
			"voice-session-id", "clarification-response-id", "answer-response-id",
		},
	}, &scriptedSpeechToText{transcripts: []string{"Where should I put it?", "Put it in the office."}}, language, fakeTextToSpeech{chunks: [][]byte{[]byte("spoken-audio")}})
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "item", "First item", "")
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "item", "Second item", "")
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "location", "Office", "")

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
	if secondPayload["kind"] != "answer" || !strings.Contains(secondPayload["spokenResponse"].(string), "Office") {
		t.Fatalf("expected follow-up answer on same session, got %+v", secondPayload)
	}
	var followUpInput *ports.LanguageInferenceInput
	for index := range language.inputs {
		if len(language.inputs[index].ConversationTurns) == 2 {
			followUpInput = &language.inputs[index]
			break
		}
	}
	if followUpInput == nil {
		t.Fatalf("expected follow-up language turn to include prior user and assistant context, got %+v", language.inputs)
	}
	if followUpInput.ConversationTurns[0].Role != ports.AgentConversationRoleUser || followUpInput.ConversationTurns[0].Text != "Where should I put it?" {
		t.Fatalf("unexpected prior user context: %+v", followUpInput.ConversationTurns)
	}
	if followUpInput.ConversationTurns[1].Role != ports.AgentConversationRoleAssistant || followUpInput.ConversationTurns[1].Kind != string(ports.StructuredAgentResponseKindClarification) {
		t.Fatalf("unexpected prior assistant context: %+v", followUpInput.ConversationTurns)
	}
}

func TestRealtimeVoiceWebSocketFailsSafelyAtClarificationTurnLimit(t *testing.T) {
	t.Parallel()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids: []string{
			"first-item-id", "first-item-undo", "first-item-audit",
			"second-item-id", "second-item-undo", "second-item-audit",
			"voice-session-id", "clarification-1", "clarification-2", "clarification-3",
		},
	}, &scriptedSpeechToText{transcripts: []string{"Move it.", "That one.", "Over there."}}, &scriptedFinalLanguageModel{alwaysAmbiguous: true}, fakeTextToSpeech{chunks: [][]byte{[]byte("spoken-audio")}})
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "item", "First item", "")
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "item", "Second item", "")

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

	writeRealtimeMessage(t, ctx, connection, realtimeVoiceStartMessage("tenant-home", "inventory-home"))
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

	writeRealtimeMessage(t, ctx, connection, realtimeVoiceStartMessage("tenant-other", "inventory-other"))
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

	writeRealtimeMessage(t, ctx, connection, realtimeVoiceStartMessage("tenant-home", "inventory-shop"))
	failed := readRealtimeMessage(t, ctx, connection)
	if failed["type"] != "session.failed" {
		t.Fatalf("expected session.failed, got %+v", failed)
	}
	if failed["code"] != "forbidden" {
		t.Fatalf("expected forbidden, got %+v", failed)
	}
}

func TestRealtimeVoiceQueryRejectsMissingRequestedCapabilities(t *testing.T) {
	t.Parallel()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"voice-session-id"},
	}, fakeSpeechToText{}, scriptedLanguageModel{}, fakeTextToSpeech{})
	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	tests := []struct {
		name         string
		capabilities []string
		omit         bool
	}{
		{name: "missing", omit: true},
		{name: "partial", capabilities: []string{"speech_to_text", "text_to_speech"}},
		{name: "unknown", capabilities: []string{"speech_to_text", "language_inference", "text_to_speech", "raw_provider"}},
		{name: "unknown replacement", capabilities: []string{"speech_to_text", "language_inference", "raw_provider"}},
		{name: "duplicate replacement", capabilities: []string{"speech_to_text", "language_inference", "language_inference"}},
		{name: "whitespace padded", capabilities: []string{" speech_to_text ", "language_inference", "text_to_speech"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			t.Cleanup(cancel)
			connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{
				HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:user-1"}},
			})
			if err != nil {
				t.Fatalf("dial realtime voice websocket: %v", err)
			}
			t.Cleanup(func() { _ = connection.Close(websocket.StatusNormalClosure, "") })

			start := realtimeVoiceStartMessage("tenant-home", "inventory-home")
			if tt.omit {
				delete(start, "requestedCapabilities")
			} else {
				start["requestedCapabilities"] = tt.capabilities
			}
			writeRealtimeMessage(t, ctx, connection, start)
			failed := readRealtimeMessage(t, ctx, connection)
			if failed["type"] != "session.failed" {
				t.Fatalf("expected session.failed, got %+v", failed)
			}
			if failed["code"] != "invalid_request" {
				t.Fatalf("expected invalid request, got %+v", failed)
			}
		})
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

	writeRealtimeMessage(t, ctx, connection, realtimeVoiceStartMessage("tenant-home", "inventory-home"))
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

	writeRealtimeMessage(t, ctx, connection, realtimeVoiceStartMessage("tenant-home", "inventory-home"))
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
	if response["spokenResponse"] != "Water bottle. Its recorded path is Office / Water bottle." {
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
	if response["spokenResponse"] != "I found 2 visible matches: Laptop, Water bottle." {
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
