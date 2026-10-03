package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"time"
)

// PairingSecrets owns randomness, token hashing and proof validation. Its inputs
// are secret and must never become audit metadata or observability attributes.
type PairingSecrets interface {
	NewToken() (string, error)
	NewUserCode() (string, error)
	Digest(string) string
	Matches(string, string) bool
	Verify([]byte, printing.PairingID, string, []byte) bool
}

type ConnectorRegistration struct {
	// Read projection only; not persisted with registration.
	Availability printing.ConnectorAvailability
	Connector    printing.Connector
	Bindings     []printing.PrinterBinding
}

type PairingApproval struct {
	PairingID    printing.PairingID
	CodeHash     string
	Now          time.Time
	Registration ConnectorRegistration
	Audit        audit.Record
}

// ConnectorAuthorizationSync executes while the connector's latest generation
// is locked, with no other transaction able to revoke/regrant concurrently.
type ConnectorAuthorizationSync func(context.Context, ConnectorRegistration, []printing.Printer) error

type ConnectorMutation func(*ConnectorRegistration) error
type ConnectorAudit func(ConnectorRegistration) (audit.Record, error)

type PrinterHealthReport struct {
	Connector printing.Connector
	Binding   printing.PrinterBinding
	Report    printing.PrinterReport
}

type RotationApproval struct {
	PairingID   printing.PairingID
	CodeHash    string
	Scope       printing.Scope
	ConnectorID printing.ConnectorID
	Generation  uint64
	Now         time.Time
	Audit       audit.Record
}

type PairingExchange struct {
	PairingID                               printing.PairingID
	Now                                     time.Time
	CredentialHash                          string
	CredentialExpiresAt, ActivationDeadline time.Time
	Audit                                   audit.Record
}
type ConnectorActivationAudit func(printing.Connector, bool) (audit.Record, error)

type ConnectorRepository interface {
	ApprovePrintCredentialRotation(context.Context, RotationApproval) (printing.Pairing, error)
	ListPrintPrinterHealth(context.Context, printing.Scope, printing.PrinterID) ([]PrinterHealthReport, error)
	UpdatePrintConnector(context.Context, printing.Scope, printing.ConnectorID, uint64, ConnectorMutation, ConnectorAudit) (ConnectorRegistration, error)
	CreatePrintPairing(context.Context, printing.Pairing) error
	GetPrintPairing(context.Context, printing.PairingID) (printing.Pairing, error)
	ApprovePrintPairing(context.Context, PairingApproval) (ConnectorRegistration, error)
	ConsumePrintPairing(context.Context, PairingExchange) (printing.Connector, error)
	GetPrintConnector(context.Context, printing.Scope, printing.ConnectorID) (ConnectorRegistration, error)
	ListPrintConnectors(context.Context, printing.Scope, int, string) ([]ConnectorRegistration, error)
	// Credential lookup is an authentication boundary, never a human resource read.
	FindPrintConnectorCredential(context.Context, string) (printing.Connector, error)
	SynchronizePrintConnector(context.Context, printing.Scope, printing.ConnectorID, ConnectorAuthorizationSync) error
	PendingPrintConnectorScopes(context.Context, int) ([]printing.Connector, error)
	HeartbeatPrintConnector(context.Context, printing.Connector, time.Time, ConnectorActivationAudit) (printing.Connector, error)
	ReportPrintPrinter(context.Context, printing.ConsumerAuthority, printing.PrinterReport, time.Time) error
}
