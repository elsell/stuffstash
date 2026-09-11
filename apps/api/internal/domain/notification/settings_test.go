package notification_test

import (
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"testing"
)

func TestSettingsValidateTimezoneAndCloneOverrides(t *testing.T) {
	value := notification.DefaultSettings("America/New_York")
	if err := value.Validate(); err != nil {
		t.Fatal(err)
	}
	value.Overrides[notification.AssetTypeID("medicine")] = notification.ExpirationPreferences{Enabled: true, Upcoming: true, AdvanceDays: 60}
	copy := value.Clone()
	delete(copy.Overrides, notification.AssetTypeID("medicine"))
	if len(value.Overrides) != 1 {
		t.Fatal("copy changed saved overrides")
	}
	for _, zone := range []string{"", "Local", "Mars/Base", " UTC "} {
		copy.Timezone = zone
		if copy.Validate() == nil {
			t.Fatalf("accepted timezone %q", zone)
		}
	}
	copy.Timezone = "UTC"
	copy.Overrides[""] = notification.DefaultExpirationPreferences()
	if copy.Validate() == nil {
		t.Fatal("accepted empty type override")
	}
}
