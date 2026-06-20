package gormstore

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestStoreSearchAssetsMatchesPersistedMetadata(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)

	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	toolsID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	medicineID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAX")
	otherTenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAY")
	otherInventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAZ")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveTenant(t, ctx, store, otherTenantID, "Other Home")
	saveInventory(t, ctx, store, toolsID.String(), tenantID, "Tools")
	saveInventory(t, ctx, store, medicineID.String(), tenantID, "Medicine")
	saveInventory(t, ctx, store, otherInventoryID.String(), otherTenantID, "Other")

	medicineType := customAssetType(t, "01ARZ3NDEKTSV4RRFFQ69G5FB0", tenantID.String(), medicineID.String(), customfield.ScopeInventory, "medicine")
	if err := saveCustomAssetType(t, ctx, store, medicineType); err != nil {
		t.Fatalf("save custom asset type: %v", err)
	}

	drill := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FB1", tenantID.String(), toolsID.String(), asset.KindItem, "")
	drillTitle, _ := asset.NewTitle("Cordless Drill")
	drill.Title = drillTitle
	drill.Description = asset.NewDescription("Driver kit")
	drillFields, ok := asset.NewCustomFields(map[string]any{"serial": "bag-42"})
	if !ok {
		t.Fatalf("expected valid drill custom fields")
	}
	drill.CustomFields = drillFields
	if err := createAsset(t, ctx, store, drill); err != nil {
		t.Fatalf("create drill: %v", err)
	}

	aspirin := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FB2", tenantID.String(), medicineID.String(), asset.KindItem, "")
	aspirinTitle, _ := asset.NewTitle("Aspirin")
	aspirin.Title = aspirinTitle
	aspirin.Description = asset.NewDescription("Pain relief tablets")
	aspirin.CustomAssetTypeID = asset.CustomAssetTypeID(medicineType.ID.String())
	aspirinFields, ok := asset.NewCustomFields(map[string]any{"expires-on": "2027-01-01"})
	if !ok {
		t.Fatalf("expected valid aspirin custom fields")
	}
	aspirin.CustomFields = aspirinFields
	if err := createAsset(t, ctx, store, aspirin); err != nil {
		t.Fatalf("create aspirin: %v", err)
	}

	otherDrill := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FB3", otherTenantID.String(), otherInventoryID.String(), asset.KindItem, "")
	otherDrill.Title = drillTitle
	if err := createAsset(t, ctx, store, otherDrill); err != nil {
		t.Fatalf("create other drill: %v", err)
	}

	saveSearchAttachment(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FB4", drill, "warranty-card.png", media.ContentTypePNG)

	attachmentResults := searchPersistedAssets(t, ctx, store, tenantID, []inventory.InventoryID{toolsID, medicineID}, "warranty", search.ModeFuzzy, ports.AssetLifecycleFilterActive, "", "", 10)
	if len(attachmentResults) != 1 || attachmentResults[0].Asset.ID != drill.ID || attachmentResults[0].Matches[0].Field != search.MatchFieldAttachmentFileName {
		t.Fatalf("expected attachment file name search to find drill, got %+v", attachmentResults)
	}

	customFieldResults := searchPersistedAssets(t, ctx, store, tenantID, []inventory.InventoryID{toolsID, medicineID}, "bag-42", search.ModeExact, ports.AssetLifecycleFilterActive, "", "", 10)
	if len(customFieldResults) != 1 || customFieldResults[0].Asset.ID != drill.ID || customFieldResults[0].Matches[0].Field != search.MatchFieldCustomField {
		t.Fatalf("expected exact custom field search to find drill, got %+v", customFieldResults)
	}

	typeResults := searchPersistedAssets(t, ctx, store, tenantID, []inventory.InventoryID{toolsID, medicineID}, "medicine", search.ModeExact, ports.AssetLifecycleFilterActive, medicineType.ID.String(), "", 10)
	if len(typeResults) != 1 || typeResults[0].Asset.ID != aspirin.ID || typeResults[0].Matches[0].Field != search.MatchFieldCustomAssetTypeKey {
		t.Fatalf("expected exact custom asset type search to find aspirin, got %+v", typeResults)
	}

	scopedResults := searchPersistedAssets(t, ctx, store, tenantID, []inventory.InventoryID{medicineID}, "Cordless", search.ModeFuzzy, ports.AssetLifecycleFilterActive, "", "", 10)
	if len(scopedResults) != 0 {
		t.Fatalf("expected inventory scoping to hide tool inventory result, got %+v", scopedResults)
	}

	firstPage := searchPersistedAssets(t, ctx, store, tenantID, []inventory.InventoryID{toolsID, medicineID}, "i", search.ModeFuzzy, ports.AssetLifecycleFilterActive, "", "", 1)
	if len(firstPage) != 1 {
		t.Fatalf("expected first paginated result, got %+v", firstPage)
	}
	secondPage := searchPersistedAssets(t, ctx, store, tenantID, []inventory.InventoryID{toolsID, medicineID}, "i", search.ModeFuzzy, ports.AssetLifecycleFilterActive, "", firstPage[0].CursorKey(), 1)
	if len(secondPage) != 1 || secondPage[0].CursorKey() <= firstPage[0].CursorKey() {
		t.Fatalf("expected second page to advance, first=%+v second=%+v", firstPage, secondPage)
	}

	drill.LifecycleState = asset.LifecycleStateArchived
	if err := store.UpdateAssetLifecycle(ctx, drill, auditRecord(t, auditIDWithSuffix(drill.ID.String(), "A"), tenantID, toolsID, audit.ActionAssetArchived), nil); err != nil {
		t.Fatalf("archive drill: %v", err)
	}
	activeAfterArchive := searchPersistedAssets(t, ctx, store, tenantID, []inventory.InventoryID{toolsID, medicineID}, "Cordless", search.ModeFuzzy, ports.AssetLifecycleFilterActive, "", "", 10)
	if len(activeAfterArchive) != 0 {
		t.Fatalf("expected active search to hide archived drill, got %+v", activeAfterArchive)
	}
	archivedResults := searchPersistedAssets(t, ctx, store, tenantID, []inventory.InventoryID{toolsID, medicineID}, "Cordless", search.ModeFuzzy, ports.AssetLifecycleFilterArchived, "", "", 10)
	if len(archivedResults) != 1 || archivedResults[0].Asset.ID != drill.ID {
		t.Fatalf("expected archived search to find drill, got %+v", archivedResults)
	}
}

func searchPersistedAssets(t *testing.T, ctx context.Context, store Store, tenantID tenant.ID, inventoryIDs []inventory.InventoryID, queryValue string, mode search.Mode, lifecycle ports.AssetLifecycleFilter, customAssetTypeID string, afterResultKey string, limit int) []ports.AssetSearchResult {
	t.Helper()

	query, ok := search.NewQuery(queryValue)
	if !ok {
		t.Fatalf("expected valid search query")
	}
	results, err := store.SearchAssets(ctx, tenantID, inventoryIDs, ports.AssetSearchPageRequest{
		Query:             query,
		Mode:              mode,
		CustomAssetTypeID: asset.CustomAssetTypeID(customAssetTypeID),
		AfterResultKey:    afterResultKey,
		Limit:             limit,
		LifecycleFilter:   lifecycle,
	})
	if err != nil {
		t.Fatalf("search persisted assets: %v", err)
	}
	return results
}

func saveSearchAttachment(t *testing.T, ctx context.Context, store Store, id string, item asset.Asset, fileNameValue string, contentType media.ContentType) {
	t.Helper()
	saveSearchAttachmentWithAuditID(t, ctx, store, id, auditIDWithSuffix(id, "M"), item, fileNameValue, contentType)
}

func saveSearchAttachmentWithAuditID(t *testing.T, ctx context.Context, store Store, id string, auditID string, item asset.Asset, fileNameValue string, contentType media.ContentType) {
	t.Helper()

	attachmentID, ok := media.NewID(id)
	if !ok {
		t.Fatalf("expected valid attachment id")
	}
	storageKey, ok := media.NewStorageKey("test/" + id)
	if !ok {
		t.Fatalf("expected valid storage key")
	}
	fileName, ok := media.NewFileName(fileNameValue)
	if !ok {
		t.Fatalf("expected valid file name")
	}
	hash, ok := media.NewSHA256("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	if !ok {
		t.Fatalf("expected valid sha256")
	}
	attachment, ok := media.NewAttachment(
		attachmentID,
		media.TenantID(item.TenantID.String()),
		media.InventoryID(item.InventoryID.String()),
		media.AssetID(item.ID.String()),
		storageKey,
		fileName,
		contentType,
		12,
		hash,
		time.Now(),
	)
	if !ok {
		t.Fatalf("expected valid attachment")
	}
	if err := store.SaveAttachment(ctx, attachment, auditRecord(t, auditID, tenant.ID(item.TenantID.String()), inventory.InventoryID(item.InventoryID.String()), audit.ActionAttachmentCreated)); err != nil {
		t.Fatalf("save attachment: %v", err)
	}
}
