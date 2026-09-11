package notification

import (
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"time"
)

type MilestoneKind string

const (
	MilestoneUpcoming MilestoneKind = "upcoming"
	MilestoneExpired  MilestoneKind = "expired"
)

type Milestone struct {
	AssetID string
	Date    expirationdate.Date
	Kind    MilestoneKind
}

// ExpirationCandidate is a projection; eligibility is resolved by the application
// from current inventory, asset and type lifecycle state before evaluation.
type ExpirationCandidate struct {
	AssetID  string
	TypeID   AssetTypeID
	Date     expirationdate.Date
	Eligible bool
}

func (c ExpirationCandidate) Due(settings Settings, now time.Time, zone *time.Location) (Milestone, bool) {
	policy := settings.ForType(c.TypeID)
	if !c.Eligible || c.AssetID == "" || !policy.Enabled {
		return Milestone{}, false
	}
	value := Milestone{AssetID: c.AssetID, Date: c.Date}
	switch c.Date.StateAt(now, zone, policy.AdvanceDays) {
	case expirationdate.Upcoming:
		if !policy.Upcoming {
			return Milestone{}, false
		}
		value.Kind = MilestoneUpcoming
	case expirationdate.Expired:
		if !policy.Expired {
			return Milestone{}, false
		}
		value.Kind = MilestoneExpired
	default:
		return Milestone{}, false
	}
	return value, true
}

// Visible ignores personal delivery switches so turning reminders off does not
// erase history. Calendar and lifecycle changes still withdraw stale entries.
func (c ExpirationCandidate) Visible(value Milestone, settings Settings, now time.Time, zone *time.Location) bool {
	if !c.Eligible || c.AssetID == "" || value.AssetID != c.AssetID || value.Date != c.Date {
		return false
	}
	state := c.Date.StateAt(now, zone, settings.ForType(c.TypeID).AdvanceDays)
	switch value.Kind {
	case MilestoneUpcoming:
		return state == expirationdate.Upcoming || state == expirationdate.Expired
	case MilestoneExpired:
		return state == expirationdate.Expired
	default:
		return false
	}
}
