package app

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
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

type SearchAssetsInput struct {
	Principal         identity.Principal
	TenantID          tenant.ID
	InventoryIDs      []inventory.InventoryID
	Source            audit.Source
	RequestID         string
	Query             string
	Mode              string
	TagIDs            []assettag.ID
	CustomAssetTypeID string
	LifecycleState    string
	CheckoutState     string
	Limit             int
	Cursor            string
}

type SearchAssetsResult struct {
	Items         []ports.AssetSearchResult
	PrimaryPhotos map[ports.AttachmentAssetReference]media.Attachment
	Limit         int
	NextCursor    *string
	HasMore       bool
}

func (a App) SearchAssets(ctx context.Context, input SearchAssetsInput) (SearchAssetsResult, error) {
	exists, err := a.tenants.TenantExists(ctx, input.TenantID)
	if err != nil {
		return SearchAssetsResult{}, err
	}
	if !exists {
		return SearchAssetsResult{}, ErrNotFound
	}

	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionView, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return SearchAssetsResult{}, err
	}

	query, ok := search.NewQuery(input.Query)
	if !ok {
		return SearchAssetsResult{}, ErrInvalidInput
	}
	tagIDs := normalizeSearchTagIDs(input.TagIDs)
	if query.String() == "" && len(tagIDs) == 0 {
		return SearchAssetsResult{}, ErrInvalidInput
	}
	mode, ok := search.NewMode(input.Mode)
	if !ok {
		return SearchAssetsResult{}, ErrInvalidInput
	}
	lifecycleFilter, err := assetapp.LifecycleFilter(input.LifecycleState)
	if err != nil {
		return SearchAssetsResult{}, ErrInvalidInput
	}
	checkoutFilter, err := searchCheckoutStateFilter(input.CheckoutState)
	if err != nil {
		return SearchAssetsResult{}, ErrInvalidInput
	}
	customAssetTypeID, err := parseSearchCustomAssetTypeID(input.CustomAssetTypeID)
	if err != nil {
		return SearchAssetsResult{}, ErrInvalidInput
	}
	limit := pageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	cursorScope := searchCursorScope(input.TenantID, input.InventoryIDs, query, mode, tagIDs, customAssetTypeID, lifecycleFilter, checkoutFilter)
	afterResultKey, err := decodePageCursor("search.assets", cursorScope, input.Cursor)
	if err != nil {
		return SearchAssetsResult{}, ErrInvalidInput
	}

	candidateInventoryIDs, err := a.inventoryIDsForTenant(ctx, input.TenantID)
	if err != nil {
		return SearchAssetsResult{}, err
	}
	if len(input.InventoryIDs) > 0 {
		candidateInventoryIDs = intersectInventoryCandidates(candidateInventoryIDs, input.InventoryIDs)
	}
	inventoryIDs, err := a.authorizer.ListViewableInventoryIDs(ctx, input.Principal, input.TenantID, candidateInventoryIDs)
	if err != nil {
		return SearchAssetsResult{}, err
	}
	if len(inventoryIDs) == 0 {
		if err := a.saveSearchAssetsReadAudit(ctx, input, inventoryIDs, limit, mode.String(), string(lifecycleFilter), string(checkoutFilter), customAssetTypeID.String(), 0); err != nil {
			return SearchAssetsResult{}, err
		}
		return SearchAssetsResult{Items: []ports.AssetSearchResult{}, Limit: limit}, nil
	}
	if a.search == nil {
		return SearchAssetsResult{}, ErrInvalidInput
	}

	items, err := a.search.SearchAssets(ctx, input.TenantID, inventoryIDs, ports.AssetSearchPageRequest{
		Query:             query,
		Mode:              mode,
		TagIDs:            tagIDs,
		CustomAssetTypeID: customAssetTypeID,
		AfterResultKey:    afterResultKey,
		Limit:             limit + 1,
		LifecycleFilter:   lifecycleFilter,
		CheckoutFilter:    checkoutFilter,
	})
	if err != nil {
		return SearchAssetsResult{}, err
	}

	hasMore := len(items) > limit
	var nextCursor *string
	if hasMore {
		items = items[:limit]
		nextCursor = encodePageCursor("search.assets", cursorScope, items[len(items)-1].CursorKey())
	}
	primaryPhotos, err := a.primaryImageAttachmentsForSearchResults(ctx, input.TenantID, items)
	if err != nil {
		return SearchAssetsResult{}, err
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAssetsSearched,
		Message: "assets searched",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"principal_id":  input.Principal.ID.String(),
			"limit":         strconv.Itoa(limit),
			"mode":          mode.String(),
			"tag_filters":   strconv.Itoa(len(tagIDs)),
			"inventory_ids": strconv.Itoa(len(inventoryIDs)),
		},
	})
	if err := a.saveSearchAssetsReadAudit(ctx, input, inventoryIDs, limit, mode.String(), string(lifecycleFilter), string(checkoutFilter), customAssetTypeID.String(), len(items)); err != nil {
		return SearchAssetsResult{}, err
	}
	a.warmPrimarySmallThumbnails(ctx, primaryPhotosForSearchResults(items, primaryPhotos))

	return SearchAssetsResult{
		Items:         items,
		PrimaryPhotos: primaryPhotos,
		Limit:         limit,
		NextCursor:    nextCursor,
		HasMore:       hasMore,
	}, nil
}

func (a App) saveSearchAssetsReadAudit(ctx context.Context, input SearchAssetsInput, authorizedInventoryIDs []inventory.InventoryID, limit int, mode string, lifecycle string, checkout string, customAssetTypeID string, resultCount int) error {
	if input.Source.String() == "" {
		return nil
	}
	targetType := audit.TargetTenant
	targetID := input.TenantID.String()
	inventoryID := inventory.InventoryID("")
	scope := "tenant"
	if len(input.InventoryIDs) == 1 {
		targetType = audit.TargetInventory
		targetID = input.InventoryIDs[0].String()
		inventoryID = input.InventoryIDs[0]
		scope = "inventory"
	}
	return appsupport.SaveReadAuditRecord(ctx, a.audit, a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: inventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionAssetSearched,
		TargetType:  targetType,
		TargetID:    targetID,
		Metadata: map[string]string{
			"scope":                    scope,
			"limit":                    strconv.Itoa(limit),
			"mode":                     mode,
			"lifecycle":                lifecycle,
			"checkout":                 checkout,
			"custom_asset_type_filter": strconv.FormatBool(strings.TrimSpace(customAssetTypeID) != ""),
			"authorized_inventories":   strconv.Itoa(len(authorizedInventoryIDs)),
			"result_count":             strconv.Itoa(resultCount),
		},
	})
}

func primaryPhotosForSearchResults(items []ports.AssetSearchResult, primaryPhotos map[ports.AttachmentAssetReference]media.Attachment) []media.Attachment {
	photos := make([]media.Attachment, 0, len(items))
	for _, item := range items {
		ref := ports.AttachmentAssetReference{
			InventoryID: inventory.InventoryID(item.Asset.InventoryID.String()),
			AssetID:     item.Asset.ID,
		}
		if photo, ok := primaryPhotos[ref]; ok {
			photos = append(photos, photo)
		}
	}
	return photos
}

func (a App) primaryImageAttachmentsForSearchResults(ctx context.Context, tenantID tenant.ID, items []ports.AssetSearchResult) (map[ports.AttachmentAssetReference]media.Attachment, error) {
	if a.attachments == nil || len(items) == 0 {
		return nil, nil
	}
	assetRefs := make([]ports.AttachmentAssetReference, 0, len(items))
	for _, item := range items {
		assetRefs = append(assetRefs, ports.AttachmentAssetReference{
			InventoryID: inventory.InventoryID(item.Asset.InventoryID.String()),
			AssetID:     item.Asset.ID,
		})
	}
	return a.attachments.FirstImageAttachmentsByAssets(ctx, tenantID, assetRefs)
}

func intersectInventoryCandidates(tenantCandidates []inventory.InventoryID, requested []inventory.InventoryID) []inventory.InventoryID {
	allowed := map[inventory.InventoryID]struct{}{}
	for _, id := range tenantCandidates {
		allowed[id] = struct{}{}
	}
	result := []inventory.InventoryID{}
	seen := map[inventory.InventoryID]struct{}{}
	for _, id := range requested {
		if _, ok := allowed[id]; !ok {
			continue
		}
		if _, duplicate := seen[id]; duplicate {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func (a App) inventoryIDsForTenant(ctx context.Context, tenantID tenant.ID) ([]inventory.InventoryID, error) {
	items, err := a.inventories.ListInventoriesByTenant(ctx, inventory.TenantID(tenantID.String()), ports.InventoryListPageRequest{})
	if err != nil {
		return nil, err
	}

	ids := make([]inventory.InventoryID, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids, nil
}

func parseSearchCustomAssetTypeID(raw string) (asset.CustomAssetTypeID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	id, ok := asset.NewCustomAssetTypeID(raw)
	if !ok {
		return "", ErrInvalidInput
	}
	return id, nil
}

func searchCursorScope(
	tenantID tenant.ID,
	inventoryIDs []inventory.InventoryID,
	query search.Query,
	mode search.Mode,
	tagIDs []assettag.ID,
	customAssetTypeID asset.CustomAssetTypeID,
	lifecycleFilter ports.AssetLifecycleFilter,
	checkoutFilter ports.AssetCheckoutStateFilter,
) string {
	return strings.Join([]string{
		tenantID.String(),
		searchInventoryCursorScope(inventoryIDs),
		query.String(),
		mode.String(),
		searchTagCursorScope(tagIDs),
		customAssetTypeID.String(),
		string(lifecycleFilter),
		string(checkoutFilter),
	}, ":")
}

func normalizeSearchTagIDs(raw []assettag.ID) []assettag.ID {
	if len(raw) == 0 {
		return nil
	}
	seen := map[assettag.ID]struct{}{}
	ids := make([]assettag.ID, 0, len(raw))
	for _, id := range raw {
		trimmed := assettag.ID(strings.TrimSpace(id.String()))
		if trimmed.String() == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		ids = append(ids, trimmed)
	}
	sort.Slice(ids, func(left int, right int) bool {
		return ids[left].String() < ids[right].String()
	})
	return ids
}

func searchCheckoutStateFilter(raw string) (ports.AssetCheckoutStateFilter, error) {
	switch strings.TrimSpace(raw) {
	case "", string(ports.AssetCheckoutStateFilterAny):
		return ports.AssetCheckoutStateFilterAny, nil
	case string(ports.AssetCheckoutStateFilterCheckedOut):
		return ports.AssetCheckoutStateFilterCheckedOut, nil
	case string(ports.AssetCheckoutStateFilterAvailable):
		return ports.AssetCheckoutStateFilterAvailable, nil
	default:
		return "", ErrInvalidInput
	}
}

func searchInventoryCursorScope(inventoryIDs []inventory.InventoryID) string {
	if len(inventoryIDs) == 0 {
		return "*"
	}
	ids := make([]string, 0, len(inventoryIDs))
	for _, id := range inventoryIDs {
		ids = append(ids, id.String())
	}
	sort.Strings(ids)
	return strings.Join(ids, ",")
}

func searchTagCursorScope(tagIDs []assettag.ID) string {
	if len(tagIDs) == 0 {
		return "*"
	}
	ids := make([]string, 0, len(tagIDs))
	for _, id := range tagIDs {
		ids = append(ids, id.String())
	}
	sort.Strings(ids)
	return strings.Join(ids, ",")
}
