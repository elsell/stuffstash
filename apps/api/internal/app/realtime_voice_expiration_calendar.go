package app

import (
	"context"
	"encoding/json"
	notificationapp "github.com/stuffstash/stuff-stash/internal/app/notifications"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

const RealtimeVoiceToolGetExpirationCalendar = "get_expiration_calendar"

func realtimeConversationExpirationCalendarTool() ports.ConversationToolDefinition {
	return ports.ConversationToolDefinition{Name: RealtimeVoiceToolGetExpirationCalendar, Description: "Get verified current date, currentMonth and nextMonth in this person's inventory timezone. Refresh for each new relative-date request; do not reuse an earlier turn's calendar. Next month means the returned YYYY-MM with month precision, not an invented day. Clarify ambiguous dates or missing years.", Parameters: json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`)}
}
func (a App) executeRealtimeVoiceExpirationCalendar(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall) (ports.AgentToolResult, error) {
	if len(call.Arguments) != 0 || a.clock == nil {
		return ports.AgentToolResult{}, ports.ErrInvalidProviderInput
	}
	preferences, err := a.notificationService.GetPreferences(ctx, notificationapp.ScopeInput{Principal: session.Principal, TenantID: session.TenantID, InventoryID: session.InventoryID, Source: audit.SourceConversation})
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	zone, err := time.LoadLocation(preferences.Settings.Timezone)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	today := a.clock.Now().In(zone)
	month := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	if err := a.ensureRealtimeVoiceAccess(ctx, session.Principal, session.TenantID, session.InventoryID); err != nil {
		return ports.AgentToolResult{}, err
	}
	encoded, err := json.Marshal(struct {
		Date         string `json:"date"`
		Timezone     string `json:"timezone"`
		CurrentMonth string `json:"currentMonth"`
		NextMonth    string `json:"nextMonth"`
	}{today.Format("2006-01-02"), preferences.Settings.Timezone, month.Format("2006-01"), month.AddDate(0, 1, 0).Format("2006-01")})
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	return ports.AgentToolResult{CallID: call.ID, Name: call.Name, Call: call, Content: string(encoded)}, nil
}
