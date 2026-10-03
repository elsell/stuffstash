package printing

import "time"

type PrinterReadiness string

const (
	PrinterUnknown     PrinterReadiness = "unknown"
	PrinterReady       PrinterReadiness = "ready"
	PrinterUnavailable PrinterReadiness = "unavailable"
	PrinterError       PrinterReadiness = "error"
)

// Printer reservation values are shared with the durable job queue. They are
// deliberately independent of presentation readiness reported by connectors.
const (
	PrinterReservationClaimed   = "claimed"
	PrinterReservationPrinting  = "printing"
	PrinterReservationUncertain = "uncertain"
)

type Printer struct {
	RequestKey, RequestFingerprint string
	ID                             PrinterID
	Scope                          Scope
	Name, AdapterID, DeviceID      string
	Media                          MediaSnapshot
	MediaFingerprint               string
	Revision                       uint64
	Retired                        bool
	ActiveJobID, ReservationState  string
	Readiness                      PrinterReadiness
	ReadinessReason                string
	ReportedAt                     *time.Time
	CreatedAt, UpdatedAt           time.Time
}
type PrinterReport struct {
	Scope       Scope
	PrinterID   PrinterID
	ConnectorID ConnectorID
	State       PrinterReadiness
	Reason      string
	ReportedAt  time.Time
}
