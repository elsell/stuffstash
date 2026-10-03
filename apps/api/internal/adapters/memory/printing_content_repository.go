package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"slices"
	"time"
)

func (s *Store) GetPrintJobContent(_ context.Context, scope printing.Scope, id printing.JobID, now time.Time) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, err := s.printJobLocked(scope, id)
	if err != nil {
		return nil, err
	}
	content := s.printingJobContents[id]
	if !now.Before(j.Artifact.ExpiresAt) || len(content) == 0 {
		return nil, ports.ErrPrintJobNotFound
	}
	return slices.Clone(content), nil
}
