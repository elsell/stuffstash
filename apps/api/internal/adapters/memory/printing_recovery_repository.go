package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sort"
)

func (s *Store) ListPrintConsumerAttempts(_ context.Context, scope printing.Scope, connector printing.ConnectorID, printer printing.PrinterID, limit int, after string) ([]printing.Job, error) {
	if scope.TenantID == "" || scope.InventoryID == "" || connector == "" || limit < 1 || limit > 101 {
		return nil, ports.ErrConflict
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []printing.Job{}
	for _, j := range s.printingJobs {
		if j.Scope != scope || string(j.ID) <= after || !j.HoldsReservation() || (printer != "" && j.PrinterID != printer) || len(j.Attempts) == 0 || j.Attempts[len(j.Attempts)-1].Authority.ConnectorID != connector {
			continue
		}
		result = append(result, clonePrintJob(j))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}
