package assets

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestListAssetsNormalizesZeroPaginationDefaults(t *testing.T) {
	tenantID := tenant.ID("tenant-1")
	inventoryID := inventory.InventoryID("inventory-1")
	assetID := asset.ID("asset-1")
	principal := identity.Principal{ID: identity.PrincipalID("principal-1")}
	item := asset.Asset{
		ID:             assetID,
		TenantID:       asset.TenantID(tenantID.String()),
		InventoryID:    asset.InventoryID(inventoryID.String()),
		Kind:           asset.KindItem,
		Title:          asset.Title("Hammer"),
		LifecycleState: asset.LifecycleStateActive,
	}
	service := New(Dependencies{
		Observer:   noopObserver{},
		Authorizer: allowAuthorizer{},
		Tenants:    tenantExistsRepository{},
		Inventories: inventoryRepository{item: inventory.Inventory{
			ID:             inventoryID,
			TenantID:       inventory.TenantID(tenantID.String()),
			Name:           inventory.Name("Home"),
			LifecycleState: inventory.LifecycleStateActive,
		}},
		Assets: assetRepository{items: []asset.Asset{item}},
		Audit:  auditRepository{},
		IDs:    fixedIDGenerator{},
	})

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("ListAssets panicked with zero pagination defaults: %v", recovered)
		}
	}()

	result, err := service.ListAssets(context.Background(), ListAssetsInput{
		Principal:   principal,
		TenantID:    tenantID,
		InventoryID: inventoryID,
	})
	if err != nil {
		t.Fatalf("ListAssets returned error: %v", err)
	}
	if result.Limit != 50 {
		t.Fatalf("expected normalized default limit 50, got %d", result.Limit)
	}
	if len(result.Items) != 1 || result.Items[0].ID != assetID {
		t.Fatalf("expected listed asset, got %+v", result.Items)
	}
	if result.HasMore || result.NextCursor != nil {
		t.Fatalf("expected no next page, got hasMore=%v nextCursor=%v", result.HasMore, result.NextCursor)
	}
}

func TestListAssetsIncludesPrimaryImageAttachments(t *testing.T) {
	tenantID := tenant.ID("tenant-1")
	inventoryID := inventory.InventoryID("inventory-1")
	assetID := asset.ID("asset-1")
	principal := identity.Principal{ID: identity.PrincipalID("principal-1")}
	item := asset.Asset{
		ID:             assetID,
		TenantID:       asset.TenantID(tenantID.String()),
		InventoryID:    asset.InventoryID(inventoryID.String()),
		Kind:           asset.KindItem,
		Title:          asset.Title("Water bottle"),
		LifecycleState: asset.LifecycleStateActive,
	}
	attachment := media.Attachment{
		ID:             media.ID("attachment-1"),
		TenantID:       media.TenantID(tenantID.String()),
		InventoryID:    media.InventoryID(inventoryID.String()),
		AssetID:        media.AssetID(assetID.String()),
		StorageKey:     media.StorageKey("tenant-1/inventory-1/asset-1/attachment-1.jpg"),
		FileName:       media.FileName("photo.jpg"),
		ContentType:    media.ContentTypeJPEG,
		SizeBytes:      1234,
		SHA256:         media.SHA256("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"),
		LifecycleState: media.LifecycleStateActive,
		CreatedAt:      time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC),
	}
	ref := ports.AttachmentAssetReference{InventoryID: inventoryID, AssetID: assetID}
	attachments := &recordingAttachmentRepository{primaryImages: map[ports.AttachmentAssetReference]media.Attachment{ref: attachment}}
	service := New(Dependencies{
		Observer:   noopObserver{},
		Authorizer: allowAuthorizer{},
		Tenants:    tenantExistsRepository{},
		Inventories: inventoryRepository{item: inventory.Inventory{
			ID:             inventoryID,
			TenantID:       inventory.TenantID(tenantID.String()),
			Name:           inventory.Name("Home"),
			LifecycleState: inventory.LifecycleStateActive,
		}},
		Assets:      assetRepository{items: []asset.Asset{item}},
		Attachments: attachments,
		Audit:       auditRepository{},
		IDs:         fixedIDGenerator{},
	})

	result, err := service.ListAssets(context.Background(), ListAssetsInput{
		Principal:   principal,
		TenantID:    tenantID,
		InventoryID: inventoryID,
	})
	if err != nil {
		t.Fatalf("ListAssets returned error: %v", err)
	}
	if len(attachments.assets) != 1 || attachments.assets[0].AssetID != assetID || attachments.assets[0].InventoryID != inventoryID {
		t.Fatalf("expected primary image lookup for returned asset, got %+v", attachments.assets)
	}
	if result.PrimaryPhotos[ref].ID != media.ID("attachment-1") {
		t.Fatalf("expected primary photo in result, got %+v", result.PrimaryPhotos)
	}
}

func TestLifecycleFilterRejectsUnknownValue(t *testing.T) {
	if _, err := LifecycleFilter("missing"); err == nil {
		t.Fatal("expected invalid lifecycle filter error")
	}
}

func TestListAssetsWithMissingDependenciesReturnsError(t *testing.T) {
	service := New(Dependencies{})
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("ListAssets panicked with missing dependencies: %v", recovered)
		}
	}()

	_, err := service.ListAssets(context.Background(), ListAssetsInput{})
	if err == nil {
		t.Fatal("expected missing dependency error")
	}
}

type allowAuthorizer struct{}

func (allowAuthorizer) CheckTenant(context.Context, identity.Principal, ports.TenantPermission, tenant.ID) error {
	return nil
}

func (allowAuthorizer) CheckInventory(context.Context, identity.Principal, ports.InventoryPermission, inventory.InventoryID) error {
	return nil
}

func (allowAuthorizer) ListViewableInventoryIDs(context.Context, identity.Principal, tenant.ID, []inventory.InventoryID) ([]inventory.InventoryID, error) {
	return nil, nil
}

func (allowAuthorizer) GrantTenantOwner(context.Context, identity.Principal, tenant.ID) error {
	return nil
}

func (allowAuthorizer) GrantInventoryOwner(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (allowAuthorizer) GrantInventoryViewer(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (allowAuthorizer) GrantInventoryEditor(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (allowAuthorizer) RevokeInventoryViewer(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (allowAuthorizer) RevokeInventoryEditor(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

type tenantExistsRepository struct{}

func (tenantExistsRepository) TenantByID(context.Context, tenant.ID) (tenant.Tenant, bool, error) {
	return tenant.Tenant{}, false, nil
}

func (tenantExistsRepository) TenantExists(context.Context, tenant.ID) (bool, error) {
	return true, nil
}

func (tenantExistsRepository) ListTenants(context.Context, ports.TenantListPageRequest) ([]tenant.Tenant, error) {
	return nil, nil
}

type inventoryRepository struct {
	item inventory.Inventory
}

func (r inventoryRepository) InventoryByID(context.Context, tenant.ID, inventory.InventoryID) (inventory.Inventory, bool, error) {
	return r.item, true, nil
}

func (inventoryRepository) InventoryHasActiveAssets(context.Context, tenant.ID, inventory.InventoryID) (bool, error) {
	return false, nil
}

func (inventoryRepository) ListInventoriesByTenant(context.Context, inventory.TenantID, ports.InventoryListPageRequest) ([]inventory.Inventory, error) {
	return nil, nil
}

type assetRepository struct {
	items []asset.Asset
}

func (r assetRepository) AssetByID(context.Context, tenant.ID, inventory.InventoryID, asset.ID) (asset.Asset, bool, error) {
	return asset.Asset{}, false, nil
}

func (assetRepository) AssetHasActiveChildren(context.Context, tenant.ID, inventory.InventoryID, asset.ID) (bool, error) {
	return false, nil
}

func (r assetRepository) ListAssetsByInventory(context.Context, tenant.ID, inventory.InventoryID, ports.AssetListPageRequest) ([]asset.Asset, error) {
	return r.items, nil
}

type auditRepository struct{}

func (auditRepository) SaveAuditRecord(context.Context, audit.Record) error {
	return nil
}

func (auditRepository) ListTenantAuditRecords(context.Context, tenant.ID, ports.AuditRecordPageRequest) ([]audit.Record, error) {
	return nil, nil
}

func (auditRepository) ListInventoryAuditRecords(context.Context, tenant.ID, inventory.InventoryID, ports.AuditRecordPageRequest) ([]audit.Record, error) {
	return nil, nil
}

func (auditRepository) ListAssetAuditRecords(context.Context, tenant.ID, inventory.InventoryID, string, ports.AssetAuditRecordListRequest) ([]audit.Record, error) {
	return nil, nil
}

type fixedIDGenerator struct{}

func (fixedIDGenerator) NewID() string {
	return "audit-1"
}

type recordingAttachmentRepository struct {
	assets        []ports.AttachmentAssetReference
	primaryImages map[ports.AttachmentAssetReference]media.Attachment
}

func (r *recordingAttachmentRepository) AttachmentByID(context.Context, tenant.ID, inventory.InventoryID, asset.ID, media.ID) (media.Attachment, bool, error) {
	return media.Attachment{}, false, nil
}

func (r *recordingAttachmentRepository) ListAttachmentsByAsset(context.Context, tenant.ID, inventory.InventoryID, asset.ID, ports.AttachmentListPageRequest) ([]media.Attachment, error) {
	return nil, nil
}

func (r *recordingAttachmentRepository) FirstImageAttachmentsByAssets(_ context.Context, _ tenant.ID, assets []ports.AttachmentAssetReference) (map[ports.AttachmentAssetReference]media.Attachment, error) {
	r.assets = append([]ports.AttachmentAssetReference{}, assets...)
	return r.primaryImages, nil
}
