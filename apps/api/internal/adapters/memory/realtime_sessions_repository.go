package memory

import (
	"context"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *Store) SaveRealtimeSession(_ context.Context, record ports.RealtimeSessionRecord) error {
	if err := validateRealtimeSessionRecord(record); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.realtimeSessions[record.ID] = record
	return nil
}

func (s *Store) UpdateRealtimeSessionOutcome(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, sessionID string, outcome ports.RealtimeSessionOutcome) error {
	if tenantID.String() == "" || inventoryID.String() == "" || strings.TrimSpace(sessionID) == "" || validateRealtimeSessionOutcome(outcome) != nil {
		return ports.ErrInvalidProviderInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	record, found := s.realtimeSessions[sessionID]
	if !found {
		return ports.ErrInvalidProviderInput
	}
	if record.TenantID != tenantID || record.InventoryID != inventoryID || record.State != ports.RealtimeSessionStateStarted || outcome.At.Before(record.StartedAt) {
		return ports.ErrInvalidProviderInput
	}
	record.State = outcome.State
	record.LastActivityAt = outcome.At
	record.EndedAt = outcome.At
	record.SafeFailureCode = strings.TrimSpace(outcome.SafeFailureCode)
	s.realtimeSessions[sessionID] = record
	return nil
}

func (s *Store) RealtimeSessionByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, sessionID string) (ports.RealtimeSessionRecord, bool, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || strings.TrimSpace(sessionID) == "" {
		return ports.RealtimeSessionRecord{}, false, ports.ErrInvalidProviderInput
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, found := s.realtimeSessions[sessionID]
	if !found || record.TenantID != tenantID || record.InventoryID != inventoryID {
		return ports.RealtimeSessionRecord{}, false, nil
	}
	return record, true, nil
}

func validateRealtimeSessionRecord(record ports.RealtimeSessionRecord) error {
	if strings.TrimSpace(record.ID) == "" ||
		record.TenantID.String() == "" ||
		record.InventoryID.String() == "" ||
		record.PrincipalID.String() == "" ||
		strings.TrimSpace(record.Source) == "" ||
		record.State != ports.RealtimeSessionStateStarted ||
		record.StartedAt.IsZero() ||
		record.LastActivityAt.IsZero() {
		return ports.ErrInvalidProviderInput
	}
	if !record.EndedAt.IsZero() || strings.TrimSpace(record.SafeFailureCode) != "" {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func validRealtimeSessionFinalState(state ports.RealtimeSessionState) bool {
	switch state {
	case ports.RealtimeSessionStateCompleted, ports.RealtimeSessionStateFailed, ports.RealtimeSessionStateCancelled:
		return true
	default:
		return false
	}
}

func validateRealtimeSessionOutcome(outcome ports.RealtimeSessionOutcome) error {
	if !validRealtimeSessionFinalState(outcome.State) || outcome.At.IsZero() {
		return ports.ErrInvalidProviderInput
	}
	if outcome.State == ports.RealtimeSessionStateFailed && strings.TrimSpace(outcome.SafeFailureCode) == "" {
		return ports.ErrInvalidProviderInput
	}
	if outcome.State != ports.RealtimeSessionStateFailed && strings.TrimSpace(outcome.SafeFailureCode) != "" {
		return ports.ErrInvalidProviderInput
	}
	return nil
}
