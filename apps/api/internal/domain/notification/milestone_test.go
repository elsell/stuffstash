package notification

import (
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"testing"
	"time"
)

func TestExpirationMilestoneEvaluationAndVisibility(t *testing.T) {
	date, err := expirationdate.ParseDate("2026-10-01", expirationdate.Day)
	if err != nil {
		t.Fatal(err)
	}
	candidate := ExpirationCandidate{AssetID: "bottle", TypeID: "medicine", Date: date, Eligible: true}
	settings := DefaultSettings("UTC")
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	upcoming, ok := candidate.Due(settings, now, time.UTC)
	if !ok || upcoming.Kind != MilestoneUpcoming || upcoming.AssetID != "bottle" || upcoming.Date != date {
		t.Fatalf("upcoming %+v %v", upcoming, ok)
	}
	settings.Defaults.Enabled = false
	if _, ok := candidate.Due(settings, now, time.UTC); ok {
		t.Fatal("disabled defaults generated notification")
	}
	if !candidate.Visible(upcoming, settings, now, time.UTC) {
		t.Fatal("disabling preferences erased history")
	}
	settings.Overrides["medicine"] = DefaultExpirationPreferences()
	if _, ok := candidate.Due(settings, now, time.UTC); !ok {
		t.Fatal("enabled override did not generate")
	}
	after := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	expired, ok := candidate.Due(settings, after, time.UTC)
	if !ok || expired.Kind != MilestoneExpired {
		t.Fatal("already expired item must produce only expired milestone")
	}
	if !candidate.Visible(upcoming, settings, after, time.UTC) {
		t.Fatal("upcoming history disappeared after expiration")
	}
	if candidate.Visible(expired, settings, now, time.UTC) {
		t.Fatal("expired milestone visible before expiry")
	}
	candidate.Eligible = false
	if candidate.Visible(upcoming, settings, now, time.UTC) {
		t.Fatal("ineligible asset visible")
	}
	if _, ok := candidate.Due(settings, now, time.UTC); ok {
		t.Fatal("ineligible asset generated")
	}
	candidate.Eligible = true
	candidate.Date, _ = expirationdate.ParseDate("2027-10-01", expirationdate.Day)
	if candidate.Visible(upcoming, settings, now, time.UTC) {
		t.Fatal("stale date visible")
	}
}

func TestExpirationMilestoneCalendarAndThreshold(t *testing.T) {
	date, _ := expirationdate.ParseDate("2026-09", expirationdate.Month)
	candidate := ExpirationCandidate{AssetID: "bottle", TypeID: "medicine", Date: date, Eligible: true}
	settings := DefaultSettings("America/New_York")
	zone, err := time.LoadLocation(settings.Timezone)
	if err != nil {
		t.Fatal(err)
	}
	lastDay := time.Date(2026, 9, 30, 23, 59, 0, 0, zone)
	value, ok := candidate.Due(settings, lastDay, zone)
	if !ok || value.Kind != MilestoneUpcoming {
		t.Fatal("month expires before month ends")
	}
	value, ok = candidate.Due(settings, lastDay.Add(time.Minute), zone)
	if !ok || value.Kind != MilestoneExpired {
		t.Fatal("month did not expire in personal timezone")
	}
	settings.Defaults.Expired = false
	if _, ok := candidate.Due(settings, lastDay.Add(time.Minute), zone); ok {
		t.Fatal("disabled expired milestone generated")
	}
	settings.Defaults.AdvanceDays = 0
	if candidate.Visible(Milestone{AssetID: "bottle", Date: date, Kind: MilestoneUpcoming}, settings, time.Date(2026, 9, 29, 12, 0, 0, 0, zone), zone) {
		t.Fatal("outside shortened warning window")
	}
	candidate.Date = expirationdate.Date{}
	if _, ok := candidate.Due(settings, lastDay, zone); ok {
		t.Fatal("unset date generated")
	}
}
