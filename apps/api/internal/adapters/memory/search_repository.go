package memory

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

func (s *Store) SearchAssets(_ context.Context, tenantID tenant.ID, inventoryIDs []inventory.InventoryID, page ports.AssetSearchPageRequest) ([]ports.AssetSearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	visibleInventoryIDs := map[string]inventory.Inventory{}
	for _, id := range inventoryIDs {
		item, ok := s.inventories[id]
		if ok && item.TenantID.String() == tenantID.String() {
			visibleInventoryIDs[id.String()] = item
		}
	}

	results := []ports.AssetSearchResult{}
	for _, item := range s.assets {
		if item.TenantID.String() != tenantID.String() {
			continue
		}
		containingInventory, ok := visibleInventoryIDs[item.InventoryID.String()]
		if !ok {
			continue
		}
		if !assetLifecycleMatches(item.LifecycleState, page.LifecycleFilter) {
			continue
		}
		if page.CustomAssetTypeID.String() != "" && item.CustomAssetTypeID.String() != page.CustomAssetTypeID.String() {
			continue
		}
		currentCheckout, hasOpenCheckout := s.currentOpenCheckoutForSearch(item)
		if !checkoutStateMatches(hasOpenCheckout, page.CheckoutFilter) {
			continue
		}

		assignedTags := s.assetTagsByAssetLocked(tenantID, inventory.InventoryID(item.InventoryID.String()), item.ID)
		if !assignedTagsContainAll(assignedTags, page.TagIDs) {
			continue
		}
		matches := search.MatchAsset(assetDocument(item, s.customAssetTypes[customfield.AssetTypeID(item.CustomAssetTypeID.String())], assignedTags, searchAttachmentsForAsset(item, s.attachments)), page.Query, page.Mode)
		if page.Query.String() != "" && len(matches) == 0 {
			continue
		}
		result := ports.AssetSearchResult{
			Type:         search.ResultTypeAsset,
			TenantID:     tenantID,
			Inventory:    containingInventory,
			Asset:        item,
			AssignedTags: assignedTags,
			Matches:      matches,
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

func assignedTagsContainAll(assignedTags []assettag.Tag, tagIDs []assettag.ID) bool {
	if len(tagIDs) == 0 {
		return true
	}
	assigned := map[assettag.ID]struct{}{}
	for _, tag := range assignedTags {
		assigned[tag.ID] = struct{}{}
	}
	for _, id := range tagIDs {
		if _, ok := assigned[id]; !ok {
			return false
		}
	}
	return true
}

func (s *Store) currentOpenCheckoutForSearch(item asset.Asset) (asset.Checkout, bool) {
	for _, checkout := range s.checkouts {
		if checkout.TenantID == item.TenantID && checkout.InventoryID == item.InventoryID && checkout.AssetID == item.ID && checkout.State == asset.CheckoutStateOpen {
			return checkout, true
		}
	}
	return asset.Checkout{}, false
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

func searchAttachmentsForAsset(item asset.Asset, attachments map[media.ID]media.Attachment) []media.Attachment {
	result := []media.Attachment{}
	for _, attachment := range attachments {
		if attachment.TenantID.String() != item.TenantID.String() || attachment.InventoryID.String() != item.InventoryID.String() || attachment.AssetID.String() != item.ID.String() {
			continue
		}
		result = append(result, attachment)
	}
	return result
}
