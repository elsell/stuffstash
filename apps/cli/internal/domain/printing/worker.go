package printing

import "time"

type AttemptControl struct {
	AttemptID, SessionID, ClaimToken string
	Revision                         uint64
}
type Artifact struct {
	SHA256, ContentType string
	ByteLength          int64
	Width, Height       int
}
type Claim struct {
	Control          AttemptControl
	JobID, PrinterID string
	ContractVersion  int
	LeaseExpiresAt   time.Time
	Media            Media
	Copies           int
	Artifact         Artifact
}

// RemotePhase describes an attempt, not the potentially requeued parent job.
// The API adapter maps settled no-output attempts to RemoteFailed.
type RemotePhase string

const (
	RemoteClaimed   RemotePhase = "claimed"
	RemotePrinting  RemotePhase = "printing"
	RemoteUncertain RemotePhase = "uncertain"
	RemoteCompleted RemotePhase = "completed"
	RemoteFailed    RemotePhase = "failed"
	RemoteCanceled  RemotePhase = "canceled"
)

type AttemptStatus struct {
	AttemptID, SessionID, JobID string
	Phase                       RemotePhase
	Revision                    uint64
	LeaseExpiresAt              time.Time
	Outcome                     Outcome
	CompletedCopies             int
}
type Evidence struct {
	Outcome         Outcome
	CompletedCopies int
	Reason          Reason
}
