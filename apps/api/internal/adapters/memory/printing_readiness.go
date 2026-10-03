package memory

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

func (s *Store) printingDispatchReadyLocked(authority printing.ConsumerAuthority, now time.Time, maxAge time.Duration) bool {
	connector := s.printingConnectors[authority.ConnectorID]
	if connector.LastSeenAt == nil || now.Before(*connector.LastSeenAt) || now.Sub(*connector.LastSeenAt) > maxAge {
		return false
	}
	report, found := s.printingReports[string(authority.ConnectorID)+":"+string(authority.PrinterID)]
	return found && report.Scope == authority.Scope && report.State == printing.PrinterReady && !now.Before(report.ReportedAt) && now.Sub(report.ReportedAt) <= maxAge
}
