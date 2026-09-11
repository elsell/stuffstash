package expiration

import (
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"testing"
	"time"
)

func TestQueryUsesTypeOrTagAndIgnoresDeliverySwitches(t *testing.T) {
	date, _ := expirationdate.ParseDate("2028-02", expirationdate.Month)
	settings := notification.DefaultSettings("America/New_York")
	settings.Defaults = notification.ExpirationPreferences{AdvanceDays: 0}
	settings.Overrides["medicine"] = notification.ExpirationPreferences{AdvanceDays: 30}
	now := time.Date(2028, 2, 10, 12, 0, 0, 0, time.UTC)
	query := Query{Status: QueryUpcoming, TypeID: "medicine", TagKey: "medicine"}
	if !queryMatches(t, query, date, "medicine", nil, settings, now) {
		t.Fatal("type match suppressed by disabled notifications")
	}
	if queryMatches(t, query, date, "food", []string{"medicine"}, settings, now) {
		t.Fatal("ignored other type's zero-day window")
	}
	settings.Defaults.AdvanceDays = 30
	if !queryMatches(t, query, date, "food", []string{"medicine"}, settings, now) {
		t.Fatal("tag alternative not matched")
	}
	if queryMatches(t, query, date, "food", nil, settings, now) || queryMatches(t, query, expirationdate.Date{}, "medicine", nil, settings, now) {
		t.Fatal("category or unknown date leaked into results")
	}
	if queryMatches(t, query, date, "medicine", nil, settings, time.Date(2028, 3, 1, 5, 0, 0, 0, time.UTC)) {
		t.Fatal("expired counted as upcoming")
	}
	query.Status = QueryExpired
	if !queryMatches(t, query, date, "medicine", nil, settings, time.Date(2028, 3, 1, 5, 0, 0, 0, time.UTC)) {
		t.Fatal("expiration timezone boundary missed")
	}
}
func TestQueryDateRangePreservesMonthCalendarMeaning(t *testing.T) {
	date, _ := expirationdate.ParseDate("2028-02", expirationdate.Month)
	query := Query{Status: QueryAll, FromDate: "2028-02-29", ThroughDate: "2028-02-29"}
	settings := notification.DefaultSettings("UTC")
	if query.Validate() != nil || !queryMatches(t, query, date, "", nil, settings, time.Time{}) {
		t.Fatal("month last day not matched")
	}
	query.ThroughDate = "2028-02-28"
	if query.Validate() == nil {
		t.Fatal("reversed range accepted")
	}
	query.FromDate = "2028-02-01"
	if queryMatches(t, query, date, "", nil, settings, time.Time{}) {
		t.Fatal("month matched before its last day")
	}
	query.FromDate = "2028-02-30"
	if query.Validate() == nil {
		t.Fatal("invalid calendar range accepted")
	}
}

func queryMatches(t *testing.T, q Query, date expirationdate.Date, typeID notification.AssetTypeID, tags []string, settings notification.Settings, now time.Time) bool {
	t.Helper()
	matched, err := q.Matches(date, typeID, tags, settings, now)
	if err != nil {
		t.Fatal(err)
	}
	return matched
}
