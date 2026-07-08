package gormstore

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
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
	storageID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FC1")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveTenant(t, ctx, store, otherTenantID, "Other Home")
	saveInventory(t, ctx, store, toolsID.String(), tenantID, "Tools")
	saveInventory(t, ctx, store, medicineID.String(), tenantID, "Medicine")
	saveInventory(t, ctx, store, storageID.String(), tenantID, "Storage")
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
	drillBits := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FD1", tenantID.String(), toolsID.String(), asset.KindItem, "")
	drillBitsTitle, _ := asset.NewTitle("Drill Bits")
	drillBits.Title = drillBitsTitle
	if err := createAsset(t, ctx, store, drillBits); err != nil {
		t.Fatalf("create drill bits: %v", err)
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
	storedBin := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FC2", tenantID.String(), storageID.String(), asset.KindItem, "")
	storedBinTitle, _ := asset.NewTitle("Plastic bin")
	storedBin.Title = storedBinTitle
	if err := createAsset(t, ctx, store, storedBin); err != nil {
		t.Fatalf("create stored bin: %v", err)
	}

	saveSearchAttachment(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FB4", drill, "warranty-card.png", media.ContentTypePNG)

	workshopTag := assetTag(t, "01ARZ3NDEKTSV4RRFFQ69G5FB5", tenantID, toolsID, "shop-tools", "Workshop")
	if err := store.CreateAssetTag(ctx, workshopTag, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FB6", tenantID, toolsID, audit.ActionAssetTagCreated)); err != nil {
		t.Fatalf("create search tag: %v", err)
	}
	campingTag := assetTag(t, "01ARZ3NDEKTSV4RRFFQ69G5FD2", tenantID, toolsID, "camping", "Camping")
	if err := store.CreateAssetTag(ctx, campingTag, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FD3", tenantID, toolsID, audit.ActionAssetTagCreated)); err != nil {
		t.Fatalf("create camping search tag: %v", err)
	}
	if err := store.SetAssetTags(ctx, tenantID, toolsID, drill.ID, []assettag.ID{workshopTag.ID, campingTag.ID}, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FB7", tenantID, toolsID, audit.ActionAssetUpdated)); err != nil {
		t.Fatalf("assign search tag: %v", err)
	}
	if err := store.SetAssetTags(ctx, tenantID, toolsID, drillBits.ID, []assettag.ID{campingTag.ID}, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FD4", tenantID, toolsID, audit.ActionAssetUpdated)); err != nil {
		t.Fatalf("assign camping search tag: %v", err)
	}
	otherTenantTag := assetTag(t, "01ARZ3NDEKTSV4RRFFQ69G5FC3", otherTenantID, otherInventoryID, "other-secret", "Other Secret")
	if err := store.CreateAssetTag(ctx, otherTenantTag, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FC4", otherTenantID, otherInventoryID, audit.ActionAssetTagCreated)); err != nil {
		t.Fatalf("create other tenant search tag: %v", err)
	}
	if err := store.SetAssetTags(ctx, otherTenantID, otherInventoryID, otherDrill.ID, []assettag.ID{otherTenantTag.ID}, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FC5", otherTenantID, otherInventoryID, audit.ActionAssetUpdated)); err != nil {
		t.Fatalf("assign other tenant search tag: %v", err)
	}
	storageTag := assetTag(t, "01ARZ3NDEKTSV4RRFFQ69G5FC6", tenantID, storageID, "storage-only", "Storage Only")
	if err := store.CreateAssetTag(ctx, storageTag, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FC7", tenantID, storageID, audit.ActionAssetTagCreated)); err != nil {
		t.Fatalf("create excluded inventory search tag: %v", err)
	}
	if err := store.SetAssetTags(ctx, tenantID, storageID, storedBin.ID, []assettag.ID{storageTag.ID}, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FC8", tenantID, storageID, audit.ActionAssetUpdated)); err != nil {
		t.Fatalf("assign excluded inventory search tag: %v", err)
	}
	archivedTag := assetTag(t, "01ARZ3NDEKTSV4RRFFQ69G5FBA", tenantID, medicineID, "archived-tag", "Retired")
	if err := store.CreateAssetTag(ctx, archivedTag, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FBB", tenantID, medicineID, audit.ActionAssetTagCreated)); err != nil {
		t.Fatalf("create archived search tag: %v", err)
	}
	if err := store.SetAssetTags(ctx, tenantID, medicineID, aspirin.ID, []assettag.ID{archivedTag.ID}, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FBC", tenantID, medicineID, audit.ActionAssetUpdated)); err != nil {
		t.Fatalf("assign archived search tag before archival: %v", err)
	}
	archivedTag.LifecycleState = assettag.LifecycleStateArchived
	archivedTag.UpdatedAt = time.Now().UTC()
	if err := store.UpdateAssetTagLifecycle(ctx, archivedTag, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FBD", tenantID, medicineID, audit.ActionAssetTagArchived)); err != nil {
		t.Fatalf("archive search tag: %v", err)
	}

	attachmentResults := searchPersistedAssets(t, ctx, store, tenantID, []inventory.InventoryID{toolsID, medicineID}, "warranty", search.ModeFuzzy, ports.AssetLifecycleFilterActive, "", "", 10)
	if len(attachmentResults) != 1 || attachmentResults[0].Asset.ID != drill.ID || attachmentResults[0].Matches[0].Field != search.MatchFieldAttachmentFileName {
		t.Fatalf("expected attachment file name search to find drill, got %+v", attachmentResults)
	}

	tagResults := searchPersistedAssets(t, ctx, store, tenantID, []inventory.InventoryID{toolsID, medicineID}, "workshop", search.ModeExact, ports.AssetLifecycleFilterActive, "", "", 10)
	if len(tagResults) != 1 || tagResults[0].Asset.ID != drill.ID || tagResults[0].Matches[0].Field != search.MatchFieldTagDisplayName {
		t.Fatalf("expected exact tag display-name search to find drill, got %+v", tagResults)
	}

	tagKeyResults := searchPersistedAssets(t, ctx, store, tenantID, []inventory.InventoryID{toolsID, medicineID}, "shop-tools", search.ModeExact, ports.AssetLifecycleFilterActive, "", "", 10)
	if len(tagKeyResults) != 1 || tagKeyResults[0].Asset.ID != drill.ID || tagKeyResults[0].Matches[0].Field != search.MatchFieldTagKey {
		t.Fatalf("expected exact tag key search to find drill, got %+v", tagKeyResults)
	}

	tagOnlyResults := searchPersistedAssetsWithTagIDs(t, ctx, store, tenantID, []inventory.InventoryID{toolsID, medicineID}, "", search.ModeFuzzy, []assettag.ID{campingTag.ID}, ports.AssetLifecycleFilterActive, ports.AssetCheckoutStateFilterAny, "", "", 10)
	if !searchResultsContainAsset(tagOnlyResults, drill.ID) || !searchResultsContainAsset(tagOnlyResults, drillBits.ID) {
		t.Fatalf("expected tag-only facet to browse camping assets, got %+v", tagOnlyResults)
	}

	multiTagResults := searchPersistedAssetsWithTagIDs(t, ctx, store, tenantID, []inventory.InventoryID{toolsID, medicineID}, "Drill", search.ModeFuzzy, []assettag.ID{workshopTag.ID, campingTag.ID}, ports.AssetLifecycleFilterActive, ports.AssetCheckoutStateFilterAny, "", "", 10)
	if len(multiTagResults) != 1 || multiTagResults[0].Asset.ID != drill.ID {
		t.Fatalf("expected multi-tag facet to require all selected tags, got %+v", multiTagResults)
	}

	excludedInventoryTagIDResults := searchPersistedAssetsWithTagIDs(t, ctx, store, tenantID, []inventory.InventoryID{toolsID, medicineID}, "", search.ModeFuzzy, []assettag.ID{storageTag.ID}, ports.AssetLifecycleFilterActive, ports.AssetCheckoutStateFilterAny, "", "", 10)
	if len(excludedInventoryTagIDResults) != 0 {
		t.Fatalf("expected excluded inventory tag ID facet not to leak storage asset, got %+v", excludedInventoryTagIDResults)
	}

	otherTenantTagIDResults := searchPersistedAssetsWithTagIDs(t, ctx, store, tenantID, []inventory.InventoryID{toolsID, medicineID}, "", search.ModeFuzzy, []assettag.ID{otherTenantTag.ID}, ports.AssetLifecycleFilterActive, ports.AssetCheckoutStateFilterAny, "", "", 10)
	if len(otherTenantTagIDResults) != 0 {
		t.Fatalf("expected other tenant tag ID facet not to leak other tenant asset, got %+v", otherTenantTagIDResults)
	}

	archivedTagIDResults := searchPersistedAssetsWithTagIDs(t, ctx, store, tenantID, []inventory.InventoryID{toolsID, medicineID}, "", search.ModeFuzzy, []assettag.ID{archivedTag.ID}, ports.AssetLifecycleFilterActive, ports.AssetCheckoutStateFilterAny, "", "", 10)
	if len(archivedTagIDResults) != 0 {
		t.Fatalf("expected archived tag ID facet not to satisfy assigned asset, got %+v", archivedTagIDResults)
	}

	archivedTagResults := searchPersistedAssets(t, ctx, store, tenantID, []inventory.InventoryID{toolsID, medicineID}, "Retired", search.ModeExact, ports.AssetLifecycleFilterActive, "", "", 10)
	if len(archivedTagResults) != 0 {
		t.Fatalf("expected archived tag not to contribute to search, got %+v", archivedTagResults)
	}

	otherTenantTagResults := searchPersistedAssets(t, ctx, store, tenantID, []inventory.InventoryID{toolsID, medicineID}, "Other Secret", search.ModeExact, ports.AssetLifecycleFilterActive, "", "", 10)
	if len(otherTenantTagResults) != 0 {
		t.Fatalf("expected tag search not to leak other tenant assignments, got %+v", otherTenantTagResults)
	}

	excludedInventoryTagResults := searchPersistedAssets(t, ctx, store, tenantID, []inventory.InventoryID{toolsID, medicineID}, "Storage Only", search.ModeExact, ports.AssetLifecycleFilterActive, "", "", 10)
	if len(excludedInventoryTagResults) != 0 {
		t.Fatalf("expected tag search not to leak excluded inventory assignments, got %+v", excludedInventoryTagResults)
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

	checkout := checkoutRecord("01ARZ3NDEKTSV4RRFFQ69G5FC9", drill, time.Now().UTC())
	if err := store.CheckOutAsset(ctx, checkout, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FDA", tenantID, toolsID, audit.ActionAssetCheckedOut), nil); err != nil {
		t.Fatalf("check out drill: %v", err)
	}
	availableTagResults := searchPersistedAssetsWithCheckoutFilter(t, ctx, store, tenantID, []inventory.InventoryID{toolsID, medicineID}, "workshop", search.ModeExact, ports.AssetLifecycleFilterActive, ports.AssetCheckoutStateFilterAvailable, "", "", 10)
	if len(availableTagResults) != 0 {
		t.Fatalf("expected available checkout filter to hide checked-out tag match, got %+v", availableTagResults)
	}
	checkedOutTagResults := searchPersistedAssetsWithCheckoutFilter(t, ctx, store, tenantID, []inventory.InventoryID{toolsID, medicineID}, "workshop", search.ModeExact, ports.AssetLifecycleFilterActive, ports.AssetCheckoutStateFilterCheckedOut, "", "", 10)
	if len(checkedOutTagResults) != 1 || checkedOutTagResults[0].Asset.ID != drill.ID || checkedOutTagResults[0].Matches[0].Field != search.MatchFieldTagDisplayName {
		t.Fatalf("expected checked-out filter to include tag match, got %+v", checkedOutTagResults)
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
	return searchPersistedAssetsWithCheckoutFilter(t, ctx, store, tenantID, inventoryIDs, queryValue, mode, lifecycle, ports.AssetCheckoutStateFilterAny, customAssetTypeID, afterResultKey, limit)
}

func searchPersistedAssetsWithCheckoutFilter(t *testing.T, ctx context.Context, store Store, tenantID tenant.ID, inventoryIDs []inventory.InventoryID, queryValue string, mode search.Mode, lifecycle ports.AssetLifecycleFilter, checkoutFilter ports.AssetCheckoutStateFilter, customAssetTypeID string, afterResultKey string, limit int) []ports.AssetSearchResult {
	t.Helper()
	return searchPersistedAssetsWithTagIDs(t, ctx, store, tenantID, inventoryIDs, queryValue, mode, nil, lifecycle, checkoutFilter, customAssetTypeID, afterResultKey, limit)
}

func searchPersistedAssetsWithTagIDs(t *testing.T, ctx context.Context, store Store, tenantID tenant.ID, inventoryIDs []inventory.InventoryID, queryValue string, mode search.Mode, tagIDs []assettag.ID, lifecycle ports.AssetLifecycleFilter, checkoutFilter ports.AssetCheckoutStateFilter, customAssetTypeID string, afterResultKey string, limit int) []ports.AssetSearchResult {
	t.Helper()

	query, ok := search.NewQuery(queryValue)
	if !ok {
		t.Fatalf("expected valid search query")
	}
	results, err := store.SearchAssets(ctx, tenantID, inventoryIDs, ports.AssetSearchPageRequest{
		Query:             query,
		Mode:              mode,
		TagIDs:            tagIDs,
		CustomAssetTypeID: asset.CustomAssetTypeID(customAssetTypeID),
		AfterResultKey:    afterResultKey,
		Limit:             limit,
		LifecycleFilter:   lifecycle,
		CheckoutFilter:    checkoutFilter,
	})
	if err != nil {
		t.Fatalf("search persisted assets: %v", err)
	}
	return results
}

func searchResultsContainAsset(results []ports.AssetSearchResult, assetID asset.ID) bool {
	for _, result := range results {
		if result.Asset.ID == assetID {
			return true
		}
	}
	return false
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
