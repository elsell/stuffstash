package httpapi

import (
	"bytes"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"time"
)

// Share the complete unsigned consumer model while retaining the worker's
// timestamp parsing and nullable-object semantics at this transport boundary.
type workerAttempt struct {
	ports.ConsumerAttempt
	null           bool
	LeaseExpiresAt time.Time            `json:"leaseExpiresAt"`
	StartedAt      *time.Time           `json:"startedAt,omitempty"`
	SettledAt      *time.Time           `json:"settledAt,omitempty"`
	ResolvedAt     *time.Time           `json:"resolvedAt,omitempty"`
	Media          *ports.ConsumerMedia `json:"media,omitempty"`
	Artifact       *workerArtifact      `json:"artifact,omitempty"`
}
type workerArtifact struct {
	ports.ConsumerArtifact
	ExpiresAt time.Time `json:"expiresAt"`
}
type workerProof struct {
	SessionID  string `json:"sessionId"`
	ClaimToken string `json:"claimToken"`
	Revision   uint64 `json:"revision"`
}
type workerOutcome struct {
	workerProof
	Outcome generated.PrintOutcome `json:"outcome"`
}
type workerReconciliation struct {
	Revision uint64                 `json:"revision"`
	Outcome  generated.PrintOutcome `json:"outcome"`
}
type workerRevision struct {
	Revision uint64 `json:"revision"`
}
type workerHeartbeat struct {
	SessionID string                      `json:"sessionId"`
	Report    *ports.PrintConnectorReport `json:"report,omitempty"`
}

func workerRequest(value any) (io.Reader, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return nil, ports.Failure("protocol", "Could not encode the print-worker request.")
	}
	return bytes.NewReader(body), nil
}

// Decode the safe public projection separately from operational timestamp fields.
// Both use the same body; credentials and unknown fields are never retained.
func (v *workerAttempt) UnmarshalJSON(body []byte) error {
	type operational workerAttempt
	var parsed operational
	if err := json.Unmarshal(body, &parsed); err != nil {
		return err
	}
	var safe ports.ConsumerAttempt
	if err := json.Unmarshal(body, &safe); err != nil {
		return err
	}
	*v = workerAttempt(parsed)
	v.ConsumerAttempt = safe
	v.null = bytes.Equal(bytes.TrimSpace(body), []byte("null"))
	return nil
}
