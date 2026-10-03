package memory

import (
	"context"
	"sort"
	"time"

	p "github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *Store) CleanupExpiredPrintPairings(_ context.Context, now time.Time, limit int) (int, error) {
	if limit < 1 || limit > 1000 {
		return 0, ports.ErrPrintConflict
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	expired := []p.Pairing{}
	for _, pairing := range s.printingPairings {
		if !pairing.ExpiresAt.After(now) {
			expired = append(expired, pairing)
		}
	}
	sort.Slice(expired, func(i, j int) bool {
		if expired[i].ExpiresAt.Equal(expired[j].ExpiresAt) {
			return expired[i].ID < expired[j].ID
		}
		return expired[i].ExpiresAt.Before(expired[j].ExpiresAt)
	})
	if len(expired) > limit {
		expired = expired[:limit]
	}
	for _, pairing := range expired {
		delete(s.printingPairings, pairing.ID)
	}
	return len(expired), nil
}
