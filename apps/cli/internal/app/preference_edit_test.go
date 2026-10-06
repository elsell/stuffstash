package app

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"testing"
)

type keepPreferenceText struct{}

func (keepPreferenceText) ReadText(context.Context, string, int) (string, error) { return "", nil }
func TestGuidedPreferenceEditPreservesCurrentSettings(t *testing.T) {
	current := ports.NotificationPreferences{Revision: 9007199254740993, Timezone: "America/New_York", PushEnabled: false, Defaults: ports.ExpirationPolicy{Enabled: false, Upcoming: true, Expired: false, AdvanceDays: 0}, Overrides: []ports.NotificationTypeOverride{{CustomAssetTypeID: "food", Settings: ports.ExpirationPolicy{Enabled: true, Expired: true, AdvanceDays: 42}}}}
	r := Runner{Picker: firstScopeChoice{}, TextInput: keepPreferenceText{}}
	for _, action := range []string{"update", "override"} {
		body, err := r.editPreferenceBody(context.Background(), Options{Command: []string{"notification-preferences", action, "food"}}, current)
		if err != nil {
			t.Fatal(err)
		}
		var result struct {
			Revision    int64
			Defaults    ports.ExpirationPolicy
			Settings    ports.ExpirationPolicy
			Timezone    string
			PushEnabled bool
		}
		if err := json.Unmarshal(body, &result); err != nil {
			t.Fatal(err)
		}
		if result.Revision != current.Revision {
			t.Fatal("revision changed")
		}
		if action == "update" && (result.Defaults != current.Defaults || result.Timezone != current.Timezone || result.PushEnabled != current.PushEnabled) {
			t.Fatalf("defaults reset: %s", body)
		}
		if action == "override" && result.Settings != current.Overrides[0].Settings {
			t.Fatalf("override reset: %s", body)
		}
	}
}
