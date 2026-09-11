// Package notification defines personal notification preferences and inbox state.
package notification

import "errors"

const MaxAdvanceDays = 3650

var ErrInvalidPreferences = errors.New("invalid notification preferences")

// ExpirationPreferences applies either to an inventory default or to a complete
// type override. A missing override is represented by nil, not disabled values.
type ExpirationPreferences struct {
	Enabled     bool
	Upcoming    bool
	Expired     bool
	AdvanceDays int
}

func DefaultExpirationPreferences() ExpirationPreferences {
	return ExpirationPreferences{Enabled: true, Upcoming: true, Expired: true, AdvanceDays: 30}
}

func (p ExpirationPreferences) Validate() error {
	if p.AdvanceDays < 0 || p.AdvanceDays > MaxAdvanceDays {
		return ErrInvalidPreferences
	}
	return nil
}

func ResolveExpirationPreferences(defaults ExpirationPreferences, override *ExpirationPreferences) ExpirationPreferences {
	if override != nil {
		return *override
	}
	return defaults
}
