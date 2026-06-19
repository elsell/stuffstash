package memory

import (
	"context"
	"fmt"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sort"
	"time"
)

func (s *Store) SaveInventoryAccessGrantAndEnqueue(_ context.Context, eventID string, grant ports.InventoryAccessGrant, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.inventories[grant.InventoryID]
	if !ok || item.TenantID.String() != grant.TenantID.String() {
		return ports.ErrForbidden
	}

	grantKey := inventoryAccessGrantStorageKey(grant)
	if _, exists := s.accessGrants[grantKey]; exists {
		return nil
	}
	eventKind, ok := grant.Relationship.GrantOutboxKind()
	if !ok {
		return fmt.Errorf("invalid inventory access relationship %q", grant.Relationship)
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.accessGrants[grantKey] = grant
	s.auditRecords[auditRecord.ID] = auditRecord
	s.outbox[eventID] = ports.AuthorizationOutboxEvent{
		ID:          eventID,
		Kind:        eventKind,
		PrincipalID: grant.PrincipalID,
		TenantID:    grant.TenantID,
		InventoryID: grant.InventoryID,
		CreatedAt:   time.Now(),
	}
	return nil
}

func (s *Store) DeleteInventoryAccessGrantAndClaimRevoke(_ context.Context, eventID string, claimID string, leaseUntil time.Time, grant ports.InventoryAccessGrant, auditRecord audit.Record) (ports.AuthorizationOutboxEvent, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.inventories[grant.InventoryID]
	if !ok || item.TenantID.String() != grant.TenantID.String() {
		return ports.AuthorizationOutboxEvent{}, false, ports.ErrForbidden
	}

	grantKey := inventoryAccessGrantStorageKey(grant)
	eventKind, ok := grant.Relationship.RevokeOutboxKind()
	if !ok {
		return ports.AuthorizationOutboxEvent{}, false, fmt.Errorf("invalid inventory access relationship %q", grant.Relationship)
	}
	_, removed := s.accessGrants[grantKey]
	if removed {
		if _, exists := s.auditRecords[auditRecord.ID]; exists {
			return ports.AuthorizationOutboxEvent{}, false, ports.ErrConflict
		}
		delete(s.accessGrants, grantKey)
		s.auditRecords[auditRecord.ID] = auditRecord
	}
	if _, exists := s.outbox[eventID]; exists {
		return ports.AuthorizationOutboxEvent{}, false, ports.ErrConflict
	}
	event := ports.AuthorizationOutboxEvent{
		ID:           eventID,
		Kind:         eventKind,
		PrincipalID:  grant.PrincipalID,
		TenantID:     grant.TenantID,
		InventoryID:  grant.InventoryID,
		ClaimID:      claimID,
		ClaimedUntil: leaseUntil,
		CreatedAt:    time.Now(),
	}
	s.outbox[eventID] = event
	return event, removed, nil
}

func (s *Store) ListInventoryAccessGrants(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.InventoryAccessGrantPageRequest) ([]ports.InventoryAccessGrant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []ports.InventoryAccessGrant{}
	for _, grant := range s.accessGrants {
		key := grant.CursorKey()
		if grant.TenantID == tenantID && grant.InventoryID == inventoryID && key > page.AfterGrantKey {
			items = append(items, grant)
		}
	}
	sort.Slice(items, func(left int, right int) bool {
		return items[left].CursorKey() < items[right].CursorKey()
	})
	if page.Limit > 0 && len(items) > page.Limit {
		items = items[:page.Limit]
	}
	return items, nil
}

func inventoryAccessGrantStorageKey(grant ports.InventoryAccessGrant) string {
	return grant.TenantID.String() + ":" + grant.InventoryID.String() + ":" + grant.CursorKey()
}
