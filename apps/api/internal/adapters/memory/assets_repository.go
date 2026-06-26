package memory

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sort"
)

func (s *Store) CreateAsset(_ context.Context, item asset.Asset, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.createAssetLocked(item, auditRecord, undoableOperation)
}

func (s *Store) createAssetLocked(item asset.Asset, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	containingInventory, ok := s.inventories[inventory.InventoryID(item.InventoryID.String())]
	if !ok || containingInventory.TenantID.String() != item.TenantID.String() {
		return ports.ErrForbidden
	}
	if _, exists := s.assets[item.ID]; exists {
		return errors.New("asset already exists")
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	if undoableOperation != nil {
		if _, exists := s.undoables[undoableOperation.ID]; exists {
			return ports.ErrConflict
		}
	}
	if item.CustomAssetTypeID.String() != "" {
		assetType, ok := s.customAssetTypes[customfield.AssetTypeID(item.CustomAssetTypeID.String())]
		if !ok || !assetType.IsActive() || assetType.TenantID.String() != item.TenantID.String() || (assetType.Scope == customfield.ScopeInventory && assetType.InventoryID.String() != item.InventoryID.String()) {
			return ports.ErrForbidden
		}
	}

	if item.ParentAssetID.String() != "" {
		parent, ok := s.assets[item.ParentAssetID]
		if !ok {
			return ports.ErrForbidden
		}
		if parent.TenantID != item.TenantID || parent.InventoryID != item.InventoryID || !parent.Kind.CanContainChildren() || parent.LifecycleState != asset.LifecycleStateActive {
			return ports.ErrForbidden
		}
		if parent.ID == item.ID {
			return ports.ErrForbidden
		}
		for current := parent; current.ParentAssetID.String() != ""; {
			next, ok := s.assets[current.ParentAssetID]
			if !ok || next.TenantID != item.TenantID || next.InventoryID != item.InventoryID {
				return ports.ErrForbidden
			}
			if next.ID == item.ID {
				return ports.ErrForbidden
			}
			current = next
		}
	}
	s.assets[item.ID] = item
	s.auditRecords[auditRecord.ID] = auditRecord
	if undoableOperation != nil {
		s.undoables[undoableOperation.ID] = *undoableOperation
	}
	return nil
}

func (s *Store) UpdateAsset(_ context.Context, item asset.Asset, auditRecords []audit.Record, undoableOperation *ports.UndoableOperation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.assets[item.ID]
	if !exists || existing.TenantID != item.TenantID || existing.InventoryID != item.InventoryID {
		return ports.ErrForbidden
	}
	if existing.Kind != item.Kind || existing.LifecycleState != item.LifecycleState {
		return ports.ErrForbidden
	}
	if existing.CustomAssetTypeID != item.CustomAssetTypeID {
		return ports.ErrForbidden
	}
	if item.ParentAssetID.String() != "" {
		parent, ok := s.assets[item.ParentAssetID]
		if !ok {
			return ports.ErrForbidden
		}
		if parent.TenantID != item.TenantID || parent.InventoryID != item.InventoryID || !parent.Kind.CanContainChildren() || parent.LifecycleState != asset.LifecycleStateActive {
			return ports.ErrForbidden
		}
		if parent.ID == item.ID {
			return ports.ErrForbidden
		}
		for current := parent; current.ParentAssetID.String() != ""; {
			next, ok := s.assets[current.ParentAssetID]
			if !ok || next.TenantID != item.TenantID || next.InventoryID != item.InventoryID {
				return ports.ErrForbidden
			}
			if next.ID == item.ID {
				return ports.ErrForbidden
			}
			current = next
		}
	}
	seenAuditRecords := map[audit.ID]struct{}{}
	for _, auditRecord := range auditRecords {
		if _, exists := s.auditRecords[auditRecord.ID]; exists {
			return ports.ErrConflict
		}
		if _, exists := seenAuditRecords[auditRecord.ID]; exists {
			return ports.ErrConflict
		}
		seenAuditRecords[auditRecord.ID] = struct{}{}
	}
	if undoableOperation != nil {
		if _, exists := s.undoables[undoableOperation.ID]; exists {
			return ports.ErrConflict
		}
	}
	s.assets[item.ID] = item
	for _, auditRecord := range auditRecords {
		s.auditRecords[auditRecord.ID] = auditRecord
	}
	if undoableOperation != nil {
		s.undoables[undoableOperation.ID] = *undoableOperation
	}
	return nil
}

func (s *Store) UpdateAssetLifecycle(_ context.Context, item asset.Asset, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.assets[item.ID]
	if !exists || existing.TenantID != item.TenantID || existing.InventoryID != item.InventoryID {
		return ports.ErrForbidden
	}
	if existing.Kind != item.Kind || existing.Title != item.Title || existing.Description != item.Description || existing.ParentAssetID != item.ParentAssetID || existing.CustomAssetTypeID != item.CustomAssetTypeID || !existing.CustomFields.Equal(item.CustomFields) {
		return ports.ErrForbidden
	}
	if existing.LifecycleState == asset.LifecycleStateActive && item.LifecycleState == asset.LifecycleStateArchived {
		for _, child := range s.assets {
			if child.TenantID == item.TenantID && child.InventoryID == item.InventoryID && child.ParentAssetID == item.ID && child.LifecycleState == asset.LifecycleStateActive {
				return ports.ErrForbidden
			}
		}
	} else if existing.LifecycleState == asset.LifecycleStateArchived && item.LifecycleState == asset.LifecycleStateActive {
		if item.ParentAssetID.String() != "" {
			parent, ok := s.assets[item.ParentAssetID]
			if !ok || parent.TenantID != item.TenantID || parent.InventoryID != item.InventoryID || parent.LifecycleState != asset.LifecycleStateActive {
				return ports.ErrForbidden
			}
		}
	} else {
		return ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	if undoableOperation != nil {
		if _, exists := s.undoables[undoableOperation.ID]; exists {
			return ports.ErrConflict
		}
	}
	s.assets[item.ID] = item
	s.auditRecords[auditRecord.ID] = auditRecord
	if undoableOperation != nil {
		s.undoables[undoableOperation.ID] = *undoableOperation
	}
	return nil
}

func (s *Store) DeleteAsset(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.assets[assetID]
	if !ok || item.TenantID.String() != tenantID.String() || item.InventoryID.String() != inventoryID.String() {
		return ports.ErrForbidden
	}
	for _, child := range s.assets {
		if child.TenantID.String() == tenantID.String() && child.InventoryID.String() == inventoryID.String() && child.ParentAssetID == assetID && child.LifecycleState == asset.LifecycleStateActive {
			return ports.ErrForbidden
		}
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.auditRecords[auditRecord.ID] = auditRecord
	delete(s.assets, assetID)
	return nil
}

func (s *Store) AssetByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (asset.Asset, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.assets[assetID]
	if !ok || item.TenantID != asset.TenantID(tenantID.String()) || item.InventoryID != asset.InventoryID(inventoryID.String()) {
		return asset.Asset{}, false, nil
	}
	return item, true, nil
}

func (s *Store) AssetHasActiveChildren(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, item := range s.assets {
		if item.TenantID == asset.TenantID(tenantID.String()) && item.InventoryID == asset.InventoryID(inventoryID.String()) && item.ParentAssetID == assetID && item.LifecycleState == asset.LifecycleStateActive {
			return true, nil
		}
	}
	return false, nil
}

func (s *Store) ListAssetsByInventory(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.AssetListPageRequest) ([]asset.Asset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []asset.Asset{}
	sortOrder := page.Sort
	if sortOrder == "" {
		sortOrder = ports.AssetListSortIDAsc
	}
	for _, item := range s.assets {
		if item.TenantID == asset.TenantID(tenantID.String()) && item.InventoryID == asset.InventoryID(inventoryID.String()) && assetListCursorMatches(item, page, sortOrder) && assetLifecycleMatches(item.LifecycleState, page.LifecycleFilter) {
			items = append(items, item)
		}
	}
	switch sortOrder {
	case ports.AssetListSortIDAsc:
		sort.Slice(items, func(left int, right int) bool {
			return items[left].ID.String() < items[right].ID.String()
		})
	case ports.AssetListSortUpdatedDesc:
		sort.Slice(items, func(left int, right int) bool {
			if items[left].UpdatedAt.Equal(items[right].UpdatedAt) {
				return items[left].ID.String() > items[right].ID.String()
			}
			return items[left].UpdatedAt.After(items[right].UpdatedAt)
		})
	default:
		return nil, ports.ErrForbidden
	}
	if page.Limit > 0 && len(items) > page.Limit {
		items = items[:page.Limit]
	}
	return items, nil
}

func assetListCursorMatches(item asset.Asset, page ports.AssetListPageRequest, sortOrder ports.AssetListSort) bool {
	if page.AfterAssetID.String() == "" {
		return true
	}
	if sortOrder == ports.AssetListSortUpdatedDesc {
		return item.UpdatedAt.Before(page.AfterUpdatedAt) || (item.UpdatedAt.Equal(page.AfterUpdatedAt) && item.ID.String() < page.AfterAssetID.String())
	}
	return item.ID.String() > page.AfterAssetID.String()
}

func assetLifecycleMatches(state asset.LifecycleState, filter ports.AssetLifecycleFilter) bool {
	switch filter {
	case "", ports.AssetLifecycleFilterActive:
		return state == asset.LifecycleStateActive
	case ports.AssetLifecycleFilterArchived:
		return state == asset.LifecycleStateArchived
	case ports.AssetLifecycleFilterAll:
		return true
	default:
		return false
	}
}
