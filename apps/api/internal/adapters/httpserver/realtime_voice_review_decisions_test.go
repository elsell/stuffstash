package httpserver

import "testing"

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
		{
			name: "cancel with approval-only photo metadata item",
			message: map[string]any{
				"type":      "action.plan.cancel",
				"seq":       4,
				"sessionId": "voice-session-id",
				"planId":    "plan-id",
				"photoAttachments": []map[string]any{{
					"commandId":   "cmd-water-bottle",
					"photoIndex":  0,
					"fileName":    "water-bottle.jpg",
					"contentType": "image/jpeg",
					"sizeBytes":   12,
				}},
			},
		},
		{
			name: "cancel with empty approval-only photo metadata",
			message: map[string]any{
				"type":             "action.plan.cancel",
				"seq":              4,
				"sessionId":        "voice-session-id",
				"planId":           "plan-id",
				"photoAttachments": []map[string]any{},
			},
		},
		{
			name: "cancel with null approval-only photo metadata",
			message: map[string]any{
				"type":             "action.plan.cancel",
				"seq":              4,
				"sessionId":        "voice-session-id",
				"planId":           "plan-id",
				"photoAttachments": nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, connection, _, _ := openRealtimeVoiceReviewSession(t)
			writeRealtimeMessage(t, ctx, connection, tt.message)
			failed := readRealtimeMessage(t, ctx, connection)
			if failed["type"] != "session.failed" {
				t.Fatalf("expected safe failure, got %+v", failed)
			}
			assertSafeRealtimeEvents(t, []map[string]any{failed}, []string{"apiKey", "secret", "tenant-other", "inventory-other", "provider_session_id"})
		})
	}
}

func TestRealtimeVoiceActionPlanDecisionAcceptsClientAckBeforeDecision(t *testing.T) {
	t.Parallel()

	ctx, connection, sessionID, planID := openRealtimeVoiceReviewSession(t)
	writeRealtimeMessage(t, ctx, connection, map[string]any{"type": "client.ack", "seq": 4, "sessionId": sessionID, "ackSeq": 3})
	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":      "action.plan.cancel",
		"seq":       5,
		"sessionId": sessionID,
		"planId":    planID,
	})

	cancelled := readRealtimeMessage(t, ctx, connection)
	if cancelled["type"] != "action.plan.cancelled" {
		t.Fatalf("expected action.plan.cancelled after ack and decision, got %+v", cancelled)
	}
}
