package printing

import (
	"errors"
	"time"
)

type JobID string
type AttemptID string
type SessionID string
type JobStatus string
type JobKind string

const (
	JobQueued      JobStatus = "queued"
	JobClaimed     JobStatus = "claimed"
	JobPrinting    JobStatus = "printing"
	JobCompleted   JobStatus = "completed"
	JobFailed      JobStatus = "failed"
	JobUncertain   JobStatus = "uncertain"
	JobCanceled    JobStatus = "canceled"
	JobAssetLabel  JobKind   = "asset_label"
	JobPrinterTest JobKind   = "printer_test"
)

var (
	ErrJobConflict      = errors.New("print job state or revision conflict")
	ErrAttemptOwnership = errors.New("print attempt ownership mismatch")
	ErrLeaseExpired     = errors.New("print attempt lease expired")
	ErrInvalidOutcome   = errors.New("invalid print outcome evidence")
)

type Artifact struct {
	ExpiresAt                 time.Time
	Key, SHA256, ContentType  string
	ByteLength                int64
	WidthPixels, HeightPixels int
}

type Job struct {
	AssetCreationOperationID                             string
	ID                                                   JobID
	Scope                                                Scope
	PrinterID                                            PrinterID
	Kind                                                 JobKind
	AssetID, LabelReference, RequestedBy, IdempotencyKey string
	Predecessor                                          JobID
	Media                                                MediaSnapshot
	MediaFingerprint                                     string
	Template                                             TemplateSelection
	Content                                              ContentSnapshot
	Artifact                                             Artifact
	Copies                                               int
	Status                                               JobStatus
	Revision                                             uint64
	Attempts                                             []Attempt
	CreatedAt, UpdatedAt                                 time.Time
}

func (j *Job) current() *Attempt {
	if len(j.Attempts) == 0 {
		return nil
	}
	return &j.Attempts[len(j.Attempts)-1]
}
func (j *Job) changed(now time.Time) { j.Revision++; j.UpdatedAt = now }
func (j *Job) Claim(id AttemptID, owner AttemptAuthority, now time.Time, lease time.Duration, revision uint64) error {
	if j.Status != JobQueued || j.Revision != revision || id == "" || owner.AttemptID != id || !owner.Valid() || lease <= 0 {
		return ErrJobConflict
	}
	for _, a := range j.Attempts {
		if a.ID == id {
			return ErrJobConflict
		}
	}
	j.Attempts = append(j.Attempts, Attempt{ID: id, Authority: owner, LeaseExpiresAt: now.Add(lease), ClaimedAt: now})
	j.Status = JobClaimed
	j.changed(now)
	return nil
}
func (j *Job) authorize(owner AttemptAuthority, now time.Time) error {
	a := j.current()
	if a == nil || !a.Authority.Equal(owner) {
		return ErrAttemptOwnership
	}
	if !now.Before(a.LeaseExpiresAt) {
		return ErrLeaseExpired
	}
	return nil
}
func (j *Job) Start(owner AttemptAuthority, now time.Time, revision uint64) error {
	if err := j.authorize(owner, now); err != nil {
		return err
	}
	if j.Status != JobClaimed || j.Revision != revision {
		return ErrJobConflict
	}
	j.current().StartedAt = now
	j.Status = JobPrinting
	j.changed(now)
	return nil
}
func (j *Job) Renew(owner AttemptAuthority, now time.Time, lease time.Duration, revision uint64) error {
	if err := j.authorize(owner, now); err != nil {
		return err
	}
	if (j.Status != JobClaimed && j.Status != JobPrinting) || j.Revision != revision || lease <= 0 {
		return ErrJobConflict
	}
	j.current().LeaseExpiresAt = now.Add(lease)
	j.changed(now)
	return nil
}
func (j *Job) Expire(now time.Time) bool {
	a := j.current()
	if a == nil || now.Before(a.LeaseExpiresAt) {
		return false
	}
	switch j.Status {
	case JobClaimed:
		j.Status = JobQueued
		a.Outcome = Outcome{Kind: OutcomeNoOutput, Retryable: true, Reason: ReasonLeaseExpired}
		a.SettledAt = now
	case JobPrinting:
		j.Status = JobUncertain
		a.Outcome = Outcome{Kind: OutcomeUncertain, Reason: ReasonLeaseExpired}
	default:
		return false
	}
	j.changed(now)
	return true
}
func (j *Job) Cancel(now time.Time, revision uint64) error {
	if j.Revision != revision || (j.Status != JobQueued && j.Status != JobClaimed) {
		return ErrJobConflict
	}
	if a := j.current(); a != nil && j.Status == JobClaimed {
		a.Outcome = Outcome{Kind: OutcomeNoOutput, Reason: ReasonCanceled}
		a.SettledAt = now
	}
	j.Status = JobCanceled
	j.changed(now)
	return nil
}

func (j *Job) Report(owner AttemptAuthority, now time.Time, revision uint64, outcome Outcome) error {
	a := j.current()
	if a == nil || !a.Authority.Equal(owner) {
		return ErrAttemptOwnership
	}
	// Repeated settled reports only read previously recorded evidence. They never
	// extend a lease, create another attempt, or start output.
	if a.Outcome.Kind != "" && a.Outcome == outcome {
		return nil
	}
	if err := j.authorize(owner, now); err != nil {
		return err
	}
	if j.Revision != revision || (j.Status != JobClaimed && j.Status != JobPrinting) {
		return ErrJobConflict
	}
	if !outcome.Valid(j.Copies) || (j.Status == JobClaimed && outcome.Kind != OutcomeNoOutput) {
		return ErrInvalidOutcome
	}
	j.apply(outcome, now)
	return nil
}
func (j *Job) Reconcile(attemptID AttemptID, connector ConnectorID, outcome Outcome, now time.Time, revision uint64) error {
	a := j.current()
	if a == nil || a.ID != attemptID || a.Authority.ConnectorID != connector {
		return ErrAttemptOwnership
	}
	if a.Outcome.Kind != "" && a.Outcome == outcome {
		return nil
	}
	if j.Revision != revision || j.Status != JobUncertain {
		return ErrJobConflict
	}
	if !outcome.Valid(j.Copies) || outcome.Kind == OutcomeUncertain || !preservesOutputEvidence(a.Outcome, outcome) {
		return ErrInvalidOutcome
	}
	j.apply(outcome, now)
	return nil
}
func (j *Job) apply(outcome Outcome, now time.Time) {
	a := j.current()
	a.Outcome = outcome
	switch outcome.Kind {
	case OutcomeCompleted:
		j.Status = JobCompleted
		a.SettledAt = now
	case OutcomeNoOutput:
		j.Status = JobFailed
		if outcome.Retryable {
			j.Status = JobQueued
		}
		a.SettledAt = now
	case OutcomeUncertain:
		j.Status = JobUncertain
	}
	j.changed(now)
}

// HoldsReservation remains true after ambiguous physical output. Clearing it
// requires reconciliation or an explicit safe acknowledgement command.
func (j Job) HoldsReservation() bool {
	return j.Status == JobClaimed || j.Status == JobPrinting || j.Status == JobUncertain
}

func preservesOutputEvidence(previous, next Outcome) bool {
	if next.CompletedCopies < previous.CompletedCopies {
		return false
	}
	return next.Kind != OutcomeNoOutput || (previous.CompletedCopies == 0 && previous.Reason != ReasonPartialOutput)
}
