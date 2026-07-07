package gormstore

import (
	"context"
	"fmt"
	"sort"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s Store) SearchAssets(ctx context.Context, tenantID tenant.ID, inventoryIDs []inventory.InventoryID, page ports.AssetSearchPageRequest) ([]ports.AssetSearchResult, error) {
	if len(inventoryIDs) == 0 {
		return []ports.AssetSearchResult{}, nil
	}

	inventoryIDValues := make([]string, 0, len(inventoryIDs))
	for _, id := range inventoryIDs {
		inventoryIDValues = append(inventoryIDValues, id.String())
	}

	var inventoryModels []inventoryModel
	if err := s.db.WithContext(ctx).Where(&inventoryModel{TenantID: tenantID.String()}).Find(&inventoryModels, inventoryIDValues).Error; err != nil {
		return nil, err
	}
	inventories := map[string]inventory.Inventory{}
	for _, model := range inventoryModels {
		item, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid inventory row %q", model.ID)
		}
		inventories[item.ID.String()] = item
	}

	var assetModels []assetModel
	query := s.db.WithContext(ctx).
		Where(&assetModel{TenantID: tenantID.String()}).
		Where(map[string]any{"inventory_id": inventoryIDValues})
	switch page.LifecycleFilter {
	case "", ports.AssetLifecycleFilterActive:
		query = query.Where(&assetModel{LifecycleState: asset.LifecycleStateActive.String()})
	case ports.AssetLifecycleFilterArchived:
		query = query.Where(&assetModel{LifecycleState: asset.LifecycleStateArchived.String()})
	case ports.AssetLifecycleFilterAll:
	default:
		return []ports.AssetSearchResult{}, nil
	}
	if page.CustomAssetTypeID.String() != "" {
		query = query.Where(&assetModel{CustomAssetTypeID: stringPtr(page.CustomAssetTypeID.String())})
	}
	if err := query.Find(&assetModels).Error; err != nil {
		return nil, err
	}
	assetIDs := make([]asset.ID, 0, len(assetModels))
	for _, model := range assetModels {
		assetIDs = append(assetIDs, asset.ID(model.ID))
	}
	currentCheckouts, err := s.searchOpenCheckouts(ctx, tenantID, inventoryIDValues)
	if err != nil {
		return nil, err
	}

	assetTypes, err := s.searchCustomAssetTypes(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	attachments, err := s.searchAttachments(ctx, tenantID, inventoryIDValues)
	if err != nil {
		return nil, err
	}
	assignedTags, err := s.assetTagsByAssetsInInventories(ctx, tenantID, inventoryIDValues, assetIDs)
	if err != nil {
		return nil, err
	}

	results := []ports.AssetSearchResult{}
	for _, model := range assetModels {
		item, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid asset row %q", model.ID)
		}
		containingInventory, ok := inventories[item.InventoryID.String()]
		if !ok {
			continue
		}
		currentCheckout, hasOpenCheckout := currentCheckouts[item.ID.String()]
		if !checkoutStateMatches(hasOpenCheckout, page.CheckoutFilter) {
			continue
		}
		matches := search.MatchAsset(assetDocument(item, assetTypes[customfield.AssetTypeID(item.CustomAssetTypeID.String())], assignedTags[item.ID], attachments[item.ID.String()]), page.Query, page.Mode)
		if len(matches) == 0 {
			continue
		}
		result := ports.AssetSearchResult{
			Type:      search.ResultTypeAsset,
			TenantID:  tenantID,
			Inventory: containingInventory,
			Asset:     item,
			Matches:   matches,
		}
		if hasOpenCheckout {
			result.CurrentCheckout = &currentCheckout
		}
		if result.CursorKey() <= page.AfterResultKey {
			continue
		}
		results = append(results, result)
	}

	sort.Slice(results, func(left int, right int) bool {
		return results[left].CursorKey() < results[right].CursorKey()
	})
	if page.Limit > 0 && len(results) > page.Limit {
		results = results[:page.Limit]
	}
	return results, nil
}

func (s Store) searchOpenCheckouts(ctx context.Context, tenantID tenant.ID, inventoryIDValues []string) (map[string]asset.Checkout, error) {
	var models []assetCheckoutModel
	if err := s.db.WithContext(ctx).
		Where(&assetCheckoutModel{TenantID: tenantID.String(), State: asset.CheckoutStateOpen.String()}).
		Where(map[string]any{"inventory_id": inventoryIDValues}).
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := map[string]asset.Checkout{}
	for _, model := range models {
		checkout, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid asset checkout row %q", model.ID)
		}
		result[checkout.AssetID.String()] = checkout
	}
	return result, nil
}

func checkoutStateMatches(hasOpenCheckout bool, filter ports.AssetCheckoutStateFilter) bool {
	switch filter {
	case "", ports.AssetCheckoutStateFilterAny:
		return true
	case ports.AssetCheckoutStateFilterCheckedOut:
		return hasOpenCheckout
	case ports.AssetCheckoutStateFilterAvailable:
		return !hasOpenCheckout
	default:
		return false
	}
}

func (s Store) searchCustomAssetTypes(ctx context.Context, tenantID tenant.ID) (map[customfield.AssetTypeID]customfield.AssetType, error) {
	var models []customAssetTypeModel
	if err := s.db.WithContext(ctx).Where(&customAssetTypeModel{TenantID: tenantID.String()}).Find(&models).Error; err != nil {
		return nil, err
	}
	result := map[customfield.AssetTypeID]customfield.AssetType{}
	for _, model := range models {
		assetType, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid custom asset type row %q", model.ID)
		}
		result[assetType.ID] = assetType
	}
	return result, nil
}

func (s Store) searchAttachments(ctx context.Context, tenantID tenant.ID, inventoryIDValues []string) (map[string][]media.Attachment, error) {
	var models []attachmentModel
	if err := s.db.WithContext(ctx).
		Where(&attachmentModel{TenantID: tenantID.String()}).
		Where(map[string]any{"inventory_id": inventoryIDValues}).
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := map[string][]media.Attachment{}
	for _, model := range models {
		attachment, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid attachment row %q", model.ID)
		}
		result[attachment.AssetID.String()] = append(result[attachment.AssetID.String()], attachment)
	}
	return result, nil
}

func assetDocument(item asset.Asset, assetType customfield.AssetType, tags []assettag.Tag, attachments []media.Attachment) search.AssetDocument {
	fields := make([]string, 0, len(item.CustomFields.Values()))
	for _, value := range item.CustomFields.Values() {
		fields = append(fields, fmt.Sprint(value))
	}
	tagDocuments := make([]search.TagDocument, 0, len(tags))
	for _, tag := range tags {
		tagDocuments = append(tagDocuments, search.TagDocument{
			Key:         tag.Key.String(),
			DisplayName: tag.DisplayName.String(),
		})
	}
	attachmentDocuments := make([]search.AttachmentDocument, 0, len(attachments))
	for _, attachment := range attachments {
		attachmentDocuments = append(attachmentDocuments, search.AttachmentDocument{
			FileName:    attachment.FileName.String(),
			ContentType: attachment.ContentType.String(),
		})
	}
	return search.AssetDocument{
		Title:               item.Title.String(),
		Description:         item.Description.String(),
		CustomFields:        fields,
		CustomAssetTypeKey:  assetType.Key.String(),
		CustomAssetTypeName: assetType.DisplayName.String(),
		CustomAssetTypeText: assetType.Description.String(),
		Tags:                tagDocuments,
		Attachments:         attachmentDocuments,
	}
}
