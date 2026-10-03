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
	PendingCredentialHash                                 string
	PendingCredentialVersion                              uint64
	PendingCredentialExpiresAt, PendingActivationDeadline time.Time
	PendingPublicKey                                      []byte
	ID                                                    ConnectorID
	Scope                                                 Scope
	ServiceAccountID                                      ServiceAccountID
	Name                                                  string
	State                                                 ConnectorState
	PublicKey                                             []byte
	CredentialHash                                        string
	CredentialVersion                                     uint64
	CredentialExpiresAt, ActivationDeadline               time.Time
	Generation, SyncedGeneration                          uint64
	LastSeenAt                                            *time.Time
	CreatedAt, UpdatedAt                                  time.Time
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

// PendingCredentialIdentity is an authenticated-but-not-activated projection.
// Printer deny fences still compare its version with the active registration.
func (c Connector) PendingCredentialIdentity() Connector {
	c.CredentialHash = c.PendingCredentialHash
	c.CredentialVersion = c.PendingCredentialVersion
	c.CredentialExpiresAt = c.PendingCredentialExpiresAt
	c.ActivationDeadline = c.PendingActivationDeadline
	c.PublicKey = append([]byte(nil), c.PendingPublicKey...)
	c.State = ConnectorAwaitingActivation
	return c
}
func (c *Connector) ActivatePending() {
	c.CredentialHash = c.PendingCredentialHash
	c.CredentialVersion = c.PendingCredentialVersion
	c.CredentialExpiresAt = c.PendingCredentialExpiresAt
	c.PublicKey = append([]byte(nil), c.PendingPublicKey...)
	c.PendingCredentialHash = ""
	c.PendingCredentialVersion = 0
	c.PendingCredentialExpiresAt = time.Time{}
	c.PendingActivationDeadline = time.Time{}
	c.PendingPublicKey = nil
	c.State = ConnectorActive
}
