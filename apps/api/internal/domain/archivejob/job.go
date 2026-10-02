// Package archivejob owns durable portable-inventory task transitions.
package archivejob

import (
	"errors"
	"strings"
	"time"
)

var ErrTransition = errors.New("archive job transition is not allowed")
var ErrInvalid = errors.New("invalid archive job")

type Kind string

const (
	Export  Kind = "export"
	Restore Kind = "restore"
)

type State string

const (
	Queued           State = "queued"
	Running          State = "running"
	AwaitingApproval State = "awaiting_approval"
	Ready            State = "ready"
	Failed           State = "failed"
	Cancelled        State = "cancelled"
	Expired          State = "expired"
)

type Phase string

const (
	Validation Phase = "validation"
	Execution  Phase = "execution"
)

type Failure string

const (
	FailureInvalidArchive Failure = "invalid_archive"
	FailureStorage        Failure = "storage_unavailable"
	FailurePermission     Failure = "permission_changed"
	FailureLimit          Failure = "limit_exceeded"
	FailureConflict       Failure = "publication_conflict"
	FailureInternal       Failure = "internal"
)

type Request struct {
	RequestKey        string
	ID                string
	TenantID          string
	PrincipalID       string
	Kind              Kind
	SourceInventoryID string
	SourceArtifactID  string
	SourceSHA256      string
	Photos            bool
	OtherFiles        bool
}
type Record struct {
	Request                `json:"-"`
	State                  State
	Phase                  Phase
	Revision               int64
	CreatedAt              time.Time
	UpdatedAt              time.Time
	ExpiresAt              time.Time
	LeaseToken             string
	LeaseUntil             time.Time
	Attempts               int
	DestinationInventoryID string
	DestinationName        string
	ResultArtifactID       string
	Failure                Failure
}

func New(request Request, now, expires time.Time) (Record, error) {
	if strings.TrimSpace(request.RequestKey) == "" || strings.TrimSpace(request.ID) == "" || strings.TrimSpace(request.TenantID) == "" || strings.TrimSpace(request.PrincipalID) == "" || now.IsZero() || !expires.After(now) {
		return Record{}, ErrInvalid
	}
	phase := Execution
	switch request.Kind {
	case Export:
		if strings.TrimSpace(request.SourceInventoryID) == "" || request.SourceArtifactID != "" || request.SourceSHA256 != "" {
			return Record{}, ErrInvalid
		}
	case Restore:
		if strings.TrimSpace(request.SourceArtifactID) == "" || !validHash(request.SourceSHA256) || request.SourceInventoryID != "" {
			return Record{}, ErrInvalid
		}
		phase = Validation
	default:
		return Record{}, ErrInvalid
	}
	return Record{Request: request, State: Queued, Phase: phase, Revision: 1, CreatedAt: now.UTC().Truncate(time.Microsecond), UpdatedAt: now.UTC().Truncate(time.Microsecond), ExpiresAt: expires.UTC().Truncate(time.Microsecond)}, nil
}
func (r Record) Claim(token string, now, until time.Time) (Record, error) {
	if !r.current(now) || strings.TrimSpace(token) == "" || token == r.LeaseToken || !until.After(now) || until.After(r.ExpiresAt) {
		return Record{}, ErrTransition
	}
	if r.State != Queued && (r.State != Running || r.LeaseUntil.After(now)) {
		return Record{}, ErrTransition
	}
	if r.Kind == Restore && r.Phase == Execution && (r.DestinationInventoryID == "" || r.DestinationName == "") {
		return Record{}, ErrTransition
	}
	r.State = Running
	r.LeaseToken = token
	r.LeaseUntil = until.UTC().Truncate(time.Microsecond)
	r.Attempts++
	return r.changed(now), nil
}
func (r Record) Heartbeat(token string, now, until time.Time) (Record, error) {
	if !r.owns(token, now) || !until.After(r.LeaseUntil) || until.After(r.ExpiresAt) {
		return Record{}, ErrTransition
	}
	r.LeaseUntil = until.UTC().Truncate(time.Microsecond)
	return r.changed(now), nil
}
func (r Record) PreviewReady(token string, now time.Time) (Record, error) {
	if !r.owns(token, now) || r.Kind != Restore || r.Phase != Validation {
		return Record{}, ErrTransition
	}
	r.State = AwaitingApproval
	r.clearLease()
	return r.changed(now), nil
}
func (r Record) Approve(inventoryID, name string, now time.Time) (Record, error) {
	if !r.current(now) || r.State != AwaitingApproval || r.Kind != Restore || strings.TrimSpace(inventoryID) == "" || strings.TrimSpace(name) == "" {
		return Record{}, ErrTransition
	}
	r.DestinationInventoryID = inventoryID
	r.DestinationName = name
	r.Phase = Execution
	r.State = Queued
	return r.changed(now), nil
}
func (r Record) Complete(token string, now time.Time, artifactID string) (Record, error) {
	if !r.owns(token, now) || r.Phase != Execution {
		return Record{}, ErrTransition
	}
	if r.Kind == Export {
		if strings.TrimSpace(artifactID) == "" {
			return Record{}, ErrInvalid
		}
		r.ResultArtifactID = artifactID
	} else if artifactID != "" || r.DestinationInventoryID == "" {
		return Record{}, ErrInvalid
	}
	r.State = Ready
	r.clearLease()
	return r.changed(now), nil
}
func (r Record) Fail(token string, now time.Time, failure Failure) (Record, error) {
	if !r.owns(token, now) || !validFailure(failure) {
		return Record{}, ErrTransition
	}
	r.State = Failed
	r.Failure = failure
	r.clearLease()
	return r.changed(now), nil
}
func (r Record) Retry(now time.Time) (Record, error) {
	if !r.current(now) || r.State != Failed {
		return Record{}, ErrTransition
	}
	r.State = Queued
	r.Failure = ""
	return r.changed(now), nil
}
func (r Record) Cancel(now time.Time) (Record, error) {
	if !r.current(now) || (r.State != Queued && r.State != Running && r.State != AwaitingApproval && r.State != Failed) {
		return Record{}, ErrTransition
	}
	r.State = Cancelled
	r.clearLease()
	return r.changed(now), nil
}
func (r Record) Expire(now time.Time) (Record, error) {
	if now.Before(r.ExpiresAt) || now.Before(r.UpdatedAt) || r.State == Expired {
		return Record{}, ErrTransition
	}
	r.State = Expired
	r.clearLease()
	return r.changed(now), nil
}
func (r Record) current(now time.Time) bool {
	return !now.IsZero() && !now.Before(r.UpdatedAt) && now.Before(r.ExpiresAt)
}
func (r Record) owns(token string, now time.Time) bool {
	return r.current(now) && r.State == Running && token != "" && token == r.LeaseToken && r.LeaseUntil.After(now)
}
func (r Record) changed(now time.Time) Record {
	r.UpdatedAt = now.UTC().Truncate(time.Microsecond)
	r.Revision++
	return r
}
func (r *Record) clearLease() { r.LeaseToken = ""; r.LeaseUntil = time.Time{} }
func validHash(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func validFailure(f Failure) bool {
	switch f {
	case FailureInvalidArchive, FailureStorage, FailurePermission, FailureLimit, FailureConflict, FailureInternal:
		return true
	}
	return false
}

// ValidateSuccessor accepts only records produced by one domain transition.
// It prevents persistence callers from changing immutable request data or
// bypassing approval, lease ownership, and publication preconditions.
func ValidateSuccessor(previous, next Record) error {
	var expected Record
	var err error
	switch next.State {
	case Running:
		if previous.State == Running && previous.LeaseToken == next.LeaseToken {
			expected, err = previous.Heartbeat(next.LeaseToken, next.UpdatedAt, next.LeaseUntil)
		} else {
			expected, err = previous.Claim(next.LeaseToken, next.UpdatedAt, next.LeaseUntil)
		}
	case AwaitingApproval:
		expected, err = previous.PreviewReady(previous.LeaseToken, next.UpdatedAt)
	case Queued:
		if previous.State == AwaitingApproval {
			expected, err = previous.Approve(next.DestinationInventoryID, next.DestinationName, next.UpdatedAt)
		} else {
			expected, err = previous.Retry(next.UpdatedAt)
		}
	case Ready:
		expected, err = previous.Complete(previous.LeaseToken, next.UpdatedAt, next.ResultArtifactID)
	case Failed:
		expected, err = previous.Fail(previous.LeaseToken, next.UpdatedAt, next.Failure)
	case Cancelled:
		expected, err = previous.Cancel(next.UpdatedAt)
	case Expired:
		expected, err = previous.Expire(next.UpdatedAt)
	default:
		return ErrTransition
	}
	if err != nil {
		return err
	}
	if expected != next {
		return ErrTransition
	}
	return nil
}
