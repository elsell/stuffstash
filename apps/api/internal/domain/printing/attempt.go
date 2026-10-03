package printing

import (
	"crypto/subtle"
	"time"
)

type OutcomeKind string
type OutcomeReason string

const (
	OutcomeCompleted        OutcomeKind   = "completed"
	OutcomeNoOutput         OutcomeKind   = "no_output"
	OutcomeUncertain        OutcomeKind   = "uncertain"
	ReasonLeaseExpired      OutcomeReason = "lease_expired"
	ReasonCanceled          OutcomeReason = "canceled"
	ReasonDeviceUnavailable OutcomeReason = "device_unavailable"
	ReasonInvalidArtifact   OutcomeReason = "invalid_artifact"
	ReasonDeviceFailure     OutcomeReason = "device_failure"
	ReasonPartialOutput     OutcomeReason = "partial_output"
	ReasonUnknown           OutcomeReason = "unknown"
)

type AttemptAuthority struct {
	AttemptID   AttemptID
	ConnectorID ConnectorID
	SessionID   SessionID
	TokenDigest [32]byte
}

func (a AttemptAuthority) Valid() bool {
	return a.AttemptID != "" && a.ConnectorID != "" && a.SessionID != "" && a.TokenDigest != [32]byte{}
}
func (a AttemptAuthority) Equal(b AttemptAuthority) bool {
	return a.AttemptID == b.AttemptID && a.ConnectorID == b.ConnectorID && a.SessionID == b.SessionID && subtle.ConstantTimeCompare(a.TokenDigest[:], b.TokenDigest[:]) == 1
}

type Outcome struct {
	Kind            OutcomeKind
	CompletedCopies int
	Retryable       bool
	Reason          OutcomeReason
}

func (o Outcome) Valid(copies int) bool {
	if copies <= 0 || o.CompletedCopies < 0 || o.CompletedCopies > copies {
		return false
	}
	switch o.Kind {
	case OutcomeCompleted:
		return o.CompletedCopies == copies && !o.Retryable
	case OutcomeNoOutput:
		return o.CompletedCopies == 0 && o.Reason != ReasonPartialOutput
	case OutcomeUncertain:
		return !o.Retryable
	default:
		return false
	}
}

type Attempt struct {
	ID                                              AttemptID
	Authority                                       AttemptAuthority
	ClaimedAt, LeaseExpiresAt, StartedAt, SettledAt time.Time
	Outcome                                         Outcome
}
