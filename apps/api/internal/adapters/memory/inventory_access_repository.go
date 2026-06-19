package memory

import (
	"context"
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
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.accessGrants[grantKey] = grant
	s.auditRecords[auditRecord.ID] = auditRecord
	s.outbox[eventID] = ports.AuthorizationOutboxEvent{
		ID:          eventID,
		Kind:        outboxKindForInventoryAccess(grant.Relationship),
		PrincipalID: grant.PrincipalID,
		TenantID:    grant.TenantID,
		InventoryID: grant.InventoryID,
		CreatedAt:   time.Now(),
	}
	return nil
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

func outboxKindForInventoryAccess(relationship ports.InventoryAccessRelationship) ports.AuthorizationOutboxEventKind {
	switch relationship {
	case ports.InventoryAccessEditor:
		return ports.AuthorizationOutboxGrantInventoryEditor
	default:
		return ports.AuthorizationOutboxGrantInventoryViewer
	}
}
