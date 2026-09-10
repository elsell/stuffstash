package notification_test

import (
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"testing"
)

func TestTypeOverrideCanEnableDisabledInventoryDefault(t *testing.T) {
	defaults := notification.DefaultExpirationPreferences()
	defaults.Enabled = false
	inherited := notification.ResolveExpirationPreferences(defaults, nil)
	if inherited.Enabled {
		t.Fatal("missing override must inherit disabled default")
	}
	override := notification.ExpirationPreferences{Enabled: true, Upcoming: true, Expired: false, AdvanceDays: 60}
	resolved := notification.ResolveExpirationPreferences(defaults, &override)
	if resolved != override {
		t.Fatalf("override was masked by inventory defaults: %+v", resolved)
	}
	if notification.ResolveExpirationPreferences(defaults, nil) != defaults {
		t.Fatal("reset must restore inheritance")
	}
}

func TestExpirationPreferencesValidateWithoutChangingIndependentSwitches(t *testing.T) {
	value := notification.DefaultExpirationPreferences()
	if !value.Enabled || !value.Upcoming || !value.Expired || value.AdvanceDays != 30 {
		t.Fatal("incorrect defaults")
	}
	for _, days := range []int{0, 1, 30, 3650} {
		value.AdvanceDays = days
		value.Upcoming = false
		if err := value.Validate(); err != nil {
			t.Fatal(err)
		}
		if value.Upcoming {
			t.Fatal("validation changed personal setting")
		}
	}
	for _, days := range []int{-1, 3651} {
		value.AdvanceDays = days
		if value.Validate() == nil {
			t.Fatalf("accepted %d days", days)
		}
	}
}
