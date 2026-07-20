package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestSearchAssetsUsesAuthorizationVisibilityPort(t *testing.T) {
	search := &recordingAssetSearchRepository{}
	authorizer := &visibilityAuthorizer{
		t:        t,
		tenantID: tenant.ID("tenant-one"),
		visible:  []inventory.InventoryID{inventory.InventoryID("inventory-two")},
	}
	application := New(Dependencies{
		Observer:   &fakeObserver{},
		Authorizer: authorizer,
		Tenants:    &fakeTenantRepository{exists: true},
		Inventories: &fakeInventoryRepository{items: []inventory.Inventory{
			inventoryItem("inventory-one", "tenant-one", "Tools"),
			inventoryItem("inventory-two", "tenant-one", "Medicine"),
			inventoryItem("inventory-other", "tenant-two", "Other Tenant"),
		}},
		Search:           search,
		Audit:            &fakeAuditRepository{},
		DefaultPageLimit: 1,
		MaxPageLimit:     10,
	})

	_, err := application.SearchAssets(context.Background(), SearchAssetsInput{
		Principal: identity.Principal{ID: identity.PrincipalID("viewer")},
		TenantID:  tenant.ID("tenant-one"),
		Query:     "aspirin",
		Mode:      "exact",
	})
	if err != nil {
		t.Fatalf("search assets: %v", err)
	}

	if !authorizer.visibilityCalled {
		t.Fatalf("expected search to use authorization visibility port")
	}
	if len(authorizer.candidates) != 2 {
		t.Fatalf("expected two tenant-scoped candidate inventories, got %+v", authorizer.candidates)
	}
	for _, candidate := range authorizer.candidates {
		if candidate == inventory.InventoryID("inventory-other") {
			t.Fatalf("authorization visibility candidates must be tenant-scoped, got %+v", authorizer.candidates)
		}
	}
	if len(search.inventoryIDs) != 1 || search.inventoryIDs[0] != inventory.InventoryID("inventory-two") {
		t.Fatalf("expected search repository to receive visible inventory IDs only, got %+v", search.inventoryIDs)
	}
}

func TestSearchAssetsIncludesPrimaryImageAttachments(t *testing.T) {
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	assetID := asset.ID("asset-one")
	item := asset.Asset{
		ID:             assetID,
		TenantID:       asset.TenantID(tenantID.String()),
		InventoryID:    asset.InventoryID(inventoryID.String()),
		Kind:           asset.KindItem,
		Title:          asset.Title("Water bottle"),
		LifecycleState: asset.LifecycleStateActive,
	}
	photo := media.Attachment{
		ID:             media.ID("attachment-one"),
		TenantID:       media.TenantID(tenantID.String()),
		InventoryID:    media.InventoryID(inventoryID.String()),
		AssetID:        media.AssetID(assetID.String()),
		StorageKey:     media.StorageKey("tenant-one/inventory-one/asset-one/photo.jpg"),
		FileName:       media.FileName("photo.jpg"),
		ContentType:    media.ContentTypeJPEG,
		SizeBytes:      2048,
		SHA256:         media.SHA256("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"),
		LifecycleState: media.LifecycleStateActive,
		CreatedAt:      time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC),
	}
	ref := ports.AttachmentAssetReference{InventoryID: inventoryID, AssetID: assetID}
	attachments := &searchAttachmentRepository{primaryImages: map[ports.AttachmentAssetReference]media.Attachment{ref: photo}}
	application := New(Dependencies{
		Observer:   &fakeObserver{},
		Authorizer: &visibilityAuthorizer{t: t, tenantID: tenantID, visible: []inventory.InventoryID{inventoryID}},
		Tenants:    &fakeTenantRepository{exists: true},
		Inventories: &fakeInventoryRepository{items: []inventory.Inventory{
			inventoryItem(inventoryID.String(), tenantID.String(), "Home"),
		}},
		Search: &recordingAssetSearchRepository{items: []ports.AssetSearchResult{{
			Type:      search.ResultTypeAsset,
			TenantID:  tenantID,
			Inventory: inventoryItem(inventoryID.String(), tenantID.String(), "Home"),
			Asset:     item,
		}}},
		Attachments:      attachments,
		Audit:            &fakeAuditRepository{},
		DefaultPageLimit: 10,
		MaxPageLimit:     20,
	})

	result, err := application.SearchAssets(context.Background(), SearchAssetsInput{
		Principal: identity.Principal{ID: identity.PrincipalID("viewer")},
		TenantID:  tenantID,
		Query:     "water",
		Mode:      "exact",
	})
	if err != nil {
		t.Fatalf("search assets: %v", err)
	}
	if len(attachments.assets) != 1 || attachments.assets[0].AssetID != assetID || attachments.assets[0].InventoryID != inventoryID {
		t.Fatalf("expected primary image lookup for returned search asset, got %+v", attachments.assets)
	}
	if result.PrimaryPhotos[ref].ID != media.ID("attachment-one") {
		t.Fatalf("expected primary photo in search result, got %+v", result.PrimaryPhotos)
	}
}

func TestSearchAssetsIncludesAssignedTags(t *testing.T) {
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	assetID := asset.ID("asset-one")
	item := asset.Asset{
		ID:             assetID,
		TenantID:       asset.TenantID(tenantID.String()),
		InventoryID:    asset.InventoryID(inventoryID.String()),
		Kind:           asset.KindItem,
		Title:          asset.Title("Water bottle"),
		LifecycleState: asset.LifecycleStateActive,
	}
	tagKey, _ := assettag.NewKey("camping")
	tagName, _ := assettag.NewDisplayName("Camping")
	tagColor, _ := assettag.NewColor("#2f80ed")
	tag, _ := assettag.NewTag("tag-camping", assettag.TenantID(tenantID.String()), assettag.InventoryID(inventoryID.String()), tagKey, tagName, tagColor, time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC))
	application := New(Dependencies{
		Observer:   &fakeObserver{},
		Authorizer: &visibilityAuthorizer{t: t, tenantID: tenantID, visible: []inventory.InventoryID{inventoryID}},
		Tenants:    &fakeTenantRepository{exists: true},
		Inventories: &fakeInventoryRepository{items: []inventory.Inventory{
			inventoryItem(inventoryID.String(), tenantID.String(), "Home"),
		}},
		Search: &recordingAssetSearchRepository{items: []ports.AssetSearchResult{{
			Type:         search.ResultTypeAsset,
			TenantID:     tenantID,
			Inventory:    inventoryItem(inventoryID.String(), tenantID.String(), "Home"),
			Asset:        item,
			AssignedTags: []assettag.Tag{tag},
		}}},
		Audit:            &fakeAuditRepository{},
		DefaultPageLimit: 10,
		MaxPageLimit:     20,
	})

	result, err := application.SearchAssets(context.Background(), SearchAssetsInput{
		Principal: identity.Principal{ID: identity.PrincipalID("viewer")},
		TenantID:  tenantID,
		Query:     "water",
		Mode:      "exact",
	})
	if err != nil {
		t.Fatalf("search assets: %v", err)
	}
	if tags := result.Items[0].AssignedTags; len(tags) != 1 || tags[0].ID != tag.ID || tags[0].Color != tagColor {
		t.Fatalf("expected assigned tag in search result, got %+v", result.Items)
	}
}

func TestSearchAssetsWritesSafeReadAudit(t *testing.T) {
	t.Parallel()

	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	auditRepo := &fakeAuditRepository{}
	application := New(Dependencies{
		Observer:   &fakeObserver{},
		Authorizer: &visibilityAuthorizer{t: t, tenantID: tenantID, visible: []inventory.InventoryID{inventoryID}},
		Tenants:    &fakeTenantRepository{exists: true},
		Inventories: &fakeInventoryRepository{items: []inventory.Inventory{
			inventoryItem(inventoryID.String(), tenantID.String(), "Home"),
		}},
		Search:           &recordingAssetSearchRepository{},
		Audit:            auditRepo,
		IDs:              &fakeIDGenerator{ids: []string{"audit-search-one"}},
		DefaultPageLimit: 10,
		MaxPageLimit:     20,
	})

	_, err := application.SearchAssets(context.Background(), SearchAssetsInput{
		Principal:         identity.Principal{ID: identity.PrincipalID("viewer")},
		TenantID:          tenantID,
		InventoryIDs:      []inventory.InventoryID{inventoryID},
		Source:            audit.SourceAPI,
		RequestID:         "request-search-one",
		Query:             "water bottle bearer secret",
		Mode:              "exact",
		CustomAssetTypeID: "type-tools",
		LifecycleState:    "active",
		CheckoutState:     "available",
		Limit:             7,
	})
	if err != nil {
		t.Fatalf("search assets: %v", err)
	}

	record, ok := auditRepo.recordForAction(audit.ActionAssetSearched)
	if !ok {
		t.Fatalf("expected search read audit record, got %+v", auditRepo.items)
	}
	if record.Source != audit.SourceAPI || record.RequestID != "request-search-one" || record.TargetType != audit.TargetInventory || record.TargetID != inventoryID.String() || record.InventoryID.String() != inventoryID.String() {
		t.Fatalf("unexpected search audit record: %+v", record)
	}
	expectedMetadata := map[string]string{
		"scope":                    "inventory",
		"limit":                    "7",
		"mode":                     "exact",
		"lifecycle":                "active",
		"checkout":                 "available",
		"custom_asset_type_filter": "true",
		"authorized_inventories":   "1",
		"result_count":             "0",
	}
	for key, value := range expectedMetadata {
		if record.Metadata[key] != value {
			t.Fatalf("expected audit metadata %s=%q, got %+v", key, value, record.Metadata)
		}
	}
	searchAuditMustNotContain(t, record, "water", "bottle", "bearer", "secret", "type-tools")
}

func TestRealtimeVoiceSearchToolWritesConversationReadAudit(t *testing.T) {
	t.Parallel()

	tenantID := tenant.ID("tenant-home")
	inventoryID := inventory.InventoryID("inventory-home")
	auditRepo := &fakeAuditRepository{}
	application := New(Dependencies{
		Observer:   &fakeObserver{},
		Authorizer: &visibilityAuthorizer{t: t, tenantID: tenantID, visible: []inventory.InventoryID{inventoryID}},
		Tenants:    &fakeTenantRepository{exists: true},
		Inventories: &fakeInventoryRepository{items: []inventory.Inventory{
			inventoryItem(inventoryID.String(), tenantID.String(), "Home"),
		}},
		Search:           &recordingAssetSearchRepository{},
		Audit:            auditRepo,
		IDs:              &fakeIDGenerator{ids: []string{"audit-voice-search"}},
		DefaultPageLimit: 10,
		MaxPageLimit:     20,
	})

	_, err := application.executeRealtimeVoiceTool(context.Background(), RealtimeVoiceSession{
		TenantID:    tenantID,
		InventoryID: inventoryID,
		Principal:   identity.Principal{ID: identity.PrincipalID("speaker")},
	}, ports.AgentToolCall{
		ID:   "tool-search",
		Name: RealtimeVoiceToolSearchAuthorizedAssets,
		Arguments: map[string]any{
			"query": "passport prompt token",
			"limit": float64(3),
		},
	}, map[string]struct{}{})
	if err != nil {
		t.Fatalf("execute realtime voice search tool: %v", err)
	}

	record, ok := auditRepo.recordForAction(audit.ActionAssetSearched)
	if !ok {
		t.Fatalf("expected voice search read audit record, got %+v", auditRepo.items)
	}
	if record.Source != audit.SourceConversation || record.TargetType != audit.TargetInventory || record.TargetID != inventoryID.String() {
		t.Fatalf("unexpected voice search audit record: %+v", record)
	}
	searchAuditMustNotContain(t, record, "passport", "prompt", "token", "tool-search", RealtimeVoiceToolSearchAuthorizedAssets)
}

func searchAuditMustNotContain(t *testing.T, record audit.Record, unsafe ...string) {
	t.Helper()

	combined := record.TargetID + " " + record.RequestID
	for key, value := range record.Metadata {
		combined += " " + key + " " + value
	}
	for _, value := range unsafe {
		if strings.Contains(combined, value) {
			t.Fatalf("search audit leaked %q in %+v", value, record)
		}
	}
}

type visibilityAuthorizer struct {
	t                *testing.T
	tenantID         tenant.ID
	visible          []inventory.InventoryID
	candidates       []inventory.InventoryID
	visibilityCalled bool
}

func (v *visibilityAuthorizer) CheckTenant(context.Context, identity.Principal, ports.TenantPermission, tenant.ID) error {
	return nil
}

func (v *visibilityAuthorizer) CheckInventory(context.Context, identity.Principal, ports.InventoryPermission, inventory.InventoryID) error {
	v.t.Fatalf("search must use ListViewableInventoryIDs instead of per-inventory checks")
	return nil
}

func (v *visibilityAuthorizer) ListViewableInventoryIDs(_ context.Context, _ identity.Principal, tenantID tenant.ID, candidates []inventory.InventoryID) ([]inventory.InventoryID, error) {
	if tenantID != v.tenantID {
		v.t.Fatalf("expected tenant %q, got %q", v.tenantID, tenantID)
	}
	v.visibilityCalled = true
	v.candidates = append([]inventory.InventoryID{}, candidates...)
	return append([]inventory.InventoryID{}, v.visible...), nil
}

func (v *visibilityAuthorizer) GrantTenantOwner(context.Context, identity.Principal, tenant.ID) error {
	return nil
}

func (v *visibilityAuthorizer) GrantInventoryOwner(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (v *visibilityAuthorizer) GrantInventoryViewer(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (v *visibilityAuthorizer) GrantInventoryEditor(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (v *visibilityAuthorizer) RevokeInventoryViewer(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (v *visibilityAuthorizer) RevokeInventoryEditor(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

type recordingAssetSearchRepository struct {
	inventoryIDs []inventory.InventoryID
	items        []ports.AssetSearchResult
}

func (r *recordingAssetSearchRepository) SearchAssets(_ context.Context, _ tenant.ID, inventoryIDs []inventory.InventoryID, _ ports.AssetSearchPageRequest) ([]ports.AssetSearchResult, error) {
	r.inventoryIDs = append([]inventory.InventoryID{}, inventoryIDs...)
	return r.items, nil
}

type searchAttachmentRepository struct {
	assets        []ports.AttachmentAssetReference
	primaryImages map[ports.AttachmentAssetReference]media.Attachment
}

func (r *searchAttachmentRepository) AttachmentByID(context.Context, tenant.ID, inventory.InventoryID, asset.ID, media.ID) (media.Attachment, bool, error) {
	return media.Attachment{}, false, nil
}

func (r *searchAttachmentRepository) ListAttachmentsByAsset(context.Context, tenant.ID, inventory.InventoryID, asset.ID, ports.AttachmentListPageRequest) ([]media.Attachment, error) {
	return nil, nil
}

func (r *searchAttachmentRepository) FirstImageAttachmentsByAssets(_ context.Context, _ tenant.ID, assets []ports.AttachmentAssetReference) (map[ports.AttachmentAssetReference]media.Attachment, error) {
	r.assets = append([]ports.AttachmentAssetReference{}, assets...)
	return r.primaryImages, nil
}
