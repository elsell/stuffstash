package printing

import "time"

type PairingID string
type PairingState string

const (
	PairingPending  PairingState = "pending"
	PairingApproved PairingState = "approved"
	PairingConsumed PairingState = "consumed"
)

type PairingCandidate struct {
	ID, Name, AdapterID, DeviceID string
}

type Pairing struct {
	RotationVersion      uint64
	Rotation             bool
	Candidates           []PairingCandidate
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

// Approve is evaluated under the pairing transaction lock after checking the
// human's inventory.configure permission. Approval cannot rebind an old request.
func (p *Pairing) Approve(scope Scope, connectorID ConnectorID, now time.Time) bool {
	if p.State != PairingPending || !p.ExpiresAt.After(now) || scope.TenantID == "" || scope.InventoryID == "" || connectorID == "" {
		return false
	}
	p.Scope, p.ConnectorID, p.State = scope, connectorID, PairingApproved
	return true
}

// Consume runs atomically with credential issuance, after proof validation and
// relationship delivery. A lost successful response requires a new pairing.
func (p *Pairing) Consume(now time.Time) bool {
	if p.State != PairingApproved || !p.ExpiresAt.After(now) {
		return false
	}
	p.State = PairingConsumed
	return true
}
