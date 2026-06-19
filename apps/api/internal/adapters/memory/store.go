package memory

import (
	"context"
	"errors"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type Store struct {
	mu               sync.RWMutex
	tenants          map[tenant.ID]tenant.Tenant
	inventories      map[inventory.InventoryID]inventory.Inventory
	accessGrants     map[string]ports.InventoryAccessGrant
	customAssetTypes map[customfield.AssetTypeID]customfield.AssetType
	customFields     map[customfield.ID]customfield.Definition
	assets           map[asset.ID]asset.Asset
	attachments      map[media.ID]media.Attachment
	blobs            map[media.StorageKey][]byte
	auditRecords     map[audit.ID]audit.Record
	outbox           map[string]ports.AuthorizationOutboxEvent
}

func NewStore() *Store {
	return &Store{
		tenants:          map[tenant.ID]tenant.Tenant{},
		inventories:      map[inventory.InventoryID]inventory.Inventory{},
		accessGrants:     map[string]ports.InventoryAccessGrant{},
		customAssetTypes: map[customfield.AssetTypeID]customfield.AssetType{},
		customFields:     map[customfield.ID]customfield.Definition{},
		assets:           map[asset.ID]asset.Asset{},
		attachments:      map[media.ID]media.Attachment{},
		blobs:            map[media.StorageKey][]byte{},
		auditRecords:     map[audit.ID]audit.Record{},
		outbox:           map[string]ports.AuthorizationOutboxEvent{},
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

func (s *Store) InventoryByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (inventory.Inventory, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.inventories[inventoryID]
	if !ok || item.TenantID != inventory.TenantID(tenantID.String()) {
		return inventory.Inventory{}, false, nil
	}
	return item, true, nil
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

func (s *Store) SaveInventoryAndEnqueueOwnerGrant(_ context.Context, eventID string, item inventory.Inventory, tenantID tenant.ID, principal identity.Principal, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
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

func (s *Store) ClaimPendingAuthorizationOutboxEvents(_ context.Context, claimID string, limit int, leaseUntil time.Time) ([]ports.AuthorizationOutboxEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if limit <= 0 {
		limit = len(s.outbox)
	}
	now := time.Now()
	events := []ports.AuthorizationOutboxEvent{}
	for _, event := range s.outbox {
		if !event.DeadLetteredAt.IsZero() {
			continue
		}
		if !event.ClaimedUntil.IsZero() && event.ClaimedUntil.After(now) {
			continue
		}
		events = append(events, event)
	}
	sort.Slice(events, func(left int, right int) bool {
		if events[left].CreatedAt.Equal(events[right].CreatedAt) {
			return events[left].ID < events[right].ID
		}
		return events[left].CreatedAt.Before(events[right].CreatedAt)
	})
	if len(events) > limit {
		events = events[:limit]
	}
	for index, event := range events {
		event.ClaimID = claimID
		event.ClaimedUntil = leaseUntil
		s.outbox[event.ID] = event
		events[index] = event
	}
	return events, nil
}

func (s *Store) MarkAuthorizationOutboxEventProcessed(_ context.Context, eventID string, claimID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.outbox[eventID]
	if !ok || event.ClaimID != claimID {
		return ports.ErrAuthorizationOutboxClaimLost
	}
	delete(s.outbox, eventID)
	return nil
}

func (s *Store) MarkAuthorizationOutboxEventFailed(_ context.Context, eventID string, claimID string, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.outbox[eventID]
	if !ok || event.ClaimID != claimID {
		return ports.ErrAuthorizationOutboxClaimLost
	}
	event.Attempts++
	event.LastError = reason
	event.ClaimID = ""
	event.ClaimedUntil = time.Time{}
	s.outbox[eventID] = event
	return nil
}

func (s *Store) MarkAuthorizationOutboxEventDeadLettered(_ context.Context, eventID string, claimID string, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.outbox[eventID]
	if !ok || event.ClaimID != claimID {
		return ports.ErrAuthorizationOutboxClaimLost
	}
	event.DeadLetteredAt = time.Now()
	event.DeadLetterReason = reason
	event.ClaimID = ""
	event.ClaimedUntil = time.Time{}
	s.outbox[eventID] = event
	return nil
}

func (s *Store) ListInventoriesByTenant(_ context.Context, tenantID inventory.TenantID, page ports.InventoryListPageRequest) ([]inventory.Inventory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []inventory.Inventory{}
	for _, item := range s.inventories {
		if item.TenantID == tenantID && item.ID.String() > page.AfterInventoryID.String() {
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

func (s *Store) SaveCustomFieldDefinition(_ context.Context, definition customfield.Definition, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.customFieldDefinitionParentIsValid(definition); err != nil {
		return err
	}
	for _, targetID := range definition.CustomAssetTypeIDs {
		target, ok := s.customAssetTypes[targetID]
		if !ok || !customFieldTargetInScope(definition, target) {
			return ports.ErrForbidden
		}
	}
	for _, existing := range s.customFields {
		if customfield.DefinitionsConflict(existing, definition) {
			return ports.ErrConflict
		}
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.customFields[definition.ID] = definition
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) UpdateCustomFieldDefinition(_ context.Context, definition customfield.Definition, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.customFields[definition.ID]
	if !exists || existing.TenantID != definition.TenantID || existing.InventoryID != definition.InventoryID || existing.Scope != definition.Scope {
		return ports.ErrForbidden
	}
	if existing.Key != definition.Key || existing.Type != definition.Type || existing.Applicability != definition.Applicability {
		return ports.ErrForbidden
	}
	if !slices.Equal(existing.EnumOptions, definition.EnumOptions) || !slices.Equal(existing.CustomAssetTypeIDs, definition.CustomAssetTypeIDs) {
		return ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.customFields[definition.ID] = definition
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) CustomFieldDefinitionByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, definitionID customfield.ID) (customfield.Definition, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	definition, ok := s.customFields[definitionID]
	if !ok || definition.TenantID.String() != tenantID.String() {
		return customfield.Definition{}, false, nil
	}
	if inventoryID.String() == "" {
		if definition.Scope != customfield.ScopeTenant {
			return customfield.Definition{}, false, nil
		}
		return definition, true, nil
	}
	if definition.Scope == customfield.ScopeInventory && definition.InventoryID.String() != inventoryID.String() {
		return customfield.Definition{}, false, nil
	}
	return definition, true, nil
}

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
	if existing.Key != assetType.Key {
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
	if !ok || assetType.TenantID.String() != tenantID.String() {
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
		if assetType.TenantID.String() == tenantID.String() && assetType.Scope == customfield.ScopeTenant && assetType.CursorKey() > page.AfterAssetTypeKey {
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
		if assetType.TenantID.String() != tenantID.String() || assetType.CursorKey() <= page.AfterAssetTypeKey {
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
		if !ok || assetType.TenantID.String() != tenantID.String() {
			continue
		}
		if assetType.Scope == customfield.ScopeInventory && assetType.InventoryID.String() != inventoryID.String() {
			continue
		}
		items = append(items, assetType)
	}
	return items, nil
}

func (s *Store) ListTenantCustomFieldDefinitions(_ context.Context, tenantID tenant.ID, page ports.CustomFieldDefinitionPageRequest) ([]customfield.Definition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []customfield.Definition{}
	for _, definition := range s.customFields {
		if definition.TenantID.String() == tenantID.String() && definition.Scope == customfield.ScopeTenant && definition.CursorKey() > page.AfterDefinitionKey {
			items = append(items, definition)
		}
	}
	return pagedCustomFieldDefinitions(items, page.Limit), nil
}

func (s *Store) ListInventoryCustomFieldDefinitions(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.CustomFieldDefinitionPageRequest) ([]customfield.Definition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []customfield.Definition{}
	for _, definition := range s.customFields {
		if definition.TenantID.String() != tenantID.String() || definition.CursorKey() <= page.AfterDefinitionKey {
			continue
		}
		if definition.Scope == customfield.ScopeTenant || definition.InventoryID.String() == inventoryID.String() {
			items = append(items, definition)
		}
	}
	return pagedCustomFieldDefinitions(items, page.Limit), nil
}

func (s *Store) ListEffectiveCustomFieldDefinitions(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) ([]customfield.Definition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []customfield.Definition{}
	for _, definition := range s.customFields {
		if definition.TenantID.String() != tenantID.String() {
			continue
		}
		if inventoryID.String() == "" {
			if definition.Scope == customfield.ScopeTenant {
				items = append(items, definition)
			}
			continue
		}
		if definition.Scope == customfield.ScopeTenant || definition.InventoryID.String() == inventoryID.String() {
			items = append(items, definition)
		}
	}
	return pagedCustomFieldDefinitions(items, 0), nil
}

func (s *Store) CreateAsset(_ context.Context, item asset.Asset, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

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
	if item.CustomAssetTypeID.String() != "" {
		assetType, ok := s.customAssetTypes[customfield.AssetTypeID(item.CustomAssetTypeID.String())]
		if !ok || assetType.TenantID.String() != item.TenantID.String() || (assetType.Scope == customfield.ScopeInventory && assetType.InventoryID.String() != item.InventoryID.String()) {
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
	return nil
}

func (s *Store) UpdateAsset(_ context.Context, item asset.Asset, auditRecords []audit.Record) error {
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
	s.assets[item.ID] = item
	for _, auditRecord := range auditRecords {
		s.auditRecords[auditRecord.ID] = auditRecord
	}
	return nil
}

func (s *Store) UpdateAssetLifecycle(_ context.Context, item asset.Asset, auditRecord audit.Record) error {
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
	s.assets[item.ID] = item
	s.auditRecords[auditRecord.ID] = auditRecord
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
	for _, item := range s.assets {
		if item.TenantID == asset.TenantID(tenantID.String()) && item.InventoryID == asset.InventoryID(inventoryID.String()) && item.ID.String() > page.AfterAssetID.String() && assetLifecycleMatches(item.LifecycleState, page.LifecycleFilter) {
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

func (s *Store) SaveAuditRecord(_ context.Context, record audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tenants[tenant.ID(record.TenantID.String())]; !exists {
		return ports.ErrForbidden
	}
	if record.InventoryID.String() != "" {
		item, ok := s.inventories[inventory.InventoryID(record.InventoryID.String())]
		if !ok || item.TenantID.String() != record.TenantID.String() {
			return ports.ErrForbidden
		}
	}
	if _, exists := s.auditRecords[record.ID]; exists {
		return ports.ErrConflict
	}
	s.auditRecords[record.ID] = record
	return nil
}

func (s *Store) SaveAttachment(_ context.Context, attachment media.Attachment, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.attachments[attachment.ID]; exists {
		return ports.ErrConflict
	}
	item, ok := s.assets[asset.ID(attachment.AssetID.String())]
	if !ok || item.TenantID.String() != attachment.TenantID.String() || item.InventoryID.String() != attachment.InventoryID.String() {
		return ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.attachments[attachment.ID] = attachment
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) AttachmentByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, attachmentID media.ID) (media.Attachment, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	attachment, ok := s.attachments[attachmentID]
	if !ok || attachment.TenantID.String() != tenantID.String() || attachment.InventoryID.String() != inventoryID.String() || attachment.AssetID.String() != assetID.String() {
		return media.Attachment{}, false, nil
	}
	return attachment, true, nil
}

func (s *Store) ListAttachmentsByAsset(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, page ports.AttachmentListPageRequest) ([]media.Attachment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []media.Attachment{}
	for _, attachment := range s.attachments {
		if attachment.TenantID.String() == tenantID.String() && attachment.InventoryID.String() == inventoryID.String() && attachment.AssetID.String() == assetID.String() && attachment.ID.String() > page.AfterAttachmentID.String() {
			items = append(items, attachment)
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

func (s *Store) PutBlob(_ context.Context, key media.StorageKey, _ media.ContentType, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.blobs[key] = append([]byte(nil), data...)
	return nil
}

func (s *Store) GetBlob(_ context.Context, key media.StorageKey) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.blobs[key]
	if !ok {
		return nil, ports.ErrBlobNotFound
	}
	return append([]byte(nil), data...), nil
}

func (s *Store) DeleteBlob(_ context.Context, key media.StorageKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.blobs, key)
	return nil
}

func (s *Store) ListTenantAuditRecords(_ context.Context, tenantID tenant.ID, page ports.AuditRecordPageRequest) ([]audit.Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []audit.Record{}
	for _, record := range s.auditRecords {
		if record.TenantID.String() == tenantID.String() && auditRecordAfter(record, page.AfterOccurredAt, page.AfterRecordID) {
			items = append(items, record)
		}
	}
	return pagedAuditRecords(items, page.Limit), nil
}

func (s *Store) ListInventoryAuditRecords(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.AuditRecordPageRequest) ([]audit.Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []audit.Record{}
	for _, record := range s.auditRecords {
		if record.TenantID.String() == tenantID.String() && record.InventoryID.String() == inventoryID.String() && auditRecordAfter(record, page.AfterOccurredAt, page.AfterRecordID) {
			items = append(items, record)
		}
	}
	return pagedAuditRecords(items, page.Limit), nil
}

func pagedAuditRecords(items []audit.Record, limit int) []audit.Record {
	sort.Slice(items, func(left int, right int) bool {
		return items[left].Before(items[right])
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func auditRecordAfter(record audit.Record, occurredAt time.Time, id audit.ID) bool {
	if occurredAt.IsZero() || id.String() == "" {
		return true
	}
	if record.OccurredAt.After(occurredAt) {
		return true
	}
	return record.OccurredAt.Equal(occurredAt) && record.ID.String() > id.String()
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

func pagedCustomFieldDefinitions(items []customfield.Definition, limit int) []customfield.Definition {
	sort.Slice(items, func(left int, right int) bool {
		return items[left].CursorKey() < items[right].CursorKey()
	})
	if limit > 0 && len(items) > limit {
		return items[:limit]
	}
	return items
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

func (s *Store) customFieldDefinitionParentIsValid(definition customfield.Definition) error {
	if _, exists := s.tenants[tenant.ID(definition.TenantID.String())]; !exists {
		return ports.ErrForbidden
	}
	if definition.Scope == customfield.ScopeInventory {
		item, ok := s.inventories[inventory.InventoryID(definition.InventoryID.String())]
		if !ok || item.TenantID.String() != definition.TenantID.String() {
			return ports.ErrForbidden
		}
	}
	return nil
}

func customFieldTargetInScope(definition customfield.Definition, target customfield.AssetType) bool {
	if target.TenantID != definition.TenantID {
		return false
	}
	if definition.Scope == customfield.ScopeTenant {
		return target.Scope == customfield.ScopeTenant
	}
	return target.Scope == customfield.ScopeTenant || target.InventoryID == definition.InventoryID
}
