package httpserver

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceQueryStreamsActionPlanProposalForReview(t *testing.T) {
	t.Parallel()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"voice-session-id", "plan-id", "command-id", "response-id"},
	}, fakeSpeechToText{transcript: "Add a water bottle."}, actionPlanProposalLanguageModel{}, fakeTextToSpeech{
		chunks: [][]byte{[]byte("spoken-audio")},
	})
	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	events := runRealtimeVoiceQuestion(t, server.URL, "tenant-home", "inventory-home", "user-1")
	proposed := findRealtimeEvent(t, events, "action.plan.proposed")
	actionPlan, ok := proposed["actionPlan"].(map[string]any)
	if !ok {
		t.Fatalf("expected action plan payload, got %+v", proposed)
	}
	if actionPlan["planId"] != "plan-id" || actionPlan["confirmationSummary"] != "Create item water bottle?" {
		t.Fatalf("unexpected action plan payload: %+v", actionPlan)
	}
	commands, ok := actionPlan["commands"].([]any)
	if !ok || len(commands) != 1 {
		t.Fatalf("expected one safe command, got %+v", actionPlan["commands"])
	}
	command, ok := commands[0].(map[string]any)
	if !ok || command["kind"] != "create_asset" || command["summary"] != "Create item water bottle" {
		t.Fatalf("unexpected command payload: %+v", commands[0])
	}
	assertSafeRealtimeEvents(t, events, []string{"fake-audio", "apiKey", "Bearer", "provider_session_id"})
}

func TestRealtimeVoiceActionPlanApprovalUsesOpenReviewSession(t *testing.T) {
	t.Parallel()

	ctx, connection, sessionID := openRealtimeVoiceReviewSession(t)

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":      "action.plan.approve",
		"seq":       4,
		"sessionId": sessionID,
		"planId":    "plan-id",
	})
	approved := readRealtimeMessage(t, ctx, connection)
	if approved["type"] != "action.plan.approved" || approved["planId"] != "plan-id" || approved["status"] != "approved" {
		t.Fatalf("expected safe approval event, got %+v", approved)
	}
	assertSafeRealtimeEvents(t, []map[string]any{approved}, []string{"fake-audio", "apiKey", "Bearer", "provider_session_id"})
}

func TestRealtimeVoiceActionPlanCancellationUsesOpenReviewSession(t *testing.T) {
	t.Parallel()

	ctx, connection, sessionID := openRealtimeVoiceReviewSession(t)
	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":      "action.plan.cancel",
		"seq":       4,
		"sessionId": sessionID,
		"planId":    "plan-id",
	})
	cancelled := readRealtimeMessage(t, ctx, connection)
	if cancelled["type"] != "action.plan.cancelled" || cancelled["planId"] != "plan-id" || cancelled["status"] != "cancelled" {
		t.Fatalf("expected safe cancellation event, got %+v", cancelled)
	}
}

func TestRealtimeVoiceActionPlanDecisionRejectsUnsafeMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message map[string]any
	}{
		{
			name: "stale sequence",
			message: map[string]any{
				"type":      "action.plan.approve",
				"seq":       3,
				"sessionId": "voice-session-id",
				"planId":    "plan-id",
			},
		},
		{
			name: "forged session",
			message: map[string]any{
				"type":      "action.plan.approve",
				"seq":       4,
				"sessionId": "voice-session-other",
				"planId":    "plan-id",
			},
		},
		{
			name: "wrong plan",
			message: map[string]any{
				"type":      "action.plan.approve",
				"seq":       4,
				"sessionId": "voice-session-id",
				"planId":    "plan-other",
			},
		},
		{
			name: "forbidden fields",
			message: map[string]any{
				"type":        "action.plan.approve",
				"seq":         4,
				"sessionId":   "voice-session-id",
				"planId":      "plan-id",
				"tenantId":    "tenant-other",
				"inventoryId": "inventory-other",
				"arguments":   map[string]any{"apiKey": "secret"},
			},
		},
		{
			name: "malformed type",
			message: map[string]any{
				"type":      "action.plan.execute",
				"seq":       4,
				"sessionId": "voice-session-id",
				"planId":    "plan-id",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, connection, _ := openRealtimeVoiceReviewSession(t)
			writeRealtimeMessage(t, ctx, connection, tt.message)
			failed := readRealtimeMessage(t, ctx, connection)
			if failed["type"] != "session.failed" {
				t.Fatalf("expected safe failure, got %+v", failed)
			}
			assertSafeRealtimeEvents(t, []map[string]any{failed}, []string{"apiKey", "secret", "tenant-other", "inventory-other", "provider_session_id"})
		})
	}
}

func openRealtimeVoiceReviewSession(t *testing.T) (context.Context, *websocket.Conn, string) {
	t.Helper()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"voice-session-id", "plan-id", "command-id", "response-id"},
	}, fakeSpeechToText{transcript: "Add a water bottle."}, actionPlanProposalLanguageModel{}, fakeTextToSpeech{
		chunks: [][]byte{[]byte("spoken-audio")},
	})
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
	if sessionID != "voice-session-id" {
		t.Fatalf("expected deterministic session id, got %+v", started)
	}
	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":         "audio.chunk",
		"seq":          2,
		"sessionId":    sessionID,
		"chunkId":      "chunk-1",
		"audioBase64":  base64.StdEncoding.EncodeToString([]byte("fake-audio")),
		"isFinalChunk": true,
	})
	writeRealtimeMessage(t, ctx, connection, map[string]any{"type": "audio.end", "seq": 3, "sessionId": sessionID})

	events := readRealtimeMessagesUntil(t, ctx, connection, "session.completed")
	findRealtimeEvent(t, events, "action.plan.proposed")
	return ctx, connection, sessionID
}

type actionPlanProposalLanguageModel struct{}

func (m actionPlanProposalLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	if len(input.ToolResults) == 0 {
		return ports.LanguageInferenceTurn{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "plan-tool-call",
				Name: "propose_action_plan",
				Arguments: map[string]any{
					"commandKind":                "create_asset",
					"intentSummary":              "Create a water bottle item.",
					"modelInterpretationSummary": "The user wants to add a water bottle to this inventory.",
					"confirmationSummary":        "Create item water bottle?",
					"commandSummary":             "Create item water bottle",
					"argumentsJson":              `{"kind":"item","name":"water bottle"}`,
					"riskSummary":                "Adds a new item to this inventory.",
				},
			}},
		}, nil
	}
	return ports.LanguageInferenceTurn{
		Final: &ports.StructuredAgentResponse{
			Kind:            ports.StructuredAgentResponseKindClarification,
			SpokenResponse:  "I prepared that change for review.",
			DisplayResponse: "I prepared that change for review.",
		},
	}, nil
}
