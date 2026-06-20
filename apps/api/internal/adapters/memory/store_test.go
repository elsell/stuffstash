package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestStoreRejectsDuplicateAuditRecordIDs(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	tenantID := tenant.ID("tenant-one")
	name, ok := tenant.NewName("Home")
	if !ok {
		t.Fatalf("expected valid tenant name")
	}
	if err := store.SaveTenant(ctx, tenant.Tenant{ID: tenantID, Name: name}); err != nil {
		t.Fatalf("save tenant: %v", err)
	}

	record := memoryAuditRecord(t, "audit-one", tenantID)
	if err := store.SaveAuditRecord(ctx, record); err != nil {
		t.Fatalf("save first audit record: %v", err)
	}
	if err := store.SaveAuditRecord(ctx, record); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("expected duplicate audit conflict, got %v", err)
	}
}

func TestStoreRejectsDuplicateAuditRecordIDsInsideAssetUpdateBatch(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	saveMemoryTenant(t, ctx, store, tenantID)
	saveMemoryInventory(t, ctx, store, tenantID, inventoryID)

	item := memoryAsset(t, "asset-one", tenantID, inventoryID)
	if err := store.CreateAsset(ctx, item, memoryAuditRecord(t, "audit-create", tenantID)); err != nil {
		t.Fatalf("create asset: %v", err)
	}
	title, ok := asset.NewTitle("Moved Drill")
	if !ok {
		t.Fatalf("expected valid title")
	}
	item.Title = title

	duplicate := memoryAuditRecord(t, "audit-duplicate", tenantID)
	err := store.UpdateAsset(ctx, item, []audit.Record{duplicate, duplicate})
	if !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("expected duplicate audit conflict, got %v", err)
	}
	found, ok, err := store.AssetByID(ctx, tenantID, inventoryID, item.ID)
	if err != nil {
		t.Fatalf("find asset: %v", err)
	}
	if !ok || found.Title.String() != "Drill" {
		t.Fatalf("expected asset update to roll back, found=%t item=%+v", ok, found)
	}
}

func TestStoreRejectsArchiveCustomAssetTypeWithoutArchivedLifecycle(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	saveMemoryTenant(t, ctx, store, tenantID)
	saveMemoryInventory(t, ctx, store, tenantID, inventoryID)

	assetType := memoryCustomAssetType(t, "type-one", tenantID, inventoryID)
	if err := store.SaveCustomAssetType(ctx, assetType, memoryAuditRecord(t, "audit-create", tenantID)); err != nil {
		t.Fatalf("save custom asset type: %v", err)
	}
	if err := store.ArchiveCustomAssetType(ctx, assetType, memoryAuditRecord(t, "audit-archive", tenantID)); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected active lifecycle archive rejection, got %v", err)
	}
	archived, ok := assetType.Archive()
	if !ok {
		t.Fatalf("expected archive transition")
	}
	if err := store.ArchiveCustomAssetType(ctx, archived, memoryAuditRecord(t, "audit-archive", tenantID)); err != nil {
		t.Fatalf("archive custom asset type: %v", err)
	}
}

func saveMemoryTenant(t *testing.T, ctx context.Context, store *Store, tenantID tenant.ID) {
	t.Helper()

	name, ok := tenant.NewName("Home")
	if !ok {
		t.Fatalf("expected valid tenant name")
	}
	if err := store.SaveTenant(ctx, tenant.Tenant{ID: tenantID, Name: name}); err != nil {
		t.Fatalf("save tenant: %v", err)
	}
}

func saveMemoryInventory(t *testing.T, ctx context.Context, store *Store, tenantID tenant.ID, inventoryID inventory.InventoryID) {
	t.Helper()

	name, ok := inventory.NewName("Tools")
	if !ok {
		t.Fatalf("expected valid inventory name")
	}
	if err := store.SaveInventory(ctx, inventory.Inventory{
		ID:       inventoryID,
		TenantID: inventory.TenantID(tenantID.String()),
		Name:     name,
	}); err != nil {
		t.Fatalf("save inventory: %v", err)
	}
}

func memoryCustomAssetType(t *testing.T, id string, tenantID tenant.ID, inventoryID inventory.InventoryID) customfield.AssetType {
	t.Helper()

	assetType, ok := customfield.NewAssetType(
		customfield.AssetTypeID(id),
		customfield.TenantID(tenantID.String()),
		customfield.InventoryID(inventoryID.String()),
		customfield.ScopeInventory,
		customfield.Key("medicine"),
		customfield.DisplayName("Medicine"),
		customfield.Description(""),
	)
	if !ok {
		t.Fatalf("expected valid custom asset type")
	}
	return assetType
}

func memoryAsset(t *testing.T, id string, tenantID tenant.ID, inventoryID inventory.InventoryID) asset.Asset {
	t.Helper()

	title, ok := asset.NewTitle("Drill")
	if !ok {
		t.Fatalf("expected valid title")
	}
	return asset.Asset{
		ID:             asset.ID(id),
		TenantID:       asset.TenantID(tenantID.String()),
		InventoryID:    asset.InventoryID(inventoryID.String()),
		Kind:           asset.KindItem,
		Title:          title,
		Description:    asset.NewDescription(""),
		CustomFields:   asset.NewEmptyCustomFields(),
		LifecycleState: asset.LifecycleStateActive,
	}
}

func memoryAuditRecord(t *testing.T, id string, tenantID tenant.ID) audit.Record {
	t.Helper()

	record, ok := audit.NewRecord(
		audit.ID(id),
		audit.TenantID(tenantID.String()),
		audit.InventoryID(""),
		audit.PrincipalID("user-one"),
		audit.ActionTenantCreated,
		audit.SourceAPI,
		audit.TargetTenant,
		tenantID.String(),
		time.Now(),
		"",
		map[string]string{},
	)
	if !ok {
		t.Fatalf("expected valid audit record")
	}
	return record
}
