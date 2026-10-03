package printing

import "time"

type PairingID string
type PairingState string

const (
	PairingPending  PairingState = "pending"
	PairingApproved PairingState = "approved"
	PairingConsumed PairingState = "consumed"
)

type Pairing struct {
	ID                   PairingID
	Name                 string
	PublicKey            []byte
	PollHash, CodeHash   string
	State                PairingState
	Scope                Scope
	ConnectorID          ConnectorID
	ExpiresAt, CreatedAt time.Time
}

func PairingProofMessage(id PairingID, pollToken string) []byte {
	return []byte("stuffstash-print-pairing-v1\n" + string(id) + "\n" + pollToken)
}
