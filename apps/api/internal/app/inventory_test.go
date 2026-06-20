package app

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/adapters/gormstore"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"slices"
	"sort"
	"testing"
	"time"
)

type fakeAuthorizer struct {
	checkInventoryErr      error
	checkTenantErr         error
	grantTenantOwnerErr    error
	tenantOwnerGrants      []string
	inventoryOwnerGrants   []string
	inventoryViewerGrants  []string
	inventoryEditorGrants  []string
	inventoryViewerRevokes []string
	inventoryEditorRevokes []string
}

func (f *fakeAuthorizer) CheckTenant(context.Context, identity.Principal, ports.TenantPermission, tenant.ID) error {
	if f.checkTenantErr != nil {
		return f.checkTenantErr
	}
	return nil
}

func (f *fakeAuthorizer) CheckInventory(context.Context, identity.Principal, ports.InventoryPermission, inventory.InventoryID) error {
	return f.checkInventoryErr
}

func (f *fakeAuthorizer) GrantTenantOwner(_ context.Context, principal identity.Principal, tenantID tenant.ID) error {
	if f.grantTenantOwnerErr != nil {
		return f.grantTenantOwnerErr
	}
	f.tenantOwnerGrants = append(f.tenantOwnerGrants, principal.ID.String()+":"+tenantID.String())
	return nil
}

func (f *fakeAuthorizer) GrantInventoryOwner(_ context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	f.inventoryOwnerGrants = append(f.inventoryOwnerGrants, principal.ID.String()+":"+tenantID.String()+":"+inventoryID.String())
	return nil
}

func (f *fakeAuthorizer) GrantInventoryViewer(_ context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	f.inventoryViewerGrants = append(f.inventoryViewerGrants, principal.ID.String()+":"+tenantID.String()+":"+inventoryID.String())
	return nil
}

func (f *fakeAuthorizer) GrantInventoryEditor(_ context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	f.inventoryEditorGrants = append(f.inventoryEditorGrants, principal.ID.String()+":"+tenantID.String()+":"+inventoryID.String())
	return nil
}

func (f *fakeAuthorizer) RevokeInventoryViewer(_ context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	f.inventoryViewerRevokes = append(f.inventoryViewerRevokes, principal.ID.String()+":"+tenantID.String()+":"+inventoryID.String())
	return nil
}

func (f *fakeAuthorizer) RevokeInventoryEditor(_ context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	f.inventoryEditorRevokes = append(f.inventoryEditorRevokes, principal.ID.String()+":"+tenantID.String()+":"+inventoryID.String())
	return nil
}

type fakeTenantRepository struct {
	exists bool
}

func (f *fakeTenantRepository) SaveTenant(context.Context, tenant.Tenant) error {
	return nil
}

func (f *fakeTenantRepository) TenantExists(context.Context, tenant.ID) (bool, error) {
	return f.exists, nil
}

type fakeInventoryRepository struct {
	items        []inventory.Inventory
	accessGrants []ports.InventoryAccessGrant
	invitations  []ports.InventoryAccessInvitation
	auditRecords []audit.Record
	outbox       *fakeOutbox
	calls        int
	limits       []int
}

func (f *fakeInventoryRepository) InventoryByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (inventory.Inventory, bool, error) {
	for _, item := range f.items {
		if item.ID == inventoryID && item.TenantID == inventory.TenantID(tenantID.String()) {
			return item, true, nil
		}
	}
	return inventory.Inventory{}, false, nil
}

type fakeAssetRepository struct {
	items        map[asset.ID]asset.Asset
	auditRecords []audit.Record
}

type fakeCustomFieldRepository struct {
	items        []customfield.Definition
	auditRecords []audit.Record
}

type fakeOutbox struct {
	events       []ports.AuthorizationOutboxEvent
	auditRecords []audit.Record
	processed    []string
	failed       []string
	deadLettered []string
}

type fakeAuditRepository struct {
	items []audit.Record
}

func (f *fakeAuditRepository) SaveAuditRecord(_ context.Context, record audit.Record) error {
	f.items = append(f.items, record)
	return nil
}

func (f *fakeAuditRepository) hasAction(action audit.Action) bool {
	_, ok := f.recordForAction(action)
	return ok
}

func (f *fakeAuditRepository) recordForAction(action audit.Action) (audit.Record, bool) {
	for _, record := range f.items {
		if record.Action == action {
			return record, true
		}
	}
	return audit.Record{}, false
}

func (f *fakeAuditRepository) ListTenantAuditRecords(_ context.Context, tenantID tenant.ID, page ports.AuditRecordPageRequest) ([]audit.Record, error) {
	items := []audit.Record{}
	for _, record := range f.items {
		if record.TenantID.String() == tenantID.String() && fakeAuditRecordAfter(record, page.AfterOccurredAt, page.AfterRecordID) {
			items = append(items, record)
		}
	}
	return pagedFakeAuditRecords(items, page.Limit), nil
}

func (f *fakeAuditRepository) ListInventoryAuditRecords(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.AuditRecordPageRequest) ([]audit.Record, error) {
	items := []audit.Record{}
	for _, record := range f.items {
		if record.TenantID.String() == tenantID.String() && record.InventoryID.String() == inventoryID.String() && fakeAuditRecordAfter(record, page.AfterOccurredAt, page.AfterRecordID) {
			items = append(items, record)
		}
	}
	return pagedFakeAuditRecords(items, page.Limit), nil
}

func pagedFakeAuditRecords(items []audit.Record, limit int) []audit.Record {
	sort.Slice(items, func(left int, right int) bool {
		return items[left].Before(items[right])
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func fakeAuditRecordAfter(record audit.Record, occurredAt time.Time, id audit.ID) bool {
	if occurredAt.IsZero() || id.String() == "" {
		return true
	}
	if record.OccurredAt.After(occurredAt) {
		return true
	}
	return record.OccurredAt.Equal(occurredAt) && record.ID.String() > id.String()
}

func auditRecord(id string, tenantID string, inventoryID string, action audit.Action) audit.Record {
	record, ok := audit.NewRecord(
		audit.ID(id),
		audit.TenantID(tenantID),
		audit.InventoryID(inventoryID),
		audit.PrincipalID("owner"),
		action,
		audit.SourceAPI,
		audit.TargetAsset,
		id+"-target",
		time.Now(),
		"",
		map[string]string{},
	)
	if !ok {
		panic("invalid test audit record")
	}
	return record
}

func (f *fakeOutbox) SaveTenantAndEnqueueOwnerGrant(_ context.Context, eventID string, item tenant.Tenant, principal identity.Principal, auditRecord audit.Record) error {
	f.events = append(f.events, ports.AuthorizationOutboxEvent{
		ID:          eventID,
		Kind:        ports.AuthorizationOutboxGrantTenantOwner,
		PrincipalID: principal.ID,
		TenantID:    item.ID,
	})
	f.auditRecords = append(f.auditRecords, auditRecord)
	return nil
}

func (f *fakeOutbox) SaveInventoryAndEnqueueOwnerGrant(_ context.Context, eventID string, item inventory.Inventory, tenantID tenant.ID, principal identity.Principal, auditRecord audit.Record) error {
	f.events = append(f.events, ports.AuthorizationOutboxEvent{
		ID:          eventID,
		Kind:        ports.AuthorizationOutboxGrantInventoryOwner,
		PrincipalID: principal.ID,
		TenantID:    tenantID,
		InventoryID: item.ID,
	})
	f.auditRecords = append(f.auditRecords, auditRecord)
	return nil
}

func (f *fakeOutbox) ClaimAuthorizationOutboxEvent(_ context.Context, eventID string, claimID string, leaseUntil time.Time) (ports.AuthorizationOutboxEvent, bool, error) {
	now := time.Now()
	for index, event := range f.events {
		if event.ID != eventID {
			continue
		}
		if !event.DeadLetteredAt.IsZero() {
			return ports.AuthorizationOutboxEvent{}, false, nil
		}
		if !event.ClaimedUntil.IsZero() && event.ClaimedUntil.After(now) {
			return ports.AuthorizationOutboxEvent{}, false, nil
		}
		event.ClaimID = claimID
		event.ClaimedUntil = leaseUntil
		f.events[index] = event
		return event, true, nil
	}
	return ports.AuthorizationOutboxEvent{}, false, nil
}

func (f *fakeOutbox) ClaimPendingAuthorizationOutboxEvents(_ context.Context, claimID string, limit int, leaseUntil time.Time) ([]ports.AuthorizationOutboxEvent, error) {
	if limit <= 0 {
		limit = 25
	}
	now := time.Now()
	events := make([]ports.AuthorizationOutboxEvent, 0, len(f.events))
	for index, event := range f.events {
		if !event.DeadLetteredAt.IsZero() {
			continue
		}
		if !event.ClaimedUntil.IsZero() && event.ClaimedUntil.After(now) {
			continue
		}
		event.ClaimID = claimID
		event.ClaimedUntil = leaseUntil
		f.events[index] = event
		events = append(events, event)
		if len(events) == limit {
			break
		}
	}
	return events, nil
}

func (f *fakeOutbox) MarkAuthorizationOutboxEventProcessed(_ context.Context, eventID string, claimID string) error {
	for index, event := range f.events {
		if event.ID == eventID && event.ClaimID == claimID {
			f.processed = append(f.processed, eventID)
			f.events = append(f.events[:index], f.events[index+1:]...)
			return nil
		}
	}
	return ports.ErrAuthorizationOutboxClaimLost
}

func (f *fakeOutbox) MarkAuthorizationOutboxEventFailed(_ context.Context, eventID string, claimID string, _ string) error {
	for index, event := range f.events {
		if event.ID == eventID && event.ClaimID == claimID {
			f.failed = append(f.failed, eventID)
			event.ClaimID = ""
			event.ClaimedUntil = time.Time{}
			f.events[index] = event
			return nil
		}
	}
	return ports.ErrAuthorizationOutboxClaimLost
}

func (f *fakeOutbox) MarkAuthorizationOutboxEventDeadLettered(_ context.Context, eventID string, claimID string, reason string) error {
	for index, event := range f.events {
		if event.ID == eventID && event.ClaimID == claimID {
			f.deadLettered = append(f.deadLettered, eventID)
			event.DeadLetteredAt = time.Now()
			event.DeadLetterReason = reason
			event.ClaimID = ""
			event.ClaimedUntil = time.Time{}
			f.events[index] = event
			return nil
		}
	}
	return ports.ErrAuthorizationOutboxClaimLost
}

func (f *fakeInventoryRepository) SaveInventory(context.Context, inventory.Inventory) error {
	return nil
}

func (f *fakeInventoryRepository) SaveInventoryAccessGrantAndEnqueue(_ context.Context, _ string, grant ports.InventoryAccessGrant, auditRecord audit.Record) error {
	for _, existing := range f.accessGrants {
		if existing.TenantID == grant.TenantID && existing.InventoryID == grant.InventoryID && existing.CursorKey() == grant.CursorKey() {
			return nil
		}
	}
	f.accessGrants = append(f.accessGrants, grant)
	f.auditRecords = append(f.auditRecords, auditRecord)
	return nil
}

func (f *fakeInventoryRepository) DeleteInventoryAccessGrantAndClaimRevoke(_ context.Context, eventID string, claimID string, leaseUntil time.Time, grant ports.InventoryAccessGrant, auditRecord audit.Record) (ports.AuthorizationOutboxEvent, bool, error) {
	eventKind, ok := grant.Relationship.RevokeOutboxKind()
	if !ok {
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
	}
	if f.outbox != nil {
		f.outbox.events = append(f.outbox.events, event)
	}
	for index, existing := range f.accessGrants {
		if existing.TenantID == grant.TenantID && existing.InventoryID == grant.InventoryID && existing.CursorKey() == grant.CursorKey() {
			f.accessGrants = append(f.accessGrants[:index], f.accessGrants[index+1:]...)
			f.auditRecords = append(f.auditRecords, auditRecord)
			return event, true, nil
		}
	}
	return event, false, nil
}

func (f *fakeInventoryRepository) ListInventoryAccessGrants(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.InventoryAccessGrantPageRequest) ([]ports.InventoryAccessGrant, error) {
	items := []ports.InventoryAccessGrant{}
	for _, grant := range f.accessGrants {
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

func (f *fakeInventoryRepository) ListInventoriesByTenant(_ context.Context, tenantID inventory.TenantID, page ports.InventoryListPageRequest) ([]inventory.Inventory, error) {
	f.calls++
	f.limits = append(f.limits, page.Limit)
	items := []inventory.Inventory{}
	for _, item := range f.items {
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

func (f *fakeCustomFieldRepository) SaveCustomFieldDefinition(_ context.Context, definition customfield.Definition, auditRecord audit.Record) error {
	for _, existing := range f.items {
		if customfield.DefinitionsConflict(existing, definition) {
			return ports.ErrConflict
		}
	}
	f.items = append(f.items, definition)
	f.auditRecords = append(f.auditRecords, auditRecord)
	return nil
}

func (f *fakeCustomFieldRepository) UpdateCustomFieldDefinition(_ context.Context, definition customfield.Definition, auditRecord audit.Record) error {
	for index, existing := range f.items {
		if existing.ID != definition.ID || existing.TenantID != definition.TenantID {
			continue
		}
		if existing.Scope != definition.Scope || existing.InventoryID != definition.InventoryID || existing.Key != definition.Key || existing.Type != definition.Type || existing.Applicability != definition.Applicability || !slices.Equal(existing.EnumOptions, definition.EnumOptions) || !slices.Equal(existing.CustomAssetTypeIDs, definition.CustomAssetTypeIDs) {
			return ports.ErrForbidden
		}
		f.items[index] = definition
		f.auditRecords = append(f.auditRecords, auditRecord)
		return nil
	}
	return ports.ErrForbidden
}

func (f *fakeCustomFieldRepository) CustomFieldDefinitionByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, definitionID customfield.ID) (customfield.Definition, bool, error) {
	for _, item := range f.items {
		if item.ID != definitionID || item.TenantID.String() != tenantID.String() {
			continue
		}
		if inventoryID.String() == "" {
			if item.Scope == customfield.ScopeTenant {
				return item, true, nil
			}
			continue
		}
		if item.Scope == customfield.ScopeTenant || item.InventoryID.String() == inventoryID.String() {
			return item, true, nil
		}
	}
	return customfield.Definition{}, false, nil
}

func (f *fakeCustomFieldRepository) ListTenantCustomFieldDefinitions(_ context.Context, tenantID tenant.ID, page ports.CustomFieldDefinitionPageRequest) ([]customfield.Definition, error) {
	items := []customfield.Definition{}
	for _, item := range f.items {
		if item.TenantID.String() == tenantID.String() && item.Scope == customfield.ScopeTenant && item.CursorKey() > page.AfterDefinitionKey {
			items = append(items, item)
		}
	}
	return pagedFakeCustomFieldDefinitions(items, page.Limit), nil
}

func (f *fakeCustomFieldRepository) ListInventoryCustomFieldDefinitions(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.CustomFieldDefinitionPageRequest) ([]customfield.Definition, error) {
	items := []customfield.Definition{}
	for _, item := range f.items {
		if item.TenantID.String() != tenantID.String() || item.CursorKey() <= page.AfterDefinitionKey {
			continue
		}
		if item.Scope == customfield.ScopeTenant || item.InventoryID.String() == inventoryID.String() {
			items = append(items, item)
		}
	}
	return pagedFakeCustomFieldDefinitions(items, page.Limit), nil
}

func (f *fakeCustomFieldRepository) ListEffectiveCustomFieldDefinitions(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) ([]customfield.Definition, error) {
	if inventoryID.String() == "" {
		return f.ListTenantCustomFieldDefinitions(ctx, tenantID, ports.CustomFieldDefinitionPageRequest{})
	}
	return f.ListInventoryCustomFieldDefinitions(ctx, tenantID, inventoryID, ports.CustomFieldDefinitionPageRequest{})
}

func pagedFakeCustomFieldDefinitions(items []customfield.Definition, limit int) []customfield.Definition {
	sort.Slice(items, func(left int, right int) bool {
		return items[left].CursorKey() < items[right].CursorKey()
	})
	if limit > 0 && len(items) > limit {
		return items[:limit]
	}
	return items
}

type selectiveInventoryAuthorizer struct {
	forbidden map[inventory.InventoryID]struct{}
}

func (s *selectiveInventoryAuthorizer) CheckTenant(context.Context, identity.Principal, ports.TenantPermission, tenant.ID) error {
	return nil
}

func (s *selectiveInventoryAuthorizer) CheckInventory(_ context.Context, _ identity.Principal, _ ports.InventoryPermission, inventoryID inventory.InventoryID) error {
	if _, ok := s.forbidden[inventoryID]; ok {
		return ports.ErrForbidden
	}
	return nil
}

func (s *selectiveInventoryAuthorizer) GrantTenantOwner(context.Context, identity.Principal, tenant.ID) error {
	return nil
}

func (s *selectiveInventoryAuthorizer) GrantInventoryOwner(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (s *selectiveInventoryAuthorizer) GrantInventoryViewer(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (s *selectiveInventoryAuthorizer) GrantInventoryEditor(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (s *selectiveInventoryAuthorizer) RevokeInventoryViewer(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (s *selectiveInventoryAuthorizer) RevokeInventoryEditor(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (f *fakeAssetRepository) CreateAsset(_ context.Context, item asset.Asset, auditRecord audit.Record) error {
	if f.items == nil {
		f.items = map[asset.ID]asset.Asset{}
	}
	if _, exists := f.items[item.ID]; exists {
		return errors.New("asset already exists")
	}
	if item.ParentAssetID.String() != "" {
		parent, ok := f.items[item.ParentAssetID]
		if !ok || parent.TenantID != item.TenantID || parent.InventoryID != item.InventoryID || !parent.Kind.CanContainChildren() {
			return ports.ErrForbidden
		}
	}
	f.items[item.ID] = item
	f.auditRecords = append(f.auditRecords, auditRecord)
	return nil
}

func (f *fakeAssetRepository) UpdateAsset(_ context.Context, item asset.Asset, auditRecords []audit.Record) error {
	if f.items == nil {
		f.items = map[asset.ID]asset.Asset{}
	}
	existing, exists := f.items[item.ID]
	if !exists || existing.TenantID != item.TenantID || existing.InventoryID != item.InventoryID {
		return ports.ErrForbidden
	}
	if existing.Kind != item.Kind || existing.LifecycleState != item.LifecycleState {
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
	f.auditRecords = append(f.auditRecords, auditRecords...)
	return nil
}

func (f *fakeAssetRepository) UpdateAssetLifecycle(_ context.Context, item asset.Asset, auditRecord audit.Record) error {
	if f.items == nil {
		f.items = map[asset.ID]asset.Asset{}
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

type fakeObserver struct {
	events []ports.Event
}

func (f *fakeObserver) Record(_ context.Context, event ports.Event) {
	f.events = append(f.events, event)
}

func (f *fakeObserver) hasEvent(name ports.EventName) bool {
	for _, event := range f.events {
		if event.Name == name {
			return true
		}
	}
	return false
}

func (f *fakeObserver) eventNamed(name ports.EventName) (ports.Event, bool) {
	for _, event := range f.events {
		if event.Name == name {
			return event, true
		}
	}
	return ports.Event{}, false
}

type fakeIDGenerator struct {
	ids []string
}

func (f *fakeIDGenerator) NewID() string {
	if len(f.ids) == 0 {
		return "fixed-id"
	}
	id := f.ids[0]
	f.ids = f.ids[1:]
	return id
}

func newAppTestGORMStore(t *testing.T, ctx context.Context) gormstore.Store {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite fake: %v", err)
	}
	if err := gormstore.Migrate(ctx, db); err != nil {
		t.Fatalf("migrate sqlite fake: %v", err)
	}

	return gormstore.NewStore(db)
}

func inventoryItem(id string, tenantID string, name string) inventory.Inventory {
	inventoryName, ok := inventory.NewName(name)
	if !ok {
		panic("invalid test inventory name")
	}
	return inventory.Inventory{
		ID:       inventory.InventoryID(id),
		TenantID: inventory.TenantID(tenantID),
		Name:     inventoryName,
	}
}

func customFieldDefinition(id string, tenantID string, inventoryID string, scope customfield.Scope, keyValue string, fieldType customfield.FieldType, rawOptions []string) customfield.Definition {
	definitionID, ok := customfield.NewID(id)
	if !ok {
		panic("invalid custom field definition id")
	}
	key, ok := customfield.NewKey(keyValue)
	if !ok {
		panic("invalid custom field key")
	}
	displayName, ok := customfield.NewDisplayName("Field " + keyValue)
	if !ok {
		panic("invalid custom field display name")
	}
	options := make([]customfield.Key, 0, len(rawOptions))
	for _, raw := range rawOptions {
		option, ok := customfield.NewKey(raw)
		if !ok {
			panic("invalid custom field enum option")
		}
		options = append(options, option)
	}
	definition, ok := customfield.NewDefinition(
		definitionID,
		customfield.TenantID(tenantID),
		customfield.InventoryID(inventoryID),
		scope,
		key,
		displayName,
		fieldType,
		options,
		customfield.ApplicabilityAllAssets,
		nil,
	)
	if !ok {
		panic("invalid custom field definition")
	}
	return definition
}

func assetItem(id string, tenantID string, inventoryID string, kind asset.Kind, parentID string) asset.Asset {
	title, ok := asset.NewTitle("Asset " + id)
	if !ok {
		panic("invalid test asset title")
	}
	parent := asset.ID("")
	if parentID != "" {
		var parentOK bool
		parent, parentOK = asset.NewID(parentID)
		if !parentOK {
			panic("invalid parent id")
		}
	}
	return asset.Asset{
		ID:             asset.ID(id),
		TenantID:       asset.TenantID(tenantID),
		InventoryID:    asset.InventoryID(inventoryID),
		ParentAssetID:  parent,
		Kind:           kind,
		Title:          title,
		CustomFields:   asset.NewEmptyCustomFields(),
		LifecycleState: asset.LifecycleStateActive,
	}
}
