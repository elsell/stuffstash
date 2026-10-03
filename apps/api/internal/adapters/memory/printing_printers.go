package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sort"
)

func clonePrintingPrinter(p printing.Printer) printing.Printer {
	if p.ReportedAt != nil {
		value := *p.ReportedAt
		p.ReportedAt = &value
	}
	return p
}
func (s *Store) CreatePrinter(_ context.Context, p printing.Printer, record audit.Record) (printing.Printer, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.printingInventoryExistsLocked(p.Scope) {
		return printing.Printer{}, false, ports.ErrPrintNotFound
	}
	for _, existing := range s.printingPrinters {
		if existing.Scope == p.Scope && existing.RequestKey == p.RequestKey {
			if existing.RequestFingerprint != p.RequestFingerprint {
				return printing.Printer{}, false, ports.ErrPrintConflict
			}
			return clonePrintingPrinter(existing), false, nil
		}
	}
	if _, exists := s.printingPrinters[p.ID]; exists {
		return printing.Printer{}, false, ports.ErrPrintConflict
	}
	if _, exists := s.auditRecords[record.ID]; exists {
		return printing.Printer{}, false, ports.ErrPrintConflict
	}
	if s.printingPrinters == nil {
		s.printingPrinters = map[printing.PrinterID]printing.Printer{}
	}
	s.printingPrinters[p.ID] = clonePrintingPrinter(p)
	s.auditRecords[record.ID] = record
	return clonePrintingPrinter(p), true, nil
}
func (s *Store) GetPrinter(_ context.Context, scope printing.Scope, id printing.PrinterID) (printing.Printer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.printingPrinters[id]
	if !ok || p.Scope != scope {
		return printing.Printer{}, ports.ErrPrintNotFound
	}
	return clonePrintingPrinter(p), nil
}
func (s *Store) ListPrinters(_ context.Context, scope printing.Scope, limit int, after string) ([]printing.Printer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []printing.Printer{}
	for id, p := range s.printingPrinters {
		if p.Scope == scope && string(id) > after {
			out = append(out, clonePrintingPrinter(p))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
func (s *Store) UpdatePrinter(_ context.Context, scope printing.Scope, id printing.PrinterID, revision uint64, change ports.PrinterMutation, makeAudit ports.PrinterAudit, retirement *ports.PrinterRetirement) (printing.Printer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.printingInventoryExistsLocked(scope) {
		return printing.Printer{}, ports.ErrPrintNotFound
	}
	p, ok := s.printingPrinters[id]
	if !ok || p.Scope != scope {
		return printing.Printer{}, ports.ErrPrintNotFound
	}
	if p.Revision != revision {
		return printing.Printer{}, ports.ErrPrintConflict
	}
	next := clonePrintingPrinter(p)
	if err := change(&next); err != nil {
		return printing.Printer{}, err
	}
	if next.ID != p.ID || next.Scope != p.Scope {
		return printing.Printer{}, ports.ErrPrintConflict
	}
	record, err := makeAudit(next)
	if err != nil {
		return printing.Printer{}, err
	}
	if _, exists := s.auditRecords[record.ID]; exists {
		return printing.Printer{}, ports.ErrPrintConflict
	}
	if next.Retired && !p.Retired {
		if err := s.retirePrinterWorkLocked(&next, retirement, record); err != nil {
			return printing.Printer{}, err
		}
	}
	s.printingPrinters[id] = clonePrintingPrinter(next)
	s.auditRecords[record.ID] = record
	return next, nil
}

var _ ports.PrinterRepository = (*Store)(nil)
