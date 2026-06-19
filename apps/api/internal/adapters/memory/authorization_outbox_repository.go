package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sort"
	"time"
)

func (s *Store) ClaimPendingAuthorizationOutboxEvents(_ context.Context, claimID string, limit int, leaseUntil time.Time) ([]ports.AuthorizationOutboxEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if limit <= 0 {
		limit = len(s.outbox)
	}
	now := time.Now()
	events := []ports.AuthorizationOutboxEvent{}
	for _, event := range s.outbox {
		if !event.DeadLetteredAt.IsZero() {
			continue
		}
		if !event.ClaimedUntil.IsZero() && event.ClaimedUntil.After(now) {
			continue
		}
		events = append(events, event)
	}
	sort.Slice(events, func(left int, right int) bool {
		if events[left].CreatedAt.Equal(events[right].CreatedAt) {
			return events[left].ID < events[right].ID
		}
		return events[left].CreatedAt.Before(events[right].CreatedAt)
	})
	if len(events) > limit {
		events = events[:limit]
	}
	for index, event := range events {
		event.ClaimID = claimID
		event.ClaimedUntil = leaseUntil
		s.outbox[event.ID] = event
		events[index] = event
	}
	return events, nil
}

func (s *Store) MarkAuthorizationOutboxEventProcessed(_ context.Context, eventID string, claimID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.outbox[eventID]
	if !ok || event.ClaimID != claimID {
		return ports.ErrAuthorizationOutboxClaimLost
	}
	delete(s.outbox, eventID)
	return nil
}

func (s *Store) MarkAuthorizationOutboxEventFailed(_ context.Context, eventID string, claimID string, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.outbox[eventID]
	if !ok || event.ClaimID != claimID {
		return ports.ErrAuthorizationOutboxClaimLost
	}
	event.Attempts++
	event.LastError = reason
	event.ClaimID = ""
	event.ClaimedUntil = time.Time{}
	s.outbox[eventID] = event
	return nil
}

func (s *Store) MarkAuthorizationOutboxEventDeadLettered(_ context.Context, eventID string, claimID string, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.outbox[eventID]
	if !ok || event.ClaimID != claimID {
		return ports.ErrAuthorizationOutboxClaimLost
	}
	event.DeadLetteredAt = time.Now()
	event.DeadLetterReason = reason
	event.ClaimID = ""
	event.ClaimedUntil = time.Time{}
	s.outbox[eventID] = event
	return nil
}
