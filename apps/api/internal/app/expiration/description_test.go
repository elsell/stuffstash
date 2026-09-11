package expiration

import (
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"testing"
	"time"
)

func TestDescriptionUsesPersonalCalendarRegardlessOfDelivery(t *testing.T) {
	date, _ := expirationdate.ParseDate("2026-09", expirationdate.Month)
	settings := notification.DefaultSettings("America/New_York")
	settings.Defaults = notification.ExpirationPreferences{AdvanceDays: 0}
	settings.Overrides["medicine"] = notification.ExpirationPreferences{AdvanceDays: 30}
	now := time.Date(2026, 10, 1, 3, 59, 0, 0, time.UTC)
	value, err := Describe(date, "medicine", true, settings, now)
	if err != nil || value.State != expirationdate.Upcoming || value.Date.Value() != "2026-09" || value.Date.Precision() != expirationdate.Month || value.AdvanceDays != 30 || !value.TrackingEnabled {
		t.Fatalf("wrong description: %+v %v", value, err)
	}
	value, err = Describe(date, "medicine", false, settings, now.Add(time.Minute))
	if err != nil || value.State != expirationdate.Expired || value.TrackingEnabled {
		t.Fatalf("disabled tracking erased facts: %+v %v", value, err)
	}
	unknown, err := Describe(expirationdate.Date{}, "medicine", true, settings, now)
	if err != nil || unknown.State != expirationdate.Unset {
		t.Fatal("missing date claimed to be current")
	}
	settings.Timezone = "invalid-zone"
	if _, err := Describe(date, "medicine", true, settings, now); err == nil {
		t.Fatal("invalid timezone silently accepted")
	}
}
