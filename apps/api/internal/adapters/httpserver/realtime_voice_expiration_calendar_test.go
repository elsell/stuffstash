package httpserver

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"net/http/httptest"
	"testing"
)

type expirationCalendarModel struct {
	arguments map[string]any
	result    string
}

func (m *expirationCalendarModel) Converse(_ context.Context, in ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	last := in.Messages[len(in.Messages)-1]
	if last.Role == ports.ConversationRoleUser {
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "calendar", Name: app.RealtimeVoiceToolGetExpirationCalendar, Arguments: m.arguments}}}, nil
	}
	for _, result := range last.ToolResults {
		if result.Name == app.RealtimeVoiceToolGetExpirationCalendar {
			m.result = result.Content
		}
	}
	return ports.ConversationModelTurn{Text: "Calendar checked."}, nil
}
func TestVoiceCalendarRejectsModelScopeAtWebSocketBoundary(t *testing.T) {
	for _, key := range []string{"", "tenantId", "inventoryId", "principalId", "timezone"} {
		t.Run(key, func(t *testing.T) {
			model := &expirationCalendarModel{arguments: map[string]any{}}
			if key != "" {
				model.arguments[key] = "other"
			}
			application := newSeededTestAppWithVoice(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}}}, fakeSpeechToText{transcript: "What is next month?"}, model, fakeTextToSpeech{chunks: [][]byte{[]byte("audio")}})
			server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
			t.Cleanup(server.Close)
			runRealtimeVoiceQuestion(t, server.URL, "tenant-home", "inventory-home", "user-1")
			var output struct{ Date, Timezone, Error string }
			if json.Unmarshal([]byte(model.result), &output) != nil {
				t.Fatal("missing calendar result")
			}
			if key == "" {
				if output.Date == "" || output.Timezone != "UTC" || output.Error != "" {
					t.Fatalf("valid calendar failed: %s", model.result)
				}
			} else if output.Error == "" || output.Date != "" {
				t.Fatal("model scope accepted")
			}
		})
	}
}
