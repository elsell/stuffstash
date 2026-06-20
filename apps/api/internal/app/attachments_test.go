package app

import (
	"context"
	"errors"
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

func TestCreateAttachmentDeletesBlobWhenMetadataSaveFails(t *testing.T) {
	blobStore := &recordingBlobStorage{}
	application := New(Dependencies{
		Observer:             noopObserver{},
		Authorizer:           allowInventoryAuthorizer{},
		Tenants:              attachmentTenantRepository{},
		TenantUnitOfWork:     attachmentTenantRepository{},
		Inventories:          attachmentInventoryRepository{},
		InventoryUnitOfWork:  attachmentInventoryRepository{},
		Assets:               attachmentAssetRepository{},
		Attachments:          failingAttachmentRepository{},
		AttachmentUnitOfWork: failingAttachmentRepository{},
		Blobs:                blobStore,
		IDs:                  &attachmentIDGenerator{ids: []string{"attachment-one", "audit-one"}},
		MaxAttachmentBytes:   32,
	})

	_, err := application.CreateAttachment(context.Background(), CreateAttachmentInput{
		Principal:   identity.Principal{ID: "owner"},
		Source:      audit.SourceAPI,
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("asset-one"),
		FileName:    "receipt.png",
		ContentType: "image/png",
		Content:     []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'},
	})
	if !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("expected repository conflict, got %v", err)
	}
	if !blobStore.put || !blobStore.deleted {
		t.Fatalf("expected blob put and compensating delete, got put=%t deleted=%t", blobStore.put, blobStore.deleted)
	}
}

func TestCreateAttachmentRejectsContentTypeMismatch(t *testing.T) {
	application := New(Dependencies{
		Observer:             noopObserver{},
		Authorizer:           allowInventoryAuthorizer{},
		Tenants:              attachmentTenantRepository{},
		TenantUnitOfWork:     attachmentTenantRepository{},
		Inventories:          attachmentInventoryRepository{},
		InventoryUnitOfWork:  attachmentInventoryRepository{},
		Assets:               attachmentAssetRepository{},
		Attachments:          failingAttachmentRepository{},
		AttachmentUnitOfWork: failingAttachmentRepository{},
		Blobs:                &recordingBlobStorage{},
		IDs:                  &attachmentIDGenerator{ids: []string{"attachment-one", "audit-one"}},
		MaxAttachmentBytes:   32,
	})

	_, err := application.CreateAttachment(context.Background(), CreateAttachmentInput{
		Principal:   identity.Principal{ID: "owner"},
		Source:      audit.SourceAPI,
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("asset-one"),
		FileName:    "receipt.png",
		ContentType: "image/png",
		Content:     []byte("not a png"),
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

type noopObserver struct{}

func (noopObserver) Record(context.Context, ports.Event) {}

type allowInventoryAuthorizer struct{}

func (allowInventoryAuthorizer) CheckTenant(context.Context, identity.Principal, ports.TenantPermission, tenant.ID) error {
	return nil
}

func (allowInventoryAuthorizer) CheckInventory(context.Context, identity.Principal, ports.InventoryPermission, inventory.InventoryID) error {
	return nil
}

func (allowInventoryAuthorizer) ListViewableInventoryIDs(_ context.Context, _ identity.Principal, _ tenant.ID, candidates []inventory.InventoryID) ([]inventory.InventoryID, error) {
	return append([]inventory.InventoryID{}, candidates...), nil
}

func (allowInventoryAuthorizer) GrantTenantOwner(context.Context, identity.Principal, tenant.ID) error {
	return nil
}

func (allowInventoryAuthorizer) GrantInventoryOwner(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (allowInventoryAuthorizer) GrantInventoryViewer(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (allowInventoryAuthorizer) GrantInventoryEditor(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (allowInventoryAuthorizer) RevokeInventoryViewer(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (allowInventoryAuthorizer) RevokeInventoryEditor(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

type attachmentAssetRepository struct{}

type attachmentTenantRepository struct{}

func (attachmentTenantRepository) SaveTenant(context.Context, tenant.Tenant) error {
	return nil
}

func (attachmentTenantRepository) TenantExists(context.Context, tenant.ID) (bool, error) {
	return true, nil
}

func (attachmentTenantRepository) TenantByID(_ context.Context, tenantID tenant.ID) (tenant.Tenant, bool, error) {
	name, _ := tenant.NewName("Tenant")
	return tenant.Tenant{ID: tenantID, Name: name, LifecycleState: tenant.LifecycleStateActive}, true, nil
}

func (attachmentTenantRepository) UpdateTenant(context.Context, tenant.Tenant, audit.Record) error {
	return nil
}

func (attachmentTenantRepository) UpdateTenantLifecycle(context.Context, tenant.Tenant, audit.Record) error {
	return nil
}

func (attachmentTenantRepository) DeleteTenant(context.Context, tenant.ID, audit.Record) error {
	return nil
}

type attachmentInventoryRepository struct{}

func (attachmentInventoryRepository) SaveInventory(context.Context, inventory.Inventory) error {
	return nil
}

func (attachmentInventoryRepository) InventoryByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (inventory.Inventory, bool, error) {
	return inventory.Inventory{
		ID:       inventoryID,
		TenantID: inventory.TenantID(tenantID.String()),
	}, true, nil
}

func (attachmentInventoryRepository) UpdateInventory(context.Context, inventory.Inventory, audit.Record) error {
	return nil
}

func (attachmentInventoryRepository) UpdateInventoryLifecycle(context.Context, inventory.Inventory, audit.Record) error {
	return nil
}

func (attachmentInventoryRepository) DeleteInventory(context.Context, tenant.ID, inventory.InventoryID, audit.Record) error {
	return nil
}

func (attachmentInventoryRepository) InventoryHasActiveAssets(context.Context, tenant.ID, inventory.InventoryID) (bool, error) {
	return false, nil
}

func (attachmentInventoryRepository) ListInventoriesByTenant(context.Context, inventory.TenantID, ports.InventoryListPageRequest) ([]inventory.Inventory, error) {
	return nil, nil
}

func (attachmentInventoryRepository) SaveInventoryAccessGrantAndEnqueue(context.Context, string, ports.InventoryAccessGrant, audit.Record) error {
	return nil
}

func (attachmentInventoryRepository) DeleteInventoryAccessGrantAndClaimRevoke(context.Context, string, string, time.Time, ports.InventoryAccessGrant, audit.Record) (ports.AuthorizationOutboxEvent, bool, error) {
	return ports.AuthorizationOutboxEvent{}, false, nil
}

func (attachmentInventoryRepository) InventoryAccessGrantByID(context.Context, tenant.ID, inventory.InventoryID, identity.PrincipalID, ports.InventoryAccessRelationship) (ports.InventoryAccessGrant, bool, error) {
	return ports.InventoryAccessGrant{}, false, nil
}

func (attachmentInventoryRepository) SaveInventoryAccessInvitation(context.Context, ports.InventoryAccessInvitation, audit.Record) (ports.InventoryAccessInvitation, error) {
	return ports.InventoryAccessInvitation{}, nil
}

func (attachmentInventoryRepository) AcceptInventoryAccessInvitationAndEnqueue(context.Context, tenant.ID, inventory.InventoryID, string, string, identity.Principal, string, audit.Record) (ports.InventoryAccessInvitation, ports.InventoryAccessGrant, error) {
	return ports.InventoryAccessInvitation{}, ports.InventoryAccessGrant{}, nil
}

func (attachmentInventoryRepository) RevokeInventoryAccessInvitation(context.Context, tenant.ID, inventory.InventoryID, string, audit.Record) (bool, error) {
	return false, nil
}

func (attachmentInventoryRepository) InventoryAccessInvitationByID(context.Context, tenant.ID, inventory.InventoryID, string) (ports.InventoryAccessInvitation, bool, error) {
	return ports.InventoryAccessInvitation{}, false, nil
}

func (attachmentInventoryRepository) CancelInventoryAccessInvitation(context.Context, tenant.ID, inventory.InventoryID, string, audit.Record) (bool, error) {
	return false, nil
}

func (attachmentInventoryRepository) DeleteInventoryAccessInvitation(context.Context, tenant.ID, inventory.InventoryID, string, audit.Record) (bool, error) {
	return false, nil
}

func (attachmentInventoryRepository) ListInventoryAccessGrants(context.Context, tenant.ID, inventory.InventoryID, ports.InventoryAccessGrantPageRequest) ([]ports.InventoryAccessGrant, error) {
	return nil, nil
}

func (attachmentAssetRepository) CreateAsset(context.Context, asset.Asset, audit.Record, *ports.UndoableOperation) error {
	return nil
}

func (attachmentAssetRepository) UpdateAsset(context.Context, asset.Asset, []audit.Record, *ports.UndoableOperation) error {
	return nil
}

func (attachmentAssetRepository) UpdateAssetLifecycle(context.Context, asset.Asset, audit.Record, *ports.UndoableOperation) error {
	return nil
}

func (attachmentAssetRepository) DeleteAsset(context.Context, tenant.ID, inventory.InventoryID, asset.ID, audit.Record) error {
	return nil
}

func (attachmentAssetRepository) AssetByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (asset.Asset, bool, error) {
	return asset.Asset{
		ID:             assetID,
		TenantID:       asset.TenantID(tenantID.String()),
		InventoryID:    asset.InventoryID(inventoryID.String()),
		LifecycleState: asset.LifecycleStateActive,
	}, true, nil
}

func (attachmentAssetRepository) AssetHasActiveChildren(context.Context, tenant.ID, inventory.InventoryID, asset.ID) (bool, error) {
	return false, nil
}

func (attachmentAssetRepository) ListAssetsByInventory(context.Context, tenant.ID, inventory.InventoryID, ports.AssetListPageRequest) ([]asset.Asset, error) {
	return nil, nil
}

type failingAttachmentRepository struct{}

func (failingAttachmentRepository) SaveAttachment(context.Context, media.Attachment, audit.Record) error {
	return ports.ErrConflict
}

func (failingAttachmentRepository) UpdateAttachmentLifecycle(context.Context, media.Attachment, audit.Record) error {
	return ports.ErrConflict
}

func (failingAttachmentRepository) DeleteAttachmentAndEnqueueBlobDeletion(context.Context, string, tenant.ID, inventory.InventoryID, asset.ID, media.ID, audit.Record) (media.Attachment, bool, error) {
	return media.Attachment{}, false, ports.ErrConflict
}

func (failingAttachmentRepository) AttachmentByID(context.Context, tenant.ID, inventory.InventoryID, asset.ID, media.ID) (media.Attachment, bool, error) {
	return media.Attachment{}, false, nil
}

func (failingAttachmentRepository) ListAttachmentsByAsset(context.Context, tenant.ID, inventory.InventoryID, asset.ID, ports.AttachmentListPageRequest) ([]media.Attachment, error) {
	return nil, nil
}

type recordingBlobStorage struct {
	put     bool
	deleted bool
}

func (r *recordingBlobStorage) PutBlob(context.Context, media.StorageKey, media.ContentType, []byte) error {
	r.put = true
	return nil
}

func (r *recordingBlobStorage) GetBlob(context.Context, media.StorageKey) ([]byte, error) {
	return nil, ports.ErrBlobNotFound
}

func (r *recordingBlobStorage) DeleteBlob(context.Context, media.StorageKey) error {
	r.deleted = true
	return nil
}

type attachmentIDGenerator struct {
	ids []string
}

func (g *attachmentIDGenerator) NewID() string {
	id := g.ids[0]
	g.ids = g.ids[1:]
	return id
}
