package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sort"
	"time"
)

func (s *Store) SaveInventory(_ context.Context, item inventory.Inventory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if item.LifecycleState.String() == "" {
		item.LifecycleState = inventory.LifecycleStateActive
	}
	s.inventories[item.ID] = item
	return nil
}

func (s *Store) InventoryByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (inventory.Inventory, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.inventories[inventoryID]
	if !ok || item.TenantID != inventory.TenantID(tenantID.String()) {
		return inventory.Inventory{}, false, nil
	}
	return item, true, nil
}

func (s *Store) UpdateInventory(_ context.Context, item inventory.Inventory, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.inventories[item.ID]
	if !ok || existing.TenantID != item.TenantID || !existing.IsActive() || item.LifecycleState != inventory.LifecycleStateActive {
		return ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.inventories[item.ID] = item
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) UpdateInventoryLifecycle(_ context.Context, item inventory.Inventory, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.inventories[item.ID]
	if !ok || existing.TenantID != item.TenantID || existing.Name != item.Name || existing.LifecycleState == item.LifecycleState {
		return ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.inventories[item.ID] = item
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) DeleteInventory(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.inventories[inventoryID]
	if !ok || item.TenantID.String() != tenantID.String() {
		return ports.ErrForbidden
	}
	for _, item := range s.assets {
		if item.TenantID.String() == tenantID.String() && item.InventoryID.String() == inventoryID.String() {
			return ports.ErrForbidden
		}
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.auditRecords[auditRecord.ID] = auditRecord
	delete(s.inventories, inventoryID)
	return nil
}

func (s *Store) InventoryHasActiveAssets(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, item := range s.assets {
		if item.TenantID.String() == tenantID.String() && item.InventoryID.String() == inventoryID.String() && item.LifecycleState == asset.LifecycleStateActive {
			return true, nil
		}
	}
	return false, nil
}

func (s *Store) SaveInventoryAndEnqueueOwnerGrant(_ context.Context, eventID string, item inventory.Inventory, tenantID tenant.ID, principal identity.Principal, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	if item.LifecycleState.String() == "" {
		item.LifecycleState = inventory.LifecycleStateActive
	}
	s.inventories[item.ID] = item
	s.auditRecords[auditRecord.ID] = auditRecord
	s.outbox[eventID] = ports.AuthorizationOutboxEvent{
		ID:          eventID,
		Kind:        ports.AuthorizationOutboxGrantInventoryOwner,
		PrincipalID: principal.ID,
		TenantID:    tenantID,
		InventoryID: item.ID,
		CreatedAt:   time.Now(),
	}
	return nil
}

func (s *Store) ListInventoriesByTenant(_ context.Context, tenantID inventory.TenantID, page ports.InventoryListPageRequest) ([]inventory.Inventory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []inventory.Inventory{}
	for _, item := range s.inventories {
		if item.TenantID == tenantID && item.IsActive() && item.ID.String() > page.AfterInventoryID.String() {
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
