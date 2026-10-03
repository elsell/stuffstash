package printing

import "time"

type PrinterReadiness string

const (
	PrinterUnknown     PrinterReadiness = "unknown"
	PrinterReady       PrinterReadiness = "ready"
	PrinterUnavailable PrinterReadiness = "unavailable"
	PrinterError       PrinterReadiness = "error"
)

type Printer struct {
	ID                            PrinterID
	Scope                         Scope
	Name, AdapterID, DeviceID     string
	Media                         MediaSnapshot
	MediaFingerprint              string
	Revision                      uint64
	Retired                       bool
	ActiveJobID, ReservationState string
	Readiness                     PrinterReadiness
	ReadinessReason               string
	ReportedAt                    *time.Time
	CreatedAt, UpdatedAt          time.Time
}
type PrinterReport struct {
	Scope       Scope
	PrinterID   PrinterID
	ConnectorID ConnectorID
	State       PrinterReadiness
	Reason      string
	ReportedAt  time.Time
}
