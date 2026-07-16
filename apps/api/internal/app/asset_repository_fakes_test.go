package app

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type fakeAssetRepository struct {
	items            map[asset.ID]asset.Asset
	assetTags        map[assettag.ID]assettag.Tag
	assetTagLinks    map[asset.ID]map[assettag.ID]struct{}
	checkouts        map[asset.CheckoutID]asset.Checkout
	undoables        map[string]ports.UndoableOperation
	auditRecords     []audit.Record
	checkOutAssetErr error
	returnAssetErr   error
}

func (f *fakeAssetRepository) CreateAssetTag(_ context.Context, tag assettag.Tag, auditRecord audit.Record) error {
	if f.assetTags == nil {
		f.assetTags = map[assettag.ID]assettag.Tag{}
	}
	if _, exists := f.assetTags[tag.ID]; exists {
		return ports.ErrConflict
	}
	for _, existing := range f.assetTags {
		if existing.TenantID == tag.TenantID && existing.InventoryID == tag.InventoryID && existing.Key == tag.Key {
			return ports.ErrConflict
		}
	}
	f.assetTags[tag.ID] = tag
	f.auditRecords = append(f.auditRecords, auditRecord)
	return nil
}

func (f *fakeAssetRepository) UpdateAssetTag(_ context.Context, tag assettag.Tag, auditRecord audit.Record) error {
	if f.assetTags == nil {
		return ports.ErrForbidden
	}
	current, ok := f.assetTags[tag.ID]
	if !ok || current.TenantID != tag.TenantID || current.InventoryID != tag.InventoryID || current.Key != tag.Key || current.LifecycleState != assettag.LifecycleStateActive {
		return ports.ErrForbidden
	}
	f.assetTags[tag.ID] = tag
	f.auditRecords = append(f.auditRecords, auditRecord)
	return nil
}

func (f *fakeAssetRepository) UpdateAssetTagLifecycle(_ context.Context, tag assettag.Tag, auditRecord audit.Record) error {
	if f.assetTags == nil {
		return ports.ErrForbidden
	}
	current, ok := f.assetTags[tag.ID]
	if !ok || current.TenantID != tag.TenantID || current.InventoryID != tag.InventoryID || current.Key != tag.Key || current.LifecycleState != assettag.LifecycleStateActive {
		return ports.ErrForbidden
	}
	f.assetTags[tag.ID] = tag
	f.auditRecords = append(f.auditRecords, auditRecord)
	return nil
}

func (f *fakeAssetRepository) SetAssetTags(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, tagIDs []assettag.ID, auditRecord audit.Record) error {
	item, ok := f.items[assetID]
	if !ok || item.TenantID.String() != tenantID.String() || item.InventoryID.String() != inventoryID.String() {
		return ports.ErrForbidden
	}
	for _, tagID := range tagIDs {
		tag, ok := f.assetTags[tagID]
		if !ok || tag.TenantID.String() != tenantID.String() || tag.InventoryID.String() != inventoryID.String() || tag.LifecycleState != assettag.LifecycleStateActive {
			return ports.ErrForbidden
		}
	}
	if f.assetTagLinks == nil {
		f.assetTagLinks = map[asset.ID]map[assettag.ID]struct{}{}
	}
	links := map[assettag.ID]struct{}{}
	for _, tagID := range tagIDs {
		links[tagID] = struct{}{}
	}
	f.assetTagLinks[assetID] = links
	f.auditRecords = append(f.auditRecords, auditRecord)
	return nil
}

func (f *fakeAssetRepository) AssetTagByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, tagID assettag.ID) (assettag.Tag, bool, error) {
	tag, ok := f.assetTags[tagID]
	if !ok || tag.TenantID.String() != tenantID.String() || tag.InventoryID.String() != inventoryID.String() {
		return assettag.Tag{}, false, nil
	}
	return tag, true, nil
}

func (f *fakeAssetRepository) AssetTagByKey(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, key assettag.Key) (assettag.Tag, bool, error) {
	for _, tag := range f.assetTags {
		if tag.TenantID.String() == tenantID.String() && tag.InventoryID.String() == inventoryID.String() && tag.Key == key {
			return tag, true, nil
		}
	}
	return assettag.Tag{}, false, nil
}

func (f *fakeAssetRepository) ListAssetTags(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.AssetTagPageRequest) ([]assettag.Tag, error) {
	items := []assettag.Tag{}
	for _, tag := range f.assetTags {
		if tag.TenantID.String() == tenantID.String() && tag.InventoryID.String() == inventoryID.String() && tag.LifecycleState == assettag.LifecycleStateActive && tag.ID.String() > page.AfterTagID.String() {
			items = append(items, tag)
		}
	}
	sort.Slice(items, func(left, right int) bool {
		return items[left].ID.String() < items[right].ID.String()
	})
	if page.Limit > 0 && len(items) > page.Limit {
		items = items[:page.Limit]
	}
	return items, nil
}

func (f *fakeAssetRepository) AssetTagsByAsset(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) ([]assettag.Tag, error) {
	items, err := f.AssetTagsByAssets(ctx, tenantID, inventoryID, []asset.ID{assetID})
	if err != nil {
		return nil, err
	}
	return items[assetID], nil
}

func (f *fakeAssetRepository) AssetTagsByAssets(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetIDs []asset.ID) (map[asset.ID][]assettag.Tag, error) {
	result := map[asset.ID][]assettag.Tag{}
	for _, assetID := range assetIDs {
		result[assetID] = []assettag.Tag{}
		for tagID := range f.assetTagLinks[assetID] {
			tag, ok := f.assetTags[tagID]
			if ok && tag.TenantID.String() == tenantID.String() && tag.InventoryID.String() == inventoryID.String() && tag.LifecycleState == assettag.LifecycleStateActive {
				result[assetID] = append(result[assetID], tag)
			}
		}
		sort.Slice(result[assetID], func(left, right int) bool {
			return result[assetID][left].Key.String() < result[assetID][right].Key.String()
		})
	}
	return result, nil
}

func (f *fakeAssetRepository) CreateAsset(_ context.Context, item asset.Asset, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	if f.items == nil {
		f.items = map[asset.ID]asset.Asset{}
	}
	if f.undoables == nil {
		f.undoables = map[string]ports.UndoableOperation{}
	}
	if _, exists := f.items[item.ID]; exists {
		return errors.New("asset already exists")
	}
	if undoableOperation != nil {
		if _, exists := f.undoables[undoableOperation.ID]; exists {
			return errors.New("undoable operation already exists")
		}
	}
	if item.ParentAssetID.String() != "" {
		parent, ok := f.items[item.ParentAssetID]
		if !ok || parent.TenantID != item.TenantID || parent.InventoryID != item.InventoryID || !parent.Kind.CanContainChildren() {
			return ports.ErrForbidden
		}
	}
	f.items[item.ID] = item
	if undoableOperation != nil {
		f.undoables[undoableOperation.ID] = *undoableOperation
	}
	f.auditRecords = append(f.auditRecords, auditRecord)
	return nil
}

func (f *fakeAssetRepository) CheckOutAsset(_ context.Context, checkout asset.Checkout, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	if f.checkOutAssetErr != nil {
		return f.checkOutAssetErr
	}
	if f.items == nil {
		f.items = map[asset.ID]asset.Asset{}
	}
	if f.checkouts == nil {
		f.checkouts = map[asset.CheckoutID]asset.Checkout{}
	}
	if f.undoables == nil {
		f.undoables = map[string]ports.UndoableOperation{}
	}
	item, ok := f.items[checkout.AssetID]
	if !ok || item.TenantID != checkout.TenantID || item.InventoryID != checkout.InventoryID || item.LifecycleState != asset.LifecycleStateActive || checkout.State != asset.CheckoutStateOpen {
		return ports.ErrForbidden
	}
	for _, existing := range f.checkouts {
		if existing.TenantID == checkout.TenantID && existing.InventoryID == checkout.InventoryID && existing.AssetID == checkout.AssetID && existing.State == asset.CheckoutStateOpen {
			return ports.ErrConflict
		}
	}
	if _, exists := f.checkouts[checkout.ID]; exists {
		return ports.ErrConflict
	}
	if undoableOperation != nil {
		if _, exists := f.undoables[undoableOperation.ID]; exists {
			return ports.ErrConflict
		}
		f.undoables[undoableOperation.ID] = *undoableOperation
	}
	f.checkouts[checkout.ID] = checkout
	f.auditRecords = append(f.auditRecords, auditRecord)
	return nil
}

func (f *fakeAssetRepository) ReturnAsset(_ context.Context, expectedCurrent asset.Checkout, returned asset.Checkout, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	if f.returnAssetErr != nil {
		return f.returnAssetErr
	}
	if f.checkouts == nil {
		return ports.ErrForbidden
	}
	current, ok := f.checkouts[expectedCurrent.ID]
	if !ok || !asset.CheckoutsEquivalentForStaleCheck(current, expectedCurrent) {
		return ports.ErrConflict
	}
	if current.State != asset.CheckoutStateOpen || returned.ID != current.ID || returned.State != asset.CheckoutStateReturned {
		return ports.ErrForbidden
	}
	if f.undoables == nil {
		f.undoables = map[string]ports.UndoableOperation{}
	}
	if undoableOperation != nil {
		if _, exists := f.undoables[undoableOperation.ID]; exists {
			return ports.ErrConflict
		}
		f.undoables[undoableOperation.ID] = *undoableOperation
	}
	f.checkouts[returned.ID] = returned
	f.auditRecords = append(f.auditRecords, auditRecord)
	return nil
}

func (f *fakeAssetRepository) UpdateAssetCheckoutReturnDetails(_ context.Context, expectedCurrent asset.Checkout, updated asset.Checkout, auditRecord audit.Record) error {
	if f.checkouts == nil {
		return ports.ErrForbidden
	}
	current, ok := f.checkouts[expectedCurrent.ID]
	if !ok || !asset.CheckoutsEquivalentForStaleCheck(current, expectedCurrent) {
		return ports.ErrConflict
	}
	if current.State != asset.CheckoutStateReturned || updated.ID != current.ID || updated.State != asset.CheckoutStateReturned {
		return ports.ErrForbidden
	}
	f.checkouts[updated.ID] = updated
	f.auditRecords = append(f.auditRecords, auditRecord)
	return nil
}

func (f *fakeAssetRepository) CreateAssetWithParentPromotion(_ context.Context, promotedParent asset.Asset, parentAuditRecord audit.Record, item asset.Asset, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	if f.items == nil {
		f.items = map[asset.ID]asset.Asset{}
	}
	if f.undoables == nil {
		f.undoables = map[string]ports.UndoableOperation{}
	}
	existingParent, ok := f.items[promotedParent.ID]
	if !ok || existingParent.TenantID != promotedParent.TenantID || existingParent.InventoryID != promotedParent.InventoryID || existingParent.Kind != asset.KindItem || promotedParent.Kind != asset.KindContainer || promotedParent.LifecycleState != asset.LifecycleStateActive {
		return ports.ErrForbidden
	}
	f.items[promotedParent.ID] = promotedParent
	if err := f.CreateAsset(context.Background(), item, auditRecord, undoableOperation); err != nil {
		f.items[promotedParent.ID] = existingParent
		return err
	}
	f.auditRecords = append([]audit.Record{parentAuditRecord}, f.auditRecords...)
	return nil
}

func (f *fakeAssetRepository) UpdateAsset(_ context.Context, item asset.Asset, auditRecords []audit.Record, undoableOperation *ports.UndoableOperation) error {
	if f.items == nil {
		f.items = map[asset.ID]asset.Asset{}
	}
	if f.undoables == nil {
		f.undoables = map[string]ports.UndoableOperation{}
	}
	existing, exists := f.items[item.ID]
	if !exists || existing.TenantID != item.TenantID || existing.InventoryID != item.InventoryID {
		return ports.ErrForbidden
	}
	if existing.Kind != item.Kind || existing.LifecycleState != item.LifecycleState || existing.LifecycleState != asset.LifecycleStateActive {
		return ports.ErrForbidden
	}
	if item.ParentAssetID.String() != "" {
		parent, ok := f.items[item.ParentAssetID]
		if !ok || parent.TenantID != item.TenantID || parent.InventoryID != item.InventoryID || !parent.Kind.CanContainChildren() || parent.LifecycleState != asset.LifecycleStateActive {
			return ports.ErrForbidden
		}
		if parent.ID == item.ID {
			return ports.ErrForbidden
		}
		for current := parent; current.ParentAssetID.String() != ""; {
			next, ok := f.items[current.ParentAssetID]
			if !ok || next.TenantID != item.TenantID || next.InventoryID != item.InventoryID {
				return ports.ErrForbidden
			}
			if next.ID == item.ID {
				return ports.ErrForbidden
			}
			current = next
		}
	}
	f.items[item.ID] = item
	if undoableOperation != nil {
		if _, exists := f.undoables[undoableOperation.ID]; exists {
			return errors.New("undoable operation already exists")
		}
		f.undoables[undoableOperation.ID] = *undoableOperation
	}
	f.auditRecords = append(f.auditRecords, auditRecords...)
	return nil
}

func (f *fakeAssetRepository) UpdateAssetAndTags(ctx context.Context, item asset.Asset, tagIDs []assettag.ID, auditRecords []audit.Record, undoableOperation *ports.UndoableOperation) error {
	for _, tagID := range tagIDs {
		tag, ok := f.assetTags[tagID]
		if !ok || tag.TenantID.String() != item.TenantID.String() || tag.InventoryID.String() != item.InventoryID.String() || tag.LifecycleState != assettag.LifecycleStateActive {
			return ports.ErrForbidden
		}
	}
	if err := f.UpdateAsset(ctx, item, auditRecords, undoableOperation); err != nil {
		return err
	}
	links := map[assettag.ID]struct{}{}
	for _, tagID := range tagIDs {
		links[tagID] = struct{}{}
	}
	if len(links) == 0 {
		delete(f.assetTagLinks, item.ID)
	} else {
		if f.assetTagLinks == nil {
			f.assetTagLinks = map[asset.ID]map[assettag.ID]struct{}{}
		}
		f.assetTagLinks[item.ID] = links
	}
	return nil
}

func (f *fakeAssetRepository) UpdateAssetLifecycle(_ context.Context, item asset.Asset, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	if f.items == nil {
		f.items = map[asset.ID]asset.Asset{}
	}
	if f.undoables == nil {
		f.undoables = map[string]ports.UndoableOperation{}
	}
	existing, ok := f.items[item.ID]
	if !ok || existing.TenantID != item.TenantID || existing.InventoryID != item.InventoryID {
		return ports.ErrForbidden
	}
	if existing.Kind != item.Kind || existing.Title != item.Title || existing.Description != item.Description || existing.ParentAssetID != item.ParentAssetID || existing.CustomAssetTypeID != item.CustomAssetTypeID || !existing.CustomFields.Equal(item.CustomFields) {
		return ports.ErrForbidden
	}
	if existing.LifecycleState == asset.LifecycleStateActive && item.LifecycleState == asset.LifecycleStateArchived {
		for _, child := range f.items {
			if child.TenantID == item.TenantID && child.InventoryID == item.InventoryID && child.ParentAssetID == item.ID && child.LifecycleState == asset.LifecycleStateActive {
				return ports.ErrForbidden
			}
		}
	} else if existing.LifecycleState == asset.LifecycleStateArchived && item.LifecycleState == asset.LifecycleStateActive {
		if item.ParentAssetID.String() != "" {
			parent, ok := f.items[item.ParentAssetID]
			if !ok || parent.TenantID != item.TenantID || parent.InventoryID != item.InventoryID || parent.LifecycleState != asset.LifecycleStateActive {
				return ports.ErrForbidden
			}
		}
	} else {
		return ports.ErrForbidden
	}
	f.items[item.ID] = item
	if undoableOperation != nil {
		if _, exists := f.undoables[undoableOperation.ID]; exists {
			return errors.New("undoable operation already exists")
		}
		f.undoables[undoableOperation.ID] = *undoableOperation
	}
	f.auditRecords = append(f.auditRecords, auditRecord)
	return nil
}

func (f *fakeAssetRepository) UndoableOperationByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, operationID string) (ports.UndoableOperation, bool, error) {
	operation, ok := f.undoables[operationID]
	if !ok || operation.TenantID != tenantID || operation.InventoryID != inventoryID {
		return ports.UndoableOperation{}, false, nil
	}
	return operation, true, nil
}

func (f *fakeAssetRepository) ApplyAssetUndoableOperation(_ context.Context, operationID string, direction ports.UndoableOperationDirection, expectedCurrent asset.Asset, resulting asset.Asset, auditRecord audit.Record) (ports.UndoableOperation, asset.Asset, error) {
	if f.items == nil || f.undoables == nil {
		return ports.UndoableOperation{}, asset.Asset{}, ports.ErrForbidden
	}
	operation, ok := f.undoables[operationID]
	if !ok || operation.TenantID != tenant.ID(expectedCurrent.TenantID.String()) || operation.InventoryID != inventory.InventoryID(expectedCurrent.InventoryID.String()) || operation.TargetID != expectedCurrent.ID.String() {
		return ports.UndoableOperation{}, asset.Asset{}, ports.ErrForbidden
	}
	current, ok := f.items[expectedCurrent.ID]
	if !ok || !fakeAssetsEqual(current, expectedCurrent) {
		return ports.UndoableOperation{}, asset.Asset{}, ports.ErrConflict
	}
	if operation.ReplacesTags {
		expectedTagIDs := operation.AfterTagIDs
		if direction == ports.UndoableOperationDirectionRedo {
			expectedTagIDs = operation.BeforeTagIDs
		}
		if !fakeAssetTagAssignmentsMatch(f.assetTagLinks[current.ID], expectedTagIDs) {
			return ports.UndoableOperation{}, asset.Asset{}, ports.ErrConflict
		}
	}
	if !fakeAssetsSameIdentity(current, resulting) {
		return ports.UndoableOperation{}, asset.Asset{}, ports.ErrForbidden
	}
	if resulting.ParentAssetID.String() != "" {
		parent, ok := f.items[resulting.ParentAssetID]
		if !ok || parent.TenantID != resulting.TenantID || parent.InventoryID != resulting.InventoryID || parent.LifecycleState != asset.LifecycleStateActive || !parent.Kind.CanContainChildren() || parent.ID == resulting.ID {
			return ports.UndoableOperation{}, asset.Asset{}, ports.ErrForbidden
		}
		for currentParent := parent; currentParent.ParentAssetID.String() != ""; {
			next, ok := f.items[currentParent.ParentAssetID]
			if !ok || next.TenantID != resulting.TenantID || next.InventoryID != resulting.InventoryID || next.ID == resulting.ID {
				return ports.UndoableOperation{}, asset.Asset{}, ports.ErrForbidden
			}
			currentParent = next
		}
	}
	if resulting.LifecycleState == asset.LifecycleStateArchived {
		for _, child := range f.items {
			if child.TenantID == resulting.TenantID && child.InventoryID == resulting.InventoryID && child.ParentAssetID == resulting.ID && child.LifecycleState == asset.LifecycleStateActive {
				return ports.UndoableOperation{}, asset.Asset{}, ports.ErrForbidden
			}
		}
	}
	switch direction {
	case ports.UndoableOperationDirectionUndo:
		if operation.Status != ports.UndoableOperationAvailable && operation.Status != ports.UndoableOperationRedone {
			return ports.UndoableOperation{}, asset.Asset{}, ports.ErrConflict
		}
		operation.Status = ports.UndoableOperationUndone
		operation.UndoAuditRecordID = auditRecord.ID
	case ports.UndoableOperationDirectionRedo:
		if operation.Status != ports.UndoableOperationUndone {
			return ports.UndoableOperation{}, asset.Asset{}, ports.ErrConflict
		}
		operation.Status = ports.UndoableOperationRedone
		operation.RedoAuditRecordID = auditRecord.ID
	default:
		return ports.UndoableOperation{}, asset.Asset{}, ports.ErrConflict
	}
	operation.LastAppliedAt = time.Now().UTC()
	f.items[resulting.ID] = resulting
	if operation.ReplacesTags {
		tagIDs := operation.AfterTagIDs
		if direction == ports.UndoableOperationDirectionUndo {
			tagIDs = operation.BeforeTagIDs
		}
		links := map[assettag.ID]struct{}{}
		for _, tagID := range tagIDs {
			links[tagID] = struct{}{}
		}
		if len(links) == 0 {
			delete(f.assetTagLinks, resulting.ID)
		} else {
			f.assetTagLinks[resulting.ID] = links
		}
	}
	f.undoables[operationID] = operation
	f.auditRecords = append(f.auditRecords, auditRecord)
	return operation, resulting, nil
}

func fakeAssetTagAssignmentsMatch(actual map[assettag.ID]struct{}, expected []assettag.ID) bool {
	wanted := make(map[assettag.ID]struct{}, len(expected))
	for _, tagID := range expected {
		wanted[tagID] = struct{}{}
	}
	if len(actual) != len(wanted) {
		return false
	}
	for tagID := range wanted {
		if _, ok := actual[tagID]; !ok {
			return false
		}
	}
	return true
}

func (f *fakeAssetRepository) ApplyAssetCheckoutUndoableOperation(_ context.Context, operationID string, direction ports.UndoableOperationDirection, expectedCurrent asset.Checkout, resulting asset.Checkout, auditRecord audit.Record) (ports.UndoableOperation, asset.Checkout, error) {
	if f.checkouts == nil || f.undoables == nil {
		return ports.UndoableOperation{}, asset.Checkout{}, ports.ErrForbidden
	}
	operation, ok := f.undoables[operationID]
	if !ok || operation.TenantID != tenant.ID(expectedCurrent.TenantID.String()) || operation.InventoryID != inventory.InventoryID(expectedCurrent.InventoryID.String()) || operation.TargetID != expectedCurrent.AssetID.String() {
		return ports.UndoableOperation{}, asset.Checkout{}, ports.ErrForbidden
	}
	current, ok := f.checkouts[expectedCurrent.ID]
	if !ok || !asset.CheckoutsEquivalentForStaleCheck(current, expectedCurrent) {
		return ports.UndoableOperation{}, asset.Checkout{}, ports.ErrConflict
	}
	if current.ID != resulting.ID || current.TenantID != resulting.TenantID || current.InventoryID != resulting.InventoryID || current.AssetID != resulting.AssetID {
		return ports.UndoableOperation{}, asset.Checkout{}, ports.ErrForbidden
	}
	switch direction {
	case ports.UndoableOperationDirectionUndo:
		if operation.Status != ports.UndoableOperationAvailable && operation.Status != ports.UndoableOperationRedone {
			return ports.UndoableOperation{}, asset.Checkout{}, ports.ErrConflict
		}
		operation.Status = ports.UndoableOperationUndone
		operation.UndoAuditRecordID = auditRecord.ID
	case ports.UndoableOperationDirectionRedo:
		if operation.Status != ports.UndoableOperationUndone {
			return ports.UndoableOperation{}, asset.Checkout{}, ports.ErrConflict
		}
		operation.Status = ports.UndoableOperationRedone
		operation.RedoAuditRecordID = auditRecord.ID
	default:
		return ports.UndoableOperation{}, asset.Checkout{}, ports.ErrConflict
	}
	operation.LastAppliedAt = time.Now().UTC()
	f.checkouts[resulting.ID] = resulting
	f.undoables[operationID] = operation
	f.auditRecords = append(f.auditRecords, auditRecord)
	return operation, resulting, nil
}

func fakeAssetsSameIdentity(left asset.Asset, right asset.Asset) bool {
	return left.ID == right.ID && left.TenantID == right.TenantID && left.InventoryID == right.InventoryID && left.Kind == right.Kind && left.CustomAssetTypeID == right.CustomAssetTypeID
}

func fakeAssetsEqual(left asset.Asset, right asset.Asset) bool {
	return fakeAssetsSameIdentity(left, right) &&
		left.ParentAssetID == right.ParentAssetID &&
		left.Title == right.Title &&
		left.Description == right.Description &&
		left.CustomFields.Equal(right.CustomFields) &&
		left.LifecycleState == right.LifecycleState
}

func (f *fakeAssetRepository) DeleteAsset(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, auditRecord audit.Record) error {
	if f.items == nil {
		return nil
	}
	item, ok := f.items[assetID]
	if !ok || item.TenantID.String() != tenantID.String() || item.InventoryID.String() != inventoryID.String() {
		return nil
	}
	delete(f.items, assetID)
	f.auditRecords = append(f.auditRecords, auditRecord)
	return nil
}

func (f *fakeAssetRepository) AssetByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (asset.Asset, bool, error) {
	item, ok := f.items[assetID]
	if !ok || item.TenantID != asset.TenantID(tenantID.String()) || item.InventoryID != asset.InventoryID(inventoryID.String()) {
		return asset.Asset{}, false, nil
	}
	return item, true, nil
}

func (f *fakeAssetRepository) AssetHasActiveChildren(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (bool, error) {
	for _, item := range f.items {
		if item.TenantID == asset.TenantID(tenantID.String()) && item.InventoryID == asset.InventoryID(inventoryID.String()) && item.ParentAssetID == assetID && item.LifecycleState == asset.LifecycleStateActive {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeAssetRepository) ListAssetsByInventory(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.AssetListPageRequest) ([]asset.Asset, error) {
	items := []asset.Asset{}
	for _, item := range f.items {
		if item.TenantID == asset.TenantID(tenantID.String()) && item.InventoryID == asset.InventoryID(inventoryID.String()) && item.ID.String() > page.AfterAssetID.String() && fakeAssetLifecycleMatches(item.LifecycleState, page.LifecycleFilter) {
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

func (f *fakeAssetRepository) CurrentAssetCheckout(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (asset.Checkout, bool, error) {
	for _, checkout := range f.checkouts {
		if checkout.TenantID.String() == tenantID.String() && checkout.InventoryID.String() == inventoryID.String() && checkout.AssetID == assetID && checkout.State == asset.CheckoutStateOpen {
			return checkout, true, nil
		}
	}
	return asset.Checkout{}, false, nil
}

func (f *fakeAssetRepository) CurrentAssetCheckouts(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetIDs []asset.ID) (map[asset.ID]asset.Checkout, error) {
	wanted := make(map[asset.ID]struct{}, len(assetIDs))
	for _, assetID := range assetIDs {
		wanted[assetID] = struct{}{}
	}
	checkouts := map[asset.ID]asset.Checkout{}
	for _, checkout := range f.checkouts {
		if checkout.TenantID.String() != tenantID.String() || checkout.InventoryID.String() != inventoryID.String() || checkout.State != asset.CheckoutStateOpen {
			continue
		}
		if _, ok := wanted[checkout.AssetID]; ok {
			checkouts[checkout.AssetID] = checkout
		}
	}
	return checkouts, nil
}

func (f *fakeAssetRepository) AssetCheckoutByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, checkoutID asset.CheckoutID) (asset.Checkout, bool, error) {
	checkout, ok := f.checkouts[checkoutID]
	if !ok || checkout.TenantID.String() != tenantID.String() || checkout.InventoryID.String() != inventoryID.String() {
		return asset.Checkout{}, false, nil
	}
	return checkout, true, nil
}

func (f *fakeAssetRepository) ListAssetCheckoutHistory(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, page ports.AssetCheckoutHistoryPageRequest) ([]asset.Checkout, error) {
	items := []asset.Checkout{}
	for _, checkout := range f.checkouts {
		if checkout.TenantID.String() == tenantID.String() && checkout.InventoryID.String() == inventoryID.String() && checkout.AssetID == assetID {
			items = append(items, checkout)
		}
	}
	sort.Slice(items, func(left int, right int) bool {
		if items[left].CheckedOutAt.Equal(items[right].CheckedOutAt) {
			return items[left].ID.String() > items[right].ID.String()
		}
		return items[left].CheckedOutAt.After(items[right].CheckedOutAt)
	})
	if page.Limit > 0 && len(items) > page.Limit {
		items = items[:page.Limit]
	}
	return items, nil
}

func (f *fakeAssetRepository) ListCheckedOutAssets(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.CheckedOutAssetsPageRequest) ([]ports.CheckedOutAsset, error) {
	items := []ports.CheckedOutAsset{}
	for _, checkout := range f.checkouts {
		if checkout.TenantID.String() != tenantID.String() || checkout.InventoryID.String() != inventoryID.String() || checkout.State != asset.CheckoutStateOpen {
			continue
		}
		item, ok := f.items[checkout.AssetID]
		if !ok {
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

func (f *fakeAssetRepository) HasLaterCheckout(_ context.Context, checkout asset.Checkout) (bool, error) {
	for _, candidate := range f.checkouts {
		if candidate.ID == checkout.ID || candidate.TenantID != checkout.TenantID || candidate.InventoryID != checkout.InventoryID || candidate.AssetID != checkout.AssetID {
			continue
		}
		if candidate.CheckedOutAt.After(checkout.CheckedOutAt) || (candidate.CheckedOutAt.Equal(checkout.CheckedOutAt) && candidate.ID.String() > checkout.ID.String()) {
			return true, nil
		}
	}
	return false, nil
}

func fakeAssetLifecycleMatches(state asset.LifecycleState, filter ports.AssetLifecycleFilter) bool {
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
