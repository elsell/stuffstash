package app

import (
	"context"
	"sort"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestArchiveCustomAssetTypeRecordsAuditAndObservability(t *testing.T) {
	customAssetTypes := &fakeCustomAssetTypeRepository{}
	observer := &fakeObserver{}
	application := New(Dependencies{
		Observer:         observer,
		Authorizer:       &fakeAuthorizer{},
		Tenants:          &fakeTenantRepository{exists: true},
		Inventories:      &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Medicine")}},
		CustomAssetTypes: customAssetTypes,
		IDs:              &fakeIDGenerator{ids: []string{"type-one", "audit-create", "audit-archive"}},
	})

	assetType, err := application.CreateInventoryCustomAssetType(context.Background(), CreateCustomAssetTypeInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Key:         "medicine",
		DisplayName: "Medicine",
	})
	if err != nil {
		t.Fatalf("create custom asset type: %v", err)
	}

	archived, err := application.ArchiveInventoryCustomAssetType(context.Background(), ArchiveCustomAssetTypeInput{
		Principal:         identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:          tenant.ID("tenant-one"),
		InventoryID:       inventory.InventoryID("inventory-one"),
		CustomAssetTypeID: assetType.ID,
	})
	if err != nil {
		t.Fatalf("archive custom asset type: %v", err)
	}
	if archived.LifecycleState != customfield.AssetTypeLifecycleArchived {
		t.Fatalf("expected archived lifecycle state, got %+v", archived)
	}

	event, ok := observer.eventNamed(ports.EventCustomAssetTypeArchived)
	if !ok {
		t.Fatalf("expected archive observability event, got %+v", observer.events)
	}
	if event.Fields["asset_type_id"] != assetType.ID.String() || event.Fields["type_key"] != "medicine" || event.Fields["scope"] != customfield.ScopeInventory.String() {
		t.Fatalf("expected safe archive event fields, got %+v", event.Fields)
	}

	record, ok := customAssetTypes.recordForAction(audit.ActionCustomAssetTypeArchived)
	if !ok {
		t.Fatalf("expected archive audit record, got %+v", customAssetTypes.auditRecords)
	}
	if record.TargetType != audit.TargetCustomAssetType || record.TargetID != assetType.ID.String() {
		t.Fatalf("expected custom asset type audit target, got %+v", record)
	}
	if record.Metadata["type_key"] != "medicine" || record.Metadata["scope"] != customfield.ScopeInventory.String() {
		t.Fatalf("expected safe archive audit metadata, got %+v", record.Metadata)
	}
}

type fakeCustomAssetTypeRepository struct {
	items        []customfield.AssetType
	auditRecords []audit.Record
}

func (f *fakeCustomAssetTypeRepository) SaveCustomAssetType(_ context.Context, assetType customfield.AssetType, auditRecord audit.Record) error {
	for _, existing := range f.items {
		if customfield.AssetTypesConflict(existing, assetType) {
			return ports.ErrConflict
		}
	}
	f.items = append(f.items, assetType)
	f.auditRecords = append(f.auditRecords, auditRecord)
	return nil
}

func (f *fakeCustomAssetTypeRepository) UpdateCustomAssetType(_ context.Context, assetType customfield.AssetType, auditRecord audit.Record) error {
	for index, existing := range f.items {
		if existing.ID != assetType.ID || existing.TenantID != assetType.TenantID || existing.InventoryID != assetType.InventoryID || existing.Scope != assetType.Scope || existing.Key != assetType.Key || !existing.IsActive() {
			continue
		}
		f.items[index] = assetType
		f.auditRecords = append(f.auditRecords, auditRecord)
		return nil
	}
	return ports.ErrForbidden
}

func (f *fakeCustomAssetTypeRepository) ArchiveCustomAssetType(_ context.Context, assetType customfield.AssetType, auditRecord audit.Record) error {
	if assetType.LifecycleState != customfield.AssetTypeLifecycleArchived {
		return ports.ErrForbidden
	}
	for index, existing := range f.items {
		if existing.ID != assetType.ID || existing.TenantID != assetType.TenantID || existing.InventoryID != assetType.InventoryID || existing.Scope != assetType.Scope || existing.Key != assetType.Key || !existing.IsActive() {
			continue
		}
		f.items[index] = assetType
		f.auditRecords = append(f.auditRecords, auditRecord)
		return nil
	}
	return ports.ErrForbidden
}

func (f *fakeCustomAssetTypeRepository) CustomAssetTypeByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetTypeID customfield.AssetTypeID) (customfield.AssetType, bool, error) {
	for _, item := range f.items {
		if item.ID != assetTypeID || item.TenantID.String() != tenantID.String() || !item.IsActive() {
			continue
		}
		if inventoryID.String() == "" {
			return item, item.Scope == customfield.ScopeTenant, nil
		}
		if item.Scope == customfield.ScopeTenant || item.InventoryID.String() == inventoryID.String() {
			return item, true, nil
		}
	}
	return customfield.AssetType{}, false, nil
}

func (f *fakeCustomAssetTypeRepository) ListTenantCustomAssetTypes(_ context.Context, tenantID tenant.ID, page ports.CustomAssetTypePageRequest) ([]customfield.AssetType, error) {
	items := []customfield.AssetType{}
	for _, item := range f.items {
		if item.TenantID.String() == tenantID.String() && item.Scope == customfield.ScopeTenant && item.IsActive() && item.CursorKey() > page.AfterAssetTypeKey {
			items = append(items, item)
		}
	}
	return pagedFakeCustomAssetTypes(items, page.Limit), nil
}

func (f *fakeCustomAssetTypeRepository) ListInventoryCustomAssetTypes(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.CustomAssetTypePageRequest) ([]customfield.AssetType, error) {
	items := []customfield.AssetType{}
	for _, item := range f.items {
		if item.TenantID.String() != tenantID.String() || !item.IsActive() || item.CursorKey() <= page.AfterAssetTypeKey {
			continue
		}
		if item.Scope == customfield.ScopeTenant || item.InventoryID.String() == inventoryID.String() {
			items = append(items, item)
		}
	}
	return pagedFakeCustomAssetTypes(items, page.Limit), nil
}

func (f *fakeCustomAssetTypeRepository) CustomAssetTypesByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, ids []customfield.AssetTypeID) ([]customfield.AssetType, error) {
	matches := []customfield.AssetType{}
	for _, id := range ids {
		for _, item := range f.items {
			if item.ID != id || item.TenantID.String() != tenantID.String() || !item.IsActive() {
				continue
			}
			if item.Scope == customfield.ScopeTenant || item.InventoryID.String() == inventoryID.String() {
				matches = append(matches, item)
			}
		}
	}
	return matches, nil
}

func (f *fakeCustomAssetTypeRepository) recordForAction(action audit.Action) (audit.Record, bool) {
	for _, record := range f.auditRecords {
		if record.Action == action {
			return record, true
		}
	}
	return audit.Record{}, false
}

func pagedFakeCustomAssetTypes(items []customfield.AssetType, limit int) []customfield.AssetType {
	sort.Slice(items, func(left int, right int) bool {
		return items[left].CursorKey() < items[right].CursorKey()
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}
