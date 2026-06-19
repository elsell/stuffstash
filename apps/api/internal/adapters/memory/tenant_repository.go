package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

func (s *Store) SaveTenant(_ context.Context, item tenant.Tenant) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tenants[item.ID] = item
	return nil
}

func (s *Store) TenantExists(_ context.Context, tenantID tenant.ID) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.tenants[tenantID]
	return ok, nil
}

func (s *Store) SaveTenantAndEnqueueOwnerGrant(_ context.Context, eventID string, item tenant.Tenant, principal identity.Principal, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.tenants[item.ID] = item
	s.auditRecords[auditRecord.ID] = auditRecord
	s.outbox[eventID] = ports.AuthorizationOutboxEvent{
		ID:          eventID,
		Kind:        ports.AuthorizationOutboxGrantTenantOwner,
		PrincipalID: principal.ID,
		TenantID:    item.ID,
		CreatedAt:   time.Now(),
	}
	return nil
}
