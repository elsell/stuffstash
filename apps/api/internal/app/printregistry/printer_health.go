package printregistry

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s Service) withHealth(ctx context.Context, p printing.Printer) (printing.Printer, error) {
	p.Readiness = printing.PrinterUnknown
	p.ReadinessReason = ""
	p.ReportedAt = nil
	if s.Health == nil || s.PrintingAuthorization == nil || s.ReportMaxAge <= 0 || p.Retired {
		return p, nil
	}
	reports, err := s.Health.ListPrintPrinterHealth(ctx, p.Scope, p.ID)
	if err != nil {
		return p, err
	}
	now := s.Clock.Now()
	for _, health := range reports {
		c, b, r := health.Connector, health.Binding, health.Report
		authority := printing.ConsumerAuthority{Scope: p.Scope, ConnectorID: c.ID, ServiceAccountID: c.ServiceAccountID, CredentialVersion: c.CredentialVersion, PrinterID: p.ID, BindingGeneration: b.Generation}
		if !printing.AcceptsAuthority(c, b, authority, now) || c.LastSeenAt == nil || now.Sub(*c.LastSeenAt) > s.ReportMaxAge || now.Sub(r.ReportedAt) > s.ReportMaxAge {
			continue
		}
		if s.PrintingAuthorization.CheckPrintConnector(ctx, c.ServiceAccountID, c.ID) != nil || s.PrintingAuthorization.CheckPrinter(ctx, c.ServiceAccountID, p.ID, ports.PrinterPermissionConsume) != nil {
			continue
		}
		if p.Readiness == printing.PrinterReady && r.State != printing.PrinterReady {
			continue
		}
		if p.ReportedAt != nil && r.State != printing.PrinterReady && r.ReportedAt.Before(*p.ReportedAt) {
			continue
		}
		timestamp := r.ReportedAt
		p.Readiness = r.State
		p.ReadinessReason = r.Reason
		p.ReportedAt = &timestamp
	}
	return p, nil
}
