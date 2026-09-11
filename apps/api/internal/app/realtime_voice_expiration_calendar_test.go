package app

import (
	"context"
	"encoding/json"
	notificationapp "github.com/stuffstash/stuff-stash/internal/app/notifications"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestVoiceExpirationCalendarUsesPersonalTimezoneAndClock(t *testing.T) {
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, successfulRealtimeVoiceResolver())
	application.clock = fakeClock{now: time.Date(2028, 1, 1, 2, 0, 0, 0, time.UTC)}
	application.notificationService = notificationapp.New(notificationapp.Dependencies{Authorizer: application.authorizer, Inventories: store, Preferences: store, Audit: store, Clock: application.clock, IDs: application.ids})
	scope := notificationapp.ScopeInput{Principal: checkoutToolSession().Principal, TenantID: "tenant-home", InventoryID: "inventory-home"}
	if _, err := application.notificationService.InitializePreferences(context.Background(), scope, "America/New_York"); err != nil {
		t.Fatal(err)
	}
	call := ports.AgentToolCall{ID: "calendar", Name: "get_expiration_calendar", Arguments: map[string]any{}}
	result, err := application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), call, nil)
	if err != nil {
		t.Fatal(err)
	}
	var output struct{ Date, Timezone, CurrentMonth, NextMonth string }
	if err := json.Unmarshal([]byte(result.Content), &output); err != nil {
		t.Fatal(err)
	}
	if output.Date != "2027-12-31" || output.Timezone != "America/New_York" || output.CurrentMonth != "2027-12" || output.NextMonth != "2028-01" {
		t.Fatalf("wrong calendar: %s", result.Content)
	}
	application.clock = fakeClock{now: time.Date(2028, 1, 1, 6, 0, 0, 0, time.UTC)}
	result, err = application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), call, nil)
	if err != nil {
		t.Fatal(err)
	}
	if json.Unmarshal([]byte(result.Content), &output) != nil || output.Date != "2028-01-01" || output.NextMonth != "2028-02" {
		t.Fatal("calendar reused previous day")
	}
	call.Arguments["timezone"] = "Pacific/Auckland"
	if _, err := application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), call, nil); err == nil {
		t.Fatal("model supplied timezone accepted")
	}
	delete(call.Arguments, "timezone")
	prefs, err := application.notificationService.GetPreferences(context.Background(), scope)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := application.notificationService.UpdatePreferences(context.Background(), scope, prefs.Revision, prefs.Settings.Defaults, "America/Asuncion", false); err != nil {
		t.Fatal(err)
	}
	application.clock = fakeClock{now: time.Date(2017, 10, 15, 12, 0, 0, 0, time.UTC)}
	result, err = application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), call, nil)
	if err != nil {
		t.Fatal(err)
	}
	if json.Unmarshal([]byte(result.Content), &output) != nil || output.CurrentMonth != "2017-10" || output.NextMonth != "2017-11" {
		t.Fatalf("DST month boundary shifted: %s", result.Content)
	}

}
