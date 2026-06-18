package memory

import (
	"context"
	"sync"

	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type Store struct {
	mu          sync.RWMutex
	tenants     map[tenant.ID]tenant.Tenant
	inventories map[inventory.InventoryID]inventory.Inventory
}

func NewStore() *Store {
	return &Store{
		tenants:     map[tenant.ID]tenant.Tenant{},
		inventories: map[inventory.InventoryID]inventory.Inventory{},
	}
}

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

func (s *Store) SaveInventory(_ context.Context, item inventory.Inventory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.inventories[item.ID] = item
	return nil
}

func (s *Store) ListInventoriesByTenant(_ context.Context, tenantID inventory.TenantID) ([]inventory.Inventory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []inventory.Inventory{}
	for _, item := range s.inventories {
		if item.TenantID == tenantID {
			items = append(items, item)
		}
	}
	return items, nil
}
