package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *Store) FindPrintJobRequest(_ context.Context, scope printing.Scope, actor, key string) (printing.Job, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if scope.TenantID == "" || scope.InventoryID == "" || actor == "" || key == "" {
		return printing.Job{}, "", ports.ErrPrintJobNotFound
	}
	for _, j := range s.printingJobs {
		if j.Scope == scope && j.RequestedBy == actor && j.IdempotencyKey == key {
			return clonePrintJob(j), s.printingJobFingerprints[j.ID], nil
		}
	}
	return printing.Job{}, "", ports.ErrPrintJobNotFound
}
