package expirationdate_test

import (
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
)

func TestDatePreservesPrecisionAndCalendarBoundary(t *testing.T) {
	zone, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		value     string
		precision expirationdate.Precision
		boundary  string
	}{
		{"2028-02", expirationdate.Month, "2028-03-01T00:00:00-05:00"},
		{"2027-02", expirationdate.Month, "2027-03-01T00:00:00-05:00"},
		{"2027-03-14", expirationdate.Day, "2027-03-15T00:00:00-04:00"},
		{"2027-12-31", expirationdate.Day, "2028-01-01T00:00:00-05:00"},
	} {
		t.Run(tc.value, func(t *testing.T) {
			date, err := expirationdate.ParseDate(tc.value, tc.precision)
			if err != nil {
				t.Fatal(err)
			}
			if date.Value() != tc.value || date.Precision() != tc.precision {
				t.Fatal("entered date precision was lost")
			}
			if got := date.Boundary(zone).Format(time.RFC3339); got != tc.boundary {
				t.Fatalf("boundary %s, want %s", got, tc.boundary)
			}
			if date.StateAt(date.Boundary(zone).Add(-time.Nanosecond), zone, 0) == expirationdate.Expired {
				t.Fatal("expired before date ended")
			}
			if date.StateAt(date.Boundary(zone), zone, 0) != expirationdate.Expired {
				t.Fatal("not expired at boundary")
			}
		})
	}
}

func TestDateRejectsInvalidOrMismatchedPrecision(t *testing.T) {
	for _, tc := range []struct {
		value     string
		precision expirationdate.Precision
	}{
		{"", expirationdate.Day}, {"2027-02-29", expirationdate.Day}, {"2027-13", expirationdate.Month},
		{"2027-02-30", expirationdate.Day}, {"2027-2-01", expirationdate.Day}, {"2027-02", expirationdate.Day},
		{"2027-02-01", expirationdate.Month}, {"2027-02-01T00:00:00Z", expirationdate.Day},
		{"2027-02-01", "week"}, {"0000-01", expirationdate.Month},
	} {
		if _, err := expirationdate.ParseDate(tc.value, tc.precision); err == nil {
			t.Errorf("accepted %q (%s)", tc.value, tc.precision)
		}
	}
}

func TestUpcomingUsesCalendarDaysAcrossDaylightSaving(t *testing.T) {
	zone, _ := time.LoadLocation("America/New_York")
	date, err := expirationdate.ParseDate("2027-03-15", expirationdate.Day)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2027, 3, 14, 0, 0, 0, 0, zone)
	if date.StateAt(start.Add(-time.Nanosecond), zone, 1) != expirationdate.Current {
		t.Fatal("warning started early")
	}
	if date.StateAt(start, zone, 1) != expirationdate.Upcoming {
		t.Fatal("calendar day warning missing")
	}
}

func TestMidnightClockTransitionsDoNotShiftExpirationDates(t *testing.T) {
	for _, tc := range []struct{ zone, value, boundary string }{
		{"America/Goose_Bay", "2009-10-31", "2009-11-01T00:00:00-03:00"},
		{"America/Sao_Paulo", "2018-11-04", "2018-11-05T00:00:00-02:00"},
		{"America/Santiago", "2019-09-07", "2019-09-08T01:00:00-03:00"},
	} {
		zone, err := time.LoadLocation(tc.zone)
		if err != nil {
			t.Fatal(err)
		}
		date, err := expirationdate.ParseDate(tc.value, expirationdate.Day)
		if err != nil {
			t.Fatal(err)
		}
		if got := date.Boundary(zone).Format(time.RFC3339); got != tc.boundary {
			t.Errorf("%s: got %s want %s", tc.zone, got, tc.boundary)
		}
	}
}
