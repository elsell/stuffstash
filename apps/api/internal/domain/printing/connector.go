package printing

import "time"

type ConnectorState string

const (
	ConnectorPending            ConnectorState = "authorization_pending"
	ConnectorAwaitingActivation ConnectorState = "awaiting_activation"
	ConnectorActive             ConnectorState = "active"
	ConnectorRevoked            ConnectorState = "revoked"
)

type Connector struct {
	ID                                      ConnectorID
	Scope                                   Scope
	ServiceAccountID                        ServiceAccountID
	Name                                    string
	State                                   ConnectorState
	PublicKey                               []byte
	CredentialHash                          string
	CredentialVersion                       uint64
	CredentialExpiresAt, ActivationDeadline time.Time
	Generation, SyncedGeneration            uint64
	LastSeenAt                              *time.Time
	CreatedAt, UpdatedAt                    time.Time
}
type PrinterBinding struct {
	Scope                        Scope
	PrinterID                    PrinterID
	ConnectorID                  ConnectorID
	DeviceID                     string
	Generation, SyncedGeneration uint64
	Revoked                      bool
}
type ConsumerAuthority struct {
	Scope             Scope
	ConnectorID       ConnectorID
	ServiceAccountID  ServiceAccountID
	CredentialVersion uint64
	PrinterID         PrinterID
	BindingGeneration uint64
}
