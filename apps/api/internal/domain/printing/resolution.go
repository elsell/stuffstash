package printing

import "time"

type ReportedOutcome string

const (
	ReportedPrinted    ReportedOutcome = "printed"
	ReportedNotPrinted ReportedOutcome = "not_printed"
	ReportedUnknown    ReportedOutcome = "unknown"
)

type Resolution struct {
	ReportedOutcome ReportedOutcome
	ResolvedBy      string
	ResolvedAt      time.Time
	Revision        uint64
}

func (j *Job) ConfirmIdle(id AttemptID, connector ConnectorID, now time.Time, revision uint64) error {
	a := j.current()
	if a == nil || a.ID != id || a.Authority.ConnectorID != connector {
		return ErrAttemptOwnership
	}
	if j.Status != JobUncertain {
		return ErrJobConflict
	}
	if !a.IdleConfirmedAt.IsZero() {
		return nil
	}
	if j.Revision != revision {
		return ErrJobConflict
	}
	a.IdleConfirmedAt = now
	j.changed(now)
	return nil
}
func (j *Job) Resolve(actor string, outcome ReportedOutcome, acknowledge bool, now time.Time, revision uint64) error {
	if !acknowledge || actor == "" || (outcome != ReportedPrinted && outcome != ReportedNotPrinted && outcome != ReportedUnknown) {
		return ErrInvalidOutcome
	}
	if j.Resolution != nil {
		if j.Resolution.ResolvedBy == actor && j.Resolution.ReportedOutcome == outcome && j.Resolution.Revision == revision {
			return nil
		}
		return ErrJobConflict
	}
	a := j.current()
	if j.Status != JobUncertain || j.Revision != revision || a == nil || a.IdleConfirmedAt.IsZero() {
		return ErrJobConflict
	}
	j.Resolution = &Resolution{ReportedOutcome: outcome, ResolvedBy: actor, ResolvedAt: now, Revision: revision}
	j.Status = JobFailed
	a.SettledAt = now
	j.changed(now)
	return nil
}
