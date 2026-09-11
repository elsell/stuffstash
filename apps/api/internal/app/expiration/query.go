package expiration

import (
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"slices"
	"strings"
	"time"
)

type QueryStatus string

const (
	QueryUpcoming QueryStatus = "upcoming"
	QueryExpired  QueryStatus = "expired"
	QueryAll      QueryStatus = "all"
)

var ErrInvalidQuery = errors.New("invalid expiration query")

type Query struct {
	Status      QueryStatus              `json:"status"`
	TypeID      notification.AssetTypeID `json:"customAssetTypeId,omitempty"`
	TagKey      string                   `json:"tagKey,omitempty"`
	FromDate    string                   `json:"fromDate,omitempty"`
	ThroughDate string                   `json:"throughDate,omitempty"`
}

func (q Query) Validate() error {
	if q.Status != QueryUpcoming && q.Status != QueryExpired && q.Status != QueryAll {
		return ErrInvalidQuery
	}
	if len(q.TypeID) > 128 || strings.TrimSpace(string(q.TypeID)) != string(q.TypeID) || len(q.TagKey) > 80 || strings.TrimSpace(q.TagKey) != q.TagKey {
		return ErrInvalidQuery
	}
	for _, date := range []string{q.FromDate, q.ThroughDate} {
		if date != "" {
			if _, err := expirationdate.ParseDate(date, expirationdate.Day); err != nil {
				return ErrInvalidQuery
			}
		}
	}
	if q.FromDate != "" && q.ThroughDate != "" && q.FromDate > q.ThroughDate {
		return ErrInvalidQuery
	}
	return nil
}
func (q Query) Matches(date expirationdate.Date, typeID notification.AssetTypeID, tags []string, settings notification.Settings, now time.Time) (bool, error) {
	if err := q.Validate(); err != nil {
		return false, err
	}
	description, err := Describe(date, typeID, true, settings, now)
	if err != nil {
		return false, err
	}
	if description.State == expirationdate.Unset {
		return false, nil
	}
	if q.TypeID != "" || q.TagKey != "" {
		if !(q.TypeID != "" && typeID == q.TypeID) && !(q.TagKey != "" && slices.Contains(tags, q.TagKey)) {
			return false, nil
		}
	}
	last := date.LastValidDate()
	if q.FromDate != "" && last < q.FromDate || q.ThroughDate != "" && last > q.ThroughDate {
		return false, nil
	}
	return q.Status == QueryAll || q.Status == QueryUpcoming && description.State == expirationdate.Upcoming || q.Status == QueryExpired && description.State == expirationdate.Expired, nil
}
