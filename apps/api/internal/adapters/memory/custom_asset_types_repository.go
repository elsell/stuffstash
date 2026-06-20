package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sort"
)

func (s *Store) SaveCustomAssetType(_ context.Context, assetType customfield.AssetType, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.customAssetTypeParentIsValid(assetType); err != nil {
		return err
	}
	for _, existing := range s.customAssetTypes {
		if customfield.AssetTypesConflict(existing, assetType) {
			return ports.ErrConflict
		}
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.customAssetTypes[assetType.ID] = assetType
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) UpdateCustomAssetType(_ context.Context, assetType customfield.AssetType, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.customAssetTypes[assetType.ID]
	if !exists || existing.TenantID != assetType.TenantID || existing.InventoryID != assetType.InventoryID || existing.Scope != assetType.Scope {
		return ports.ErrForbidden
	}
	if existing.Key != assetType.Key || !existing.IsActive() {
		return ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.customAssetTypes[assetType.ID] = assetType
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) ArchiveCustomAssetType(_ context.Context, assetType customfield.AssetType, auditRecord audit.Record) error {
	if assetType.LifecycleState != customfield.AssetTypeLifecycleArchived {
		return ports.ErrForbidden
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.customAssetTypes[assetType.ID]
	if !exists || existing.TenantID != assetType.TenantID || existing.InventoryID != assetType.InventoryID || existing.Scope != assetType.Scope {
		return ports.ErrForbidden
	}
	if existing.Key != assetType.Key || !existing.IsActive() {
		return ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.customAssetTypes[assetType.ID] = assetType
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) CustomAssetTypeByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetTypeID customfield.AssetTypeID) (customfield.AssetType, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	assetType, ok := s.customAssetTypes[assetTypeID]
	if !ok || assetType.TenantID.String() != tenantID.String() || !assetType.IsActive() {
		return customfield.AssetType{}, false, nil
	}
	if inventoryID.String() == "" {
		if assetType.Scope != customfield.ScopeTenant {
			return customfield.AssetType{}, false, nil
		}
		return assetType, true, nil
	}
	if assetType.Scope == customfield.ScopeInventory && assetType.InventoryID.String() != inventoryID.String() {
		return customfield.AssetType{}, false, nil
	}
	return assetType, true, nil
}

func (s *Store) ListTenantCustomAssetTypes(_ context.Context, tenantID tenant.ID, page ports.CustomAssetTypePageRequest) ([]customfield.AssetType, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []customfield.AssetType{}
	for _, assetType := range s.customAssetTypes {
		if assetType.TenantID.String() == tenantID.String() && assetType.Scope == customfield.ScopeTenant && assetType.IsActive() && assetType.CursorKey() > page.AfterAssetTypeKey {
			items = append(items, assetType)
		}
	}
	return pagedCustomAssetTypes(items, page.Limit), nil
}

func (s *Store) ListInventoryCustomAssetTypes(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.CustomAssetTypePageRequest) ([]customfield.AssetType, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []customfield.AssetType{}
	for _, assetType := range s.customAssetTypes {
		if assetType.TenantID.String() != tenantID.String() || !assetType.IsActive() || assetType.CursorKey() <= page.AfterAssetTypeKey {
			continue
		}
		if assetType.Scope == customfield.ScopeTenant || assetType.InventoryID.String() == inventoryID.String() {
			items = append(items, assetType)
		}
	}
	return pagedCustomAssetTypes(items, page.Limit), nil
}

func (s *Store) CustomAssetTypesByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, ids []customfield.AssetTypeID) ([]customfield.AssetType, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []customfield.AssetType{}
	for _, id := range ids {
		assetType, ok := s.customAssetTypes[id]
		if !ok || assetType.TenantID.String() != tenantID.String() || !assetType.IsActive() {
			continue
		}
		if assetType.Scope == customfield.ScopeInventory && assetType.InventoryID.String() != inventoryID.String() {
			continue
		}
		items = append(items, assetType)
	}
	return items, nil
}

func pagedCustomAssetTypes(items []customfield.AssetType, limit int) []customfield.AssetType {
	sort.Slice(items, func(left int, right int) bool {
		return items[left].CursorKey() < items[right].CursorKey()
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func (s *Store) customAssetTypeParentIsValid(assetType customfield.AssetType) error {
	if _, exists := s.tenants[tenant.ID(assetType.TenantID.String())]; !exists {
		return ports.ErrForbidden
	}
	if assetType.Scope == customfield.ScopeInventory {
		item, ok := s.inventories[inventory.InventoryID(assetType.InventoryID.String())]
		if !ok || item.TenantID.String() != assetType.TenantID.String() {
			return ports.ErrForbidden
		}
	}
	return nil
}
