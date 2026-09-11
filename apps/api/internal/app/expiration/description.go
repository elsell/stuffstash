package expiration

import (
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"time"
)

type Description struct {
	Date            expirationdate.Date
	State           expirationdate.State
	TrackingEnabled bool
	AdvanceDays     int
	Timezone        string
}

// Describe evaluates calendar facts independently of notification delivery switches.
func Describe(date expirationdate.Date, typeID notification.AssetTypeID, enabled bool, settings notification.Settings, now time.Time) (Description, error) {
	if err := settings.Validate(); err != nil {
		return Description{}, err
	}
	zone, err := time.LoadLocation(settings.Timezone)
	if err != nil {
		return Description{}, err
	}
	days := settings.ForType(typeID).AdvanceDays
	return Description{Date: date, State: date.StateAt(now, zone, days), TrackingEnabled: enabled, AdvanceDays: days, Timezone: settings.Timezone}, nil
}
