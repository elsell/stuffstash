package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sort"
	"time"
)

func (s *Store) SaveTenant(_ context.Context, item tenant.Tenant) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if item.LifecycleState.String() == "" {
		item.LifecycleState = tenant.LifecycleStateActive
	}
	s.tenants[item.ID] = item
	return nil
}

func (s *Store) TenantByID(_ context.Context, tenantID tenant.ID) (tenant.Tenant, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.tenants[tenantID]
	return item, ok, nil
}

func (s *Store) TenantExists(_ context.Context, tenantID tenant.ID) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.tenants[tenantID]
	return ok && item.IsActive(), nil
}

func (s *Store) ListTenants(_ context.Context, page ports.TenantListPageRequest) ([]tenant.Tenant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []tenant.Tenant{}
	for _, item := range s.tenants {
		if item.IsActive() && item.ID.String() > page.AfterTenantID.String() {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(left int, right int) bool {
		return items[left].ID.String() < items[right].ID.String()
	})
	if page.Limit > 0 && len(items) > page.Limit {
		items = items[:page.Limit]
	}
	return items, nil
}

func (s *Store) UpdateTenant(_ context.Context, item tenant.Tenant, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.tenants[item.ID]
	if !ok || !existing.IsActive() || item.LifecycleState != tenant.LifecycleStateActive {
		return ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.tenants[item.ID] = item
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) UpdateTenantLifecycle(_ context.Context, item tenant.Tenant, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.tenants[item.ID]
	if !ok || existing.Name != item.Name || existing.LifecycleState == item.LifecycleState {
		return ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.tenants[item.ID] = item
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) DeleteTenant(_ context.Context, tenantID tenant.ID, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tenants[tenantID]; !ok {
		return ports.ErrForbidden
	}
	for _, item := range s.inventories {
		if item.TenantID.String() == tenantID.String() {
			return ports.ErrForbidden
		}
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.auditRecords[auditRecord.ID] = auditRecord
	delete(s.tenants, tenantID)
	return nil
}

func (s *Store) SaveTenantAndEnqueueOwnerGrant(_ context.Context, eventID string, item tenant.Tenant, principal identity.Principal, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	if item.LifecycleState.String() == "" {
		item.LifecycleState = tenant.LifecycleStateActive
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
