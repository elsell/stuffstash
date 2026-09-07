package memory

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sort"
	"time"
)

func (s *Store) ClaimBlobDeletionRechecks(ctx context.Context, claimID string, limit int, now, leaseUntil time.Time, interval time.Duration) ([]ports.BlobDeletionEvent, error) {
	now = now.UTC().Truncate(time.Microsecond)
	leaseUntil = leaseUntil.UTC().Truncate(time.Microsecond)
	if claimID == "" || limit < 1 || limit > 1000 || now.IsZero() || !leaseUntil.After(now) || interval <= 0 {
		return nil, errors.New("invalid deletion recheck claim")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	candidates := []ports.BlobDeletionEvent{}
	for _, event := range s.blobDeletions {
		if event.ProcessedAt.IsZero() || event.ProcessedAt.After(now.Add(-interval)) || event.RecheckedAt.After(now.Add(-interval)) || (event.ClaimID != "" && event.ClaimedUntil.After(now)) {
			continue
		}
		candidates = append(candidates, event)
	}
	sort.Slice(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if !a.RecheckedAt.Equal(b.RecheckedAt) {
			return a.RecheckedAt.Before(b.RecheckedAt)
		}
		if a.RecheckedAt.IsZero() && !a.ProcessedAt.Equal(b.ProcessedAt) {
			return a.ProcessedAt.Before(b.ProcessedAt)
		}
		return a.ID < b.ID
	})
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	for i := range candidates {
		candidates[i].ClaimID = claimID
		candidates[i].ClaimedUntil = leaseUntil
		s.blobDeletions[candidates[i].ID] = candidates[i]
	}
	return candidates, nil
}
func (s *Store) ResolveBlobDeletionRecheck(ctx context.Context, event ports.BlobDeletionEvent, now time.Time, failed bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	now = now.UTC().Truncate(time.Microsecond)
	s.mu.Lock()
	defer s.mu.Unlock()
	current, exists := s.blobDeletions[event.ID]
	if !exists || event.StorageKey == "" || event.ClaimID == "" || current.ClaimID != event.ClaimID || current.StorageKey != event.StorageKey || !current.ClaimedUntil.Equal(event.ClaimedUntil) || !current.ClaimedUntil.After(now) || current.ProcessedAt.IsZero() || now.IsZero() {
		return ports.ErrOutboxClaimLost
	}
	current.RecheckedAt = now
	current.RecheckFailed = failed
	current.ClaimID = ""
	current.ClaimedUntil = time.Time{}
	s.blobDeletions[event.ID] = current
	return nil
}
