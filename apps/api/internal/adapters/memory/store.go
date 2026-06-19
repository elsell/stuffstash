package memory

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type Store struct {
	mu           sync.RWMutex
	tenants      map[tenant.ID]tenant.Tenant
	inventories  map[inventory.InventoryID]inventory.Inventory
	accessGrants map[string]ports.InventoryAccessGrant
	customFields map[customfield.ID]customfield.Definition
	assets       map[asset.ID]asset.Asset
	auditRecords map[audit.ID]audit.Record
	outbox       map[string]ports.AuthorizationOutboxEvent
}

func NewStore() *Store {
	return &Store{
		tenants:      map[tenant.ID]tenant.Tenant{},
		inventories:  map[inventory.InventoryID]inventory.Inventory{},
		accessGrants: map[string]ports.InventoryAccessGrant{},
		customFields: map[customfield.ID]customfield.Definition{},
		assets:       map[asset.ID]asset.Asset{},
		auditRecords: map[audit.ID]audit.Record{},
		outbox:       map[string]ports.AuthorizationOutboxEvent{},
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

	if _, exists := s.tenants[tenant.ID(definition.TenantID.String())]; !exists {
		return ports.ErrForbidden
	}
	if definition.Scope == customfield.ScopeInventory {
		item, ok := s.inventories[inventory.InventoryID(definition.InventoryID.String())]
		if !ok || item.TenantID.String() != definition.TenantID.String() {
			return ports.ErrForbidden
		}
	}
	for _, existing := range s.customFields {
		if customfield.DefinitionsConflict(existing, definition) {
			return ports.ErrConflict
		}
	}
	s.customFields[definition.ID] = definition
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
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
	for _, auditRecord := range auditRecords {
		s.auditRecords[auditRecord.ID] = auditRecord
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

func (s *Store) ListAssetsByInventory(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.AssetListPageRequest) ([]asset.Asset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []asset.Asset{}
	for _, item := range s.assets {
		if item.TenantID == asset.TenantID(tenantID.String()) && item.InventoryID == asset.InventoryID(inventoryID.String()) && item.ID.String() > page.AfterAssetID.String() {
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
	s.auditRecords[record.ID] = record
	return nil
}

func (s *Store) ListTenantAuditRecords(_ context.Context, tenantID tenant.ID, page ports.AuditRecordPageRequest) ([]audit.Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []audit.Record{}
	for _, record := range s.auditRecords {
		if record.TenantID.String() == tenantID.String() && record.ID.String() > page.AfterRecordID.String() {
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
		if record.TenantID.String() == tenantID.String() && record.InventoryID.String() == inventoryID.String() && record.ID.String() > page.AfterRecordID.String() {
			items = append(items, record)
		}
	}
	return pagedAuditRecords(items, page.Limit), nil
}

func pagedAuditRecords(items []audit.Record, limit int) []audit.Record {
	sort.Slice(items, func(left int, right int) bool {
		return items[left].CursorKey() < items[right].CursorKey()
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
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
