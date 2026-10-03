package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sort"
)

func (s *Store) MaintainPrintJobs(_ context.Context, in ports.PrintJobMaintenance) (ports.PrintMaintenancePage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	page := ports.PrintMaintenancePage{}
	if in.Limit < 1 || in.Limit > 1000 || in.Audit == nil {
		return page, ports.ErrConflict
	}
	ids := []string{}
	for id := range s.printingJobs {
		if string(id) > in.After {
			ids = append(ids, string(id))
		}
	}
	sort.Strings(ids)
	page.HasMore = len(ids) > in.Limit
	if page.HasMore {
		ids = ids[:in.Limit]
	}
	for _, id := range ids {
		before := s.printingJobs[printing.JobID(id)]
		j := clonePrintJob(before)
		j.Maintain(in.Now)
		if j.Revision != before.Revision {
			record, err := in.Audit(before, j)
			if err != nil {
				return page, err
			}
			if err = s.validatePrintAuditLocked(record, j); err != nil {
				return page, err
			}
			p := s.printingPrinters[j.PrinterID]
			printReservation(&p, j)
			s.printingPrinters[p.ID] = p
			s.printingJobs[j.ID] = j
			s.auditRecords[record.ID] = record
		}
		if j.Terminal() && !j.UpdatedAt.After(in.TerminalBefore) {
			delete(s.printingJobs, j.ID)
			delete(s.printingJobFingerprints, j.ID)
			delete(s.printingJobContents, j.ID)
		} else if j.Terminal() && !j.Artifact.ExpiresAt.After(in.Now) {
			delete(s.printingJobContents, j.ID)
		}
		page.After = id
	}
	return page, nil
}
