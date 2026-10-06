package ports

import (
	"context"
	"time"
)

// SafePairingStarted deliberately omits the secret polling token.
type SafePairingStarted struct {
	ID              string    `json:"id"`
	UserCode        string    `json:"userCode"`
	VerificationURL string    `json:"verificationUrl"`
	ExpiresAt       time.Time `json:"expiresAt"`
}
type SafePairingStatus struct {
	ID        string    `json:"id"`
	State     string    `json:"state"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// SafePairingCredential describes registration without its machine credential.
type SafePairingCredential struct {
	ConnectorID        string    `json:"connectorId"`
	TenantID           string    `json:"tenantId"`
	InventoryID        string    `json:"inventoryId"`
	ExpiresAt          time.Time `json:"expiresAt"`
	ActivationDeadline time.Time `json:"activationDeadline"`
}
type PairingStartedReceipt Result[SafePairingStarted]
type PairingStatusReceipt Result[SafePairingStatus]
type PairingCredentialReceipt Result[SafePairingCredential]

func (PairingStartedReceipt) protocolReceiptResult()    {}
func (PairingStatusReceipt) protocolReceiptResult()     {}
func (PairingCredentialReceipt) protocolReceiptResult() {}

// ReceiptDrain waits only until the supplied deadline for pending display output.
type ReceiptDrain interface{ Close(context.Context) bool }
