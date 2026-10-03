package memory

import (
	"context"
	"sort"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *Store) ClaimPrintJob(_ context.Context, input ports.PrintClaim) (printing.Job, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.printingConsumerFenceLocked(input.Authority, input.Now); err != nil {
		return printing.Job{}, false, err
	}
	if !input.Owner.Valid() || input.Owner.ConnectorID != input.Authority.ConnectorID || input.Lease <= 0 || input.ReportMaxAge <= 0 || input.Audit == nil {
		return printing.Job{}, false, ports.ErrConflict
	}
	p, ok := s.printingPrinters[input.Authority.PrinterID]
	if !ok || p.Scope != input.Authority.Scope || p.Retired {
		return printing.Job{}, false, ports.ErrConflict
	}
	changes := map[printing.JobID]printing.Job{}
	records := map[audit.ID]audit.Record{}
	stage := func(before, after printing.Job) error {
		record, err := input.Audit(before, after)
		if err != nil {
			return err
		}
		if err = s.validatePrintAuditLocked(record, after); err != nil {
			return err
		}
		if _, exists := records[record.ID]; exists {
			return ports.ErrConflict
		}
		changes[after.ID] = clonePrintJob(after)
		records[record.ID] = record
		printReservation(&p, after)
		return nil
	}
	commit := func() {
		for id, j := range changes {
			s.printingJobs[id] = j
		}
		for id, r := range records {
			s.auditRecords[id] = r
		}
		s.printingPrinters[p.ID] = p
	}
	if p.ActiveJobID != "" {
		before, err := s.printJobLocked(p.Scope, printing.JobID(p.ActiveJobID))
		if err != nil {
			return printing.Job{}, false, err
		}
		active := clonePrintJob(before)
		if active.Expire(input.Now) {
			if err := stage(before, active); err != nil {
				return printing.Job{}, false, err
			}
		}
		if p.ActiveJobID != "" {
			commit()
			return printing.Job{}, false, nil
		}
	}
	connector := s.printingConnectors[input.Authority.ConnectorID]
	if connector.LastSeenAt == nil || input.Now.Before(*connector.LastSeenAt) || input.Now.Sub(*connector.LastSeenAt) > input.ReportMaxAge {
		commit()
		return printing.Job{}, false, nil
	}
	report, ready := s.printingReports[string(input.Authority.ConnectorID)+":"+string(p.ID)]
	if !ready || report.Scope != p.Scope || report.State != printing.PrinterReady || input.Now.Before(report.ReportedAt) || input.Now.Sub(report.ReportedAt) > input.ReportMaxAge {
		commit()
		return printing.Job{}, false, nil
	}
	eligible := []printing.Job{}
	for id, stored := range s.printingJobs {
		j := stored
		if changed, ok := changes[id]; ok {
			j = changed
		}
		if j.Scope == p.Scope && j.PrinterID == p.ID && j.Status == printing.JobQueued && j.MediaFingerprint == p.MediaFingerprint {
			eligible = append(eligible, clonePrintJob(j))
		}
	}
	sort.Slice(eligible, func(i, j int) bool {
		if !eligible[i].CreatedAt.Equal(eligible[j].CreatedAt) {
			return eligible[i].CreatedAt.Before(eligible[j].CreatedAt)
		}
		return eligible[i].ID < eligible[j].ID
	})
	if len(eligible) == 0 {
		commit()
		return printing.Job{}, false, nil
	}
	for _, existing := range s.printingJobs {
		for _, attempt := range existing.Attempts {
			if attempt.ID == input.Owner.AttemptID {
				return printing.Job{}, false, ports.ErrConflict
			}
		}
	}
	before := eligible[0]
	j := clonePrintJob(before)
	if err := j.Claim(input.Owner.AttemptID, input.Owner, input.Now, input.Lease, j.Revision); err != nil {
		return printing.Job{}, false, err
	}
	if err := stage(before, j); err != nil {
		return printing.Job{}, false, err
	}
	commit()
	return clonePrintJob(j), true, nil
}
