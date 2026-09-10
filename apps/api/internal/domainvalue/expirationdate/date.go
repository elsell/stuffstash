// Package expirationdate defines calendar-based expiration independently of storage and delivery.
package expirationdate

import (
	"errors"
	"time"
)

type Precision string

const (
	Day   Precision = "day"
	Month Precision = "month"
)

type State string

const (
	Unset    State = "unset"
	Current  State = "current"
	Upcoming State = "upcoming"
	Expired  State = "expired"
)

var ErrInvalidDate = errors.New("invalid expiration date")

// Date preserves the precision printed on the original label.
type Date struct {
	value     string
	precision Precision
	calendar  time.Time
}

func ParseDate(value string, precision Precision) (Date, error) {
	layout := "2006-01-02"
	switch precision {
	case Day:
	case Month:
		layout = "2006-01"
	default:
		return Date{}, ErrInvalidDate
	}
	calendar, err := time.Parse(layout, value)
	if err != nil || calendar.Year() < 1 || calendar.Format(layout) != value {
		return Date{}, ErrInvalidDate
	}
	return Date{value: value, precision: precision, calendar: calendar}, nil
}

func (d Date) Value() string        { return d.value }
func (d Date) Precision() Precision { return d.precision }

// Boundary is the first instant after the recorded day or month in the recipient's timezone.
func (d Date) Boundary(zone *time.Location) time.Time {
	if d.value == "" {
		return time.Time{}
	}
	return calendarStart(d.endCalendar(), zone)
}

func (d Date) endCalendar() time.Time {
	if d.precision == Month {
		return d.calendar.AddDate(0, 1, 0)
	}
	return d.calendar.AddDate(0, 0, 1)
}

// calendarStart examines timezone intervals in chronological order. Local dates
// can reverse across midnight rollbacks, so binary search on local date is unsafe.
func calendarStart(calendar time.Time, zone *time.Location) time.Time {
	if zone == nil {
		zone = time.UTC
	}
	const calendarSearchRadius = 48 * time.Hour
	cursor := calendar.Add(-calendarSearchRadius).In(zone)
	limit := calendar.Add(calendarSearchRadius).In(zone)
	for cursor.Before(limit) {
		_, offset := cursor.Zone()
		_, end := cursor.ZoneBounds()
		if end.IsZero() || end.After(limit) {
			end = limit
		}
		candidate := time.Unix(calendar.Unix()-int64(offset), 0).In(zone)
		if candidate.Before(cursor) {
			candidate = cursor
		}
		if candidate.Before(end) {
			return candidate
		}
		cursor = end
	}
	// IANA offsets and civil-date jumps lie inside the bounded interval above.
	return limit
}

func (d Date) StateAt(now time.Time, zone *time.Location, advanceDays int) State {
	if d.value == "" {
		return Unset
	}
	if !now.Before(d.Boundary(zone)) {
		return Expired
	}
	// The warning window begins N calendar days before the last valid date,
	// including that date itself when N is zero. Calculate in neutral calendar space.
	if advanceDays < 0 {
		advanceDays = 0
	}
	start := calendarStart(d.endCalendar().AddDate(0, 0, -advanceDays-1), zone)
	if !now.Before(start) {
		return Upcoming
	}
	return Current
}
