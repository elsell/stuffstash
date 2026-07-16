package memory

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
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

func (s *Store) CreateAssetWithParentPromotion(_ context.Context, promotedParent asset.Asset, parentAuditRecord audit.Record, item asset.Asset, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.createAssetWithParentPromotionLocked(promotedParent, parentAuditRecord, item, auditRecord, undoableOperation)
}

func (s *Store) createAssetWithParentPromotionLocked(promotedParent asset.Asset, parentAuditRecord audit.Record, item asset.Asset, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	existingParent, ok := s.assets[promotedParent.ID]
	if !ok || existingParent.TenantID != promotedParent.TenantID || existingParent.InventoryID != promotedParent.InventoryID || existingParent.Kind != asset.KindItem || promotedParent.Kind != asset.KindContainer || promotedParent.LifecycleState != asset.LifecycleStateActive {
		return ports.ErrForbidden
	}
	s.assets[promotedParent.ID] = promotedParent
	if err := s.createAssetLocked(item, auditRecord, undoableOperation); err != nil {
		s.assets[promotedParent.ID] = existingParent
		return err
	}
	if _, exists := s.auditRecords[parentAuditRecord.ID]; exists {
		s.assets[promotedParent.ID] = existingParent
		delete(s.assets, item.ID)
		if undoableOperation != nil {
			delete(s.undoables, undoableOperation.ID)
		}
		delete(s.auditRecords, auditRecord.ID)
		return ports.ErrConflict
	}
	s.auditRecords[parentAuditRecord.ID] = parentAuditRecord
	return nil
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

	return s.updateAssetLocked(asset.Asset{}, item, auditRecords, undoableOperation)
}

func (s *Store) UpdateAssetAndTags(_ context.Context, item asset.Asset, tagIDs []assettag.ID, auditRecords []audit.Record, undoableOperation *ports.UndoableOperation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := map[assettag.ID]struct{}{}
	for _, tagID := range tagIDs {
		tag, found := s.assetTags[tagID]
		if !found || tag.TenantID.String() != item.TenantID.String() || tag.InventoryID.String() != item.InventoryID.String() || tag.LifecycleState != assettag.LifecycleStateActive {
			return ports.ErrForbidden
		}
		next[tagID] = struct{}{}
	}
	if err := s.updateAssetLocked(asset.Asset{}, item, auditRecords, undoableOperation); err != nil {
		return err
	}
	if len(next) == 0 {
		delete(s.assetTagLinks, item.ID)
	} else {
		s.assetTagLinks[item.ID] = next
	}
	return nil
}

func (s *Store) updateAssetLocked(expectedCurrent asset.Asset, item asset.Asset, auditRecords []audit.Record, undoableOperation *ports.UndoableOperation) error {
	existing, exists := s.assets[item.ID]
	if !exists || existing.TenantID != item.TenantID || existing.InventoryID != item.InventoryID {
		return ports.ErrForbidden
	}
	if expectedCurrent.ID.String() != "" && !assetsEquivalentForStaleCheck(existing, expectedCurrent) {
		return ports.ErrConflict
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

func assetsEquivalentForStaleCheck(left asset.Asset, right asset.Asset) bool {
	return left.ID == right.ID &&
		left.TenantID == right.TenantID &&
		left.InventoryID == right.InventoryID &&
		left.ParentAssetID == right.ParentAssetID &&
		left.CustomAssetTypeID == right.CustomAssetTypeID &&
		left.Kind == right.Kind &&
		left.Title == right.Title &&
		left.Description == right.Description &&
		left.LifecycleState == right.LifecycleState &&
		left.CreatedAt.Equal(right.CreatedAt) &&
		left.UpdatedAt.Equal(right.UpdatedAt) &&
		left.CustomFields.Equal(right.CustomFields)
}

func (s *Store) UpdateAssetLifecycle(_ context.Context, item asset.Asset, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.updateAssetLifecycleLocked(asset.Asset{}, item, auditRecord, undoableOperation)
}

func (s *Store) updateAssetLifecycleLocked(expectedCurrent asset.Asset, item asset.Asset, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	existing, exists := s.assets[item.ID]
	if !exists || existing.TenantID != item.TenantID || existing.InventoryID != item.InventoryID {
		return ports.ErrForbidden
	}
	if expectedCurrent.ID.String() != "" && !assetsEquivalentForStaleCheck(existing, expectedCurrent) {
		return ports.ErrConflict
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
	for _, checkout := range s.checkouts {
		if checkout.TenantID.String() == tenantID.String() && checkout.InventoryID.String() == inventoryID.String() && checkout.AssetID == assetID && checkout.State == asset.CheckoutStateOpen {
			return ports.ErrForbidden
		}
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.auditRecords[auditRecord.ID] = auditRecord
	delete(s.assets, assetID)
	for checkoutID, checkout := range s.checkouts {
		if checkout.TenantID.String() == tenantID.String() && checkout.InventoryID.String() == inventoryID.String() && checkout.AssetID == assetID {
			delete(s.checkouts, checkoutID)
		}
	}
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

func (s *Store) CheckOutAsset(_ context.Context, checkout asset.Checkout, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.checkOutAssetLocked(checkout, auditRecord, undoableOperation)
}

func (s *Store) checkOutAssetLocked(checkout asset.Checkout, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	item, ok := s.assets[checkout.AssetID]
	if !ok || item.TenantID != checkout.TenantID || item.InventoryID != checkout.InventoryID || item.LifecycleState != asset.LifecycleStateActive || checkout.State != asset.CheckoutStateOpen {
		return ports.ErrForbidden
	}
	for _, existing := range s.checkouts {
		if existing.TenantID == checkout.TenantID && existing.InventoryID == checkout.InventoryID && existing.AssetID == checkout.AssetID && existing.State == asset.CheckoutStateOpen {
			return ports.ErrConflict
		}
	}
	if _, exists := s.checkouts[checkout.ID]; exists {
		return ports.ErrConflict
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	if undoableOperation != nil {
		if _, exists := s.undoables[undoableOperation.ID]; exists {
			return ports.ErrConflict
		}
	}
	s.checkouts[checkout.ID] = checkout
	s.auditRecords[auditRecord.ID] = auditRecord
	if undoableOperation != nil {
		s.undoables[undoableOperation.ID] = *undoableOperation
	}
	return nil
}

func (s *Store) ReturnAsset(_ context.Context, expectedCurrent asset.Checkout, returned asset.Checkout, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.returnAssetLocked(expectedCurrent, returned, auditRecord, undoableOperation)
}

func (s *Store) returnAssetLocked(expectedCurrent asset.Checkout, returned asset.Checkout, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	current, ok := s.checkouts[expectedCurrent.ID]
	if !ok || !asset.CheckoutsEquivalentForStaleCheck(current, expectedCurrent) {
		return ports.ErrConflict
	}
	if current.State != asset.CheckoutStateOpen || returned.ID != current.ID || returned.TenantID != current.TenantID || returned.InventoryID != current.InventoryID || returned.AssetID != current.AssetID || returned.State != asset.CheckoutStateReturned || returned.ReturnedAt.IsZero() || returned.ReturnedByPrincipal == "" {
		return ports.ErrForbidden
	}
	if _, ok := s.assets[current.AssetID]; !ok {
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
	s.checkouts[returned.ID] = returned
	s.auditRecords[auditRecord.ID] = auditRecord
	if undoableOperation != nil {
		s.undoables[undoableOperation.ID] = *undoableOperation
	}
	return nil
}

func (s *Store) UpdateAssetCheckoutReturnDetails(_ context.Context, expectedCurrent asset.Checkout, updated asset.Checkout, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, ok := s.checkouts[expectedCurrent.ID]
	if !ok || !asset.CheckoutsEquivalentForStaleCheck(current, expectedCurrent) {
		return ports.ErrConflict
	}
	if current.State != asset.CheckoutStateReturned || updated.ID != current.ID || updated.TenantID != current.TenantID || updated.InventoryID != current.InventoryID || updated.AssetID != current.AssetID || updated.State != asset.CheckoutStateReturned || updated.ReturnedAt.IsZero() || updated.ReturnedByPrincipal == "" {
		return ports.ErrForbidden
	}
	if _, ok := s.assets[current.AssetID]; !ok {
		return ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.checkouts[updated.ID] = updated
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) CurrentAssetCheckout(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (asset.Checkout, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, checkout := range s.checkouts {
		if checkout.TenantID.String() == tenantID.String() && checkout.InventoryID.String() == inventoryID.String() && checkout.AssetID == assetID && checkout.State == asset.CheckoutStateOpen {
			return checkout, true, nil
		}
	}
	return asset.Checkout{}, false, nil
}

func (s *Store) CurrentAssetCheckouts(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetIDs []asset.ID) (map[asset.ID]asset.Checkout, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	wanted := make(map[asset.ID]struct{}, len(assetIDs))
	for _, assetID := range assetIDs {
		if assetID.String() != "" {
			wanted[assetID] = struct{}{}
		}
	}
	checkouts := map[asset.ID]asset.Checkout{}
	for _, checkout := range s.checkouts {
		if checkout.TenantID.String() != tenantID.String() || checkout.InventoryID.String() != inventoryID.String() || checkout.State != asset.CheckoutStateOpen {
			continue
		}
		if _, ok := wanted[checkout.AssetID]; ok {
			checkouts[checkout.AssetID] = checkout
		}
	}
	return checkouts, nil
}

func (s *Store) AssetCheckoutByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, checkoutID asset.CheckoutID) (asset.Checkout, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	checkout, ok := s.checkouts[checkoutID]
	if !ok || checkout.TenantID.String() != tenantID.String() || checkout.InventoryID.String() != inventoryID.String() {
		return asset.Checkout{}, false, nil
	}
	return checkout, true, nil
}

func (s *Store) ListAssetCheckoutHistory(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, page ports.AssetCheckoutHistoryPageRequest) ([]asset.Checkout, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	checkouts := []asset.Checkout{}
	for _, checkout := range s.checkouts {
		if checkout.TenantID.String() == tenantID.String() && checkout.InventoryID.String() == inventoryID.String() && checkout.AssetID == assetID && checkoutHistoryCursorMatches(checkout, page) {
			checkouts = append(checkouts, checkout)
		}
	}
	sort.Slice(checkouts, func(left int, right int) bool {
		if checkouts[left].CheckedOutAt.Equal(checkouts[right].CheckedOutAt) {
			return checkouts[left].ID.String() > checkouts[right].ID.String()
		}
		return checkouts[left].CheckedOutAt.After(checkouts[right].CheckedOutAt)
	})
	if page.Limit > 0 && len(checkouts) > page.Limit {
		checkouts = checkouts[:page.Limit]
	}
	return checkouts, nil
}

func checkoutHistoryCursorMatches(checkout asset.Checkout, page ports.AssetCheckoutHistoryPageRequest) bool {
	if page.AfterCheckoutID.String() == "" {
		return true
	}
	return checkout.CheckedOutAt.Before(page.AfterCheckedOutAt) || (checkout.CheckedOutAt.Equal(page.AfterCheckedOutAt) && checkout.ID.String() < page.AfterCheckoutID.String())
}

func (s *Store) ListCheckedOutAssets(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.CheckedOutAssetsPageRequest) ([]ports.CheckedOutAsset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []ports.CheckedOutAsset{}
	for _, checkout := range s.checkouts {
		if checkout.TenantID.String() != tenantID.String() || checkout.InventoryID.String() != inventoryID.String() || checkout.State != asset.CheckoutStateOpen || !checkedOutAssetsCursorMatches(checkout, page) {
			continue
		}
		item, ok := s.assets[checkout.AssetID]
		if !ok || item.TenantID != checkout.TenantID || item.InventoryID != checkout.InventoryID {
			continue
		}
		items = append(items, ports.CheckedOutAsset{Asset: item, Checkout: checkout})
	}
	sort.Slice(items, func(left int, right int) bool {
		if items[left].Checkout.CheckedOutAt.Equal(items[right].Checkout.CheckedOutAt) {
			return items[left].Asset.ID.String() > items[right].Asset.ID.String()
		}
		return items[left].Checkout.CheckedOutAt.After(items[right].Checkout.CheckedOutAt)
	})
	if page.Limit > 0 && len(items) > page.Limit {
		items = items[:page.Limit]
	}
	return items, nil
}

func checkedOutAssetsCursorMatches(checkout asset.Checkout, page ports.CheckedOutAssetsPageRequest) bool {
	if page.AfterAssetID.String() == "" {
		return true
	}
	return checkout.CheckedOutAt.Before(page.AfterCheckedOutAt) || (checkout.CheckedOutAt.Equal(page.AfterCheckedOutAt) && checkout.AssetID.String() < page.AfterAssetID.String())
}

func (s *Store) HasLaterCheckout(_ context.Context, checkout asset.Checkout) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, candidate := range s.checkouts {
		if candidate.TenantID != checkout.TenantID || candidate.InventoryID != checkout.InventoryID || candidate.AssetID != checkout.AssetID || candidate.ID == checkout.ID {
			continue
		}
		if candidate.CheckedOutAt.After(checkout.CheckedOutAt) || (candidate.CheckedOutAt.Equal(checkout.CheckedOutAt) && candidate.ID.String() > checkout.ID.String()) {
			return true, nil
		}
	}
	return false, nil
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
