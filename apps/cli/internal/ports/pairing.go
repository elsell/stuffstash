package ports

import (
	"context"
	"time"
)

type PairingCandidate struct{ ID, Name, AdapterID, DeviceID string }
type PairingRequest struct {
	Rotation   bool
	Name       string
	PublicKey  []byte
	Candidates []PairingCandidate
}
type PairingChallenge struct {
	ID, PollToken, UserCode, VerificationURL string
	ExpiresAt                                time.Time
}
type PairingState string

const (
	PairingPending  PairingState = "pending"
	PairingApproved PairingState = "approved"
	PairingConsumed PairingState = "consumed"
)

type PairingAPI interface {
	Start(context.Context, PairingRequest) (PairingChallenge, error)
	Poll(context.Context, PairingChallenge) (PairingState, error)
	Exchange(context.Context, PairingChallenge, []byte) (ConnectorRegistration, error)
	Activate(context.Context, ConnectorRegistration, string) error
}
type PairingKey interface {
	PublicKey() []byte
	Sign(string, string) []byte
}
type PairingKeys interface{ NewKey() (PairingKey, error) }
type Waiter interface {
	Wait(context.Context, time.Duration) error
}
