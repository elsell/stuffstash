package memory

import (
	"context"
	"slices"
	"sort"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

var _ ports.PrintJobRepository = (*Store)(nil)

func clonePrintJob(j printing.Job) printing.Job {
	j.Attempts = slices.Clone(j.Attempts)
	if j.Resolution != nil {
		resolution := *j.Resolution
		j.Resolution = &resolution
	}
	return j
}
func (s *Store) printJobLocked(scope printing.Scope, id printing.JobID) (printing.Job, error) {
	j, ok := s.printingJobs[id]
	if !ok || j.Scope != scope || scope.TenantID == "" || scope.InventoryID == "" {
		return printing.Job{}, ports.ErrPrintJobNotFound
	}
	return clonePrintJob(j), nil
}
func (s *Store) GetPrintJob(_ context.Context, scope printing.Scope, id printing.JobID) (printing.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.printJobLocked(scope, id)
}
func (s *Store) ListPrintJobs(_ context.Context, scope printing.Scope, printer printing.PrinterID, limit int, after string) ([]printing.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if scope.TenantID == "" || scope.InventoryID == "" || limit < 1 || limit > 101 {
		return nil, ports.ErrConflict
	}
	result := []printing.Job{}
	for _, j := range s.printingJobs {
		if j.Scope == scope && (printer == "" || j.PrinterID == printer) && string(j.ID) > after {
			result = append(result, clonePrintJob(j))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}
func (s *Store) FindPrintAttempt(_ context.Context, scope printing.Scope, connector printing.ConnectorID, id printing.AttemptID) (printing.Job, error) {
	if scope.TenantID == "" || scope.InventoryID == "" || connector == "" || id == "" {
		return printing.Job{}, ports.ErrPrintJobNotFound
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, j := range s.printingJobs {
		if j.Scope != scope {
			continue
		}
		for _, a := range j.Attempts {
			if a.ID == id && a.Authority.ConnectorID == connector {
				return clonePrintJob(j), nil
			}
		}
	}
	return printing.Job{}, ports.ErrPrintJobNotFound
}
func (s *Store) CreatePrintJob(_ context.Context, input ports.PrintJobCreate) (printing.Job, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j := input.Job
	if j.Scope.TenantID == "" || j.Scope.InventoryID == "" || j.RequestedBy == "" || j.IdempotencyKey == "" || input.RequestFingerprint == "" {
		return printing.Job{}, false, ports.ErrConflict
	}
	for _, existing := range s.printingJobs {
		if existing.Scope == j.Scope && existing.RequestedBy == j.RequestedBy && existing.IdempotencyKey == j.IdempotencyKey {
			if s.printingJobFingerprints[existing.ID] != input.RequestFingerprint {
				return printing.Job{}, false, ports.ErrConflict
			}
			return clonePrintJob(existing), false, nil
		}
	}
	p, ok := s.printingPrinters[j.PrinterID]
	if !ok || p.Scope != j.Scope || p.Retired || p.Revision != input.PrinterRevision || p.MediaFingerprint != j.MediaFingerprint {
		return printing.Job{}, false, ports.ErrConflict
	}
	if j.ID == "" || j.IdempotencyKey == "" || j.Status != printing.JobQueued || j.Revision != 1 || len(j.Attempts) != 0 || input.RequestFingerprint == "" {
		return printing.Job{}, false, ports.ErrConflict
	}
	if _, exists := s.printingJobs[j.ID]; exists {
		return printing.Job{}, false, ports.ErrConflict
	}
	if j.Predecessor != "" {
		previous, err := s.printJobLocked(j.Scope, j.Predecessor)
		if err != nil {
			return printing.Job{}, false, err
		}
		if !previous.Terminal() || previous.Kind != j.Kind || previous.AssetID != j.AssetID || previous.IdempotencyKey == j.IdempotencyKey {
			return printing.Job{}, false, ports.ErrConflict
		}
	}
	if err := s.validatePrintAuditLocked(input.Audit, j); err != nil {
		return printing.Job{}, false, err
	}
	if s.printingJobs == nil {
		s.printingJobs = map[printing.JobID]printing.Job{}
		s.printingJobFingerprints = map[printing.JobID]string{}
	}
	if s.printingJobContents == nil {
		s.printingJobContents = map[printing.JobID][]byte{}
	}
	s.printingJobContents[j.ID] = slices.Clone(input.Content)
	s.printingJobs[j.ID] = clonePrintJob(j)
	s.printingJobFingerprints[j.ID] = input.RequestFingerprint
	s.auditRecords[input.Audit.ID] = input.Audit
	return clonePrintJob(j), true, nil
}
func (s *Store) validatePrintAuditLocked(record audit.Record, j printing.Job) error {
	if record.ID == "" || string(record.TenantID) != j.Scope.TenantID || string(record.InventoryID) != j.Scope.InventoryID || record.TargetID != string(j.ID) {
		return ports.ErrConflict
	}
	if _, exists := s.auditRecords[record.ID]; exists {
		return ports.ErrConflict
	}
	return nil
}
func printReservation(p *printing.Printer, j printing.Job) {
	if j.HoldsReservation() {
		p.ActiveJobID = string(j.ID)
		p.ReservationState = string(j.Status)
	} else if p.ActiveJobID == string(j.ID) {
		p.ActiveJobID = ""
		p.ReservationState = ""
	}
}
func (s *Store) UpdatePrintJob(_ context.Context, input ports.PrintJobUpdate) (printing.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if input.Authority != nil {
		if input.Authority.Scope != input.Scope || input.Authority.PrinterID != input.PrinterID {
			return printing.Job{}, ports.ErrPrintDenied
		}
		if err := s.printingConsumerFenceLocked(*input.Authority, input.Now); err != nil {
			return printing.Job{}, err
		}
	}
	p, ok := s.printingPrinters[input.PrinterID]
	if !ok || p.Scope != input.Scope {
		return printing.Job{}, ports.ErrPrintJobNotFound
	}
	before, err := s.printJobLocked(input.Scope, input.JobID)
	if err != nil {
		return printing.Job{}, err
	}
	if before.PrinterID != p.ID || input.Change == nil || input.Audit == nil {
		return printing.Job{}, ports.ErrConflict
	}
	j := clonePrintJob(before)
	if err := input.Change(&j, p); err != nil {
		return printing.Job{}, err
	}
	if j.Revision == before.Revision {
		return before, nil
	}
	if j.ID != before.ID || j.Scope != before.Scope || j.PrinterID != before.PrinterID {
		return printing.Job{}, ports.ErrConflict
	}
	if before.Status != printing.JobPrinting && j.Status == printing.JobPrinting && input.StartReportMaxAge > 0 {
		if input.Authority == nil || !s.printingDispatchReadyLocked(*input.Authority, input.Now, input.StartReportMaxAge) {
			return printing.Job{}, ports.ErrConflict
		}
	}
	if before.Status != printing.JobPrinting && j.Status == printing.JobPrinting && (p.Retired || p.MediaFingerprint != j.MediaFingerprint) {
		return printing.Job{}, ports.ErrConflict
	}
	if j.HoldsReservation() && p.ActiveJobID != "" && p.ActiveJobID != string(j.ID) {
		return printing.Job{}, ports.ErrConflict
	}
	record, err := input.Audit(before, j)
	if err != nil {
		return printing.Job{}, err
	}
	if err = s.validatePrintAuditLocked(record, j); err != nil {
		return printing.Job{}, err
	}
	printReservation(&p, j)
	s.printingPrinters[p.ID] = p
	s.printingJobs[j.ID] = clonePrintJob(j)
	s.auditRecords[record.ID] = record
	return clonePrintJob(j), nil
}
