package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *Store) RenewPrintJob(_ context.Context, in ports.PrintLeaseRenewal) (printing.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if in.Owner.ConnectorID != in.Authority.ConnectorID {
		return printing.Job{}, ports.ErrPrintDenied
	}
	if err := s.printingConsumerFenceLocked(in.Authority, in.Now); err != nil {
		return printing.Job{}, err
	}
	p, ok := s.printingPrinters[in.Authority.PrinterID]
	if !ok || p.Scope != in.Authority.Scope {
		return printing.Job{}, ports.ErrPrintNotFound
	}
	j, err := s.printJobLocked(in.Authority.Scope, in.JobID)
	if err != nil {
		return printing.Job{}, err
	}
	if j.PrinterID != p.ID || p.ActiveJobID != string(j.ID) {
		return printing.Job{}, ports.ErrConflict
	}
	if err = j.Renew(in.Owner, in.Now, in.Lease, in.Revision); err != nil {
		return printing.Job{}, err
	}
	s.printingJobs[j.ID] = clonePrintJob(j)
	return clonePrintJob(j), nil
}
