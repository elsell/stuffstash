package app

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) existingFieldKeys(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (map[string]struct{}, error) {
	keys := map[string]struct{}{}
	if a.customFields == nil {
		return keys, nil
	}
	fields, err := a.customFields.ListEffectiveCustomFieldDefinitions(ctx, tenantID, inventoryID)
	if err != nil {
		return nil, err
	}
	for _, field := range fields {
		keys[field.Key.String()] = struct{}{}
	}
	return keys, nil
}

func (a App) existingHomeboxReferences(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (map[string]struct{}, error) {
	ids := map[string]struct{}{}
	if a.assets == nil {
		return ids, nil
	}
	items, err := a.assets.ListAssetsByInventory(ctx, tenantID, inventoryID, ports.AssetListPageRequest{
		Limit:           10000,
		LifecycleFilter: ports.AssetLifecycleFilterAll,
		Sort:            ports.AssetListSortIDAsc,
	})
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		for _, key := range []string{"homebox-source-id", "homebox-asset-id"} {
			if value, ok := item.CustomFields.Values()[key].(string); ok && strings.TrimSpace(value) != "" {
				ids[key+"="+strings.TrimSpace(value)] = struct{}{}
			}
		}
	}
	return ids, nil
}

func (a App) duplicateWarnings(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, plan importplan.Plan, linkedAssetSourceIDs map[string]struct{}) []importplan.Message {
	duplicates, err := a.existingHomeboxReferences(ctx, tenantID, inventoryID)
	if err != nil {
		return nil
	}
	var messages []importplan.Message
	for _, planned := range plan.Assets {
		if _, linked := linkedAssetSourceIDs[planned.SourceID]; linked {
			continue
		}
		for _, key := range []string{"homebox-source-id", "homebox-asset-id"} {
			value, ok := planned.CustomFields[key].(string)
			if !ok || strings.TrimSpace(value) == "" {
				continue
			}
			if _, duplicate := duplicates[key+"="+strings.TrimSpace(value)]; duplicate {
				messages = append(messages, importplan.Message{
					Code:       "duplicate-source-asset",
					Severity:   importplan.SeverityWarning,
					Summary:    "Asset appears to have already been imported",
					Detail:     key + "=" + strings.TrimSpace(value),
					SourceID:   planned.SourceID,
					SourceName: planned.Title,
				})
				break
			}
		}
	}
	return messages
}

func archivedWarnings(plan importplan.Plan) []importplan.Message {
	var messages []importplan.Message
	for _, planned := range plan.Assets {
		if !planned.Archived {
			continue
		}
		messages = append(messages, importplan.Message{
			Code:       "archived-source-asset-skipped",
			Severity:   importplan.SeverityWarning,
			Summary:    "Archived Homebox asset will be skipped",
			Detail:     "archived source assets are not imported in this version",
			SourceID:   planned.SourceID,
			SourceName: planned.Title,
		})
	}
	return messages
}

func sortedImportAssets(items []importplan.Asset, kind string) []importplan.Asset {
	return sortedImportAssetsByPredicate(items, func(item importplan.Asset) bool {
		return item.Kind == kind
	})
}

func sortedNonLocationImportAssets(items []importplan.Asset) []importplan.Asset {
	return sortedImportAssetsByPredicate(items, func(item importplan.Asset) bool {
		return item.Kind != "location"
	})
}

func sortedImportAssetsByPredicate(items []importplan.Asset, include func(importplan.Asset) bool) []importplan.Asset {
	bySourceID := map[string]importplan.Asset{}
	children := map[string][]importplan.Asset{}
	for _, item := range items {
		if include(item) {
			bySourceID[item.SourceID] = item
			children[item.ParentSourceID] = append(children[item.ParentSourceID], item)
		}
	}
	for parentID := range children {
		sort.SliceStable(children[parentID], func(left, right int) bool {
			return children[parentID][left].Title < children[parentID][right].Title
		})
	}
	var sorted []importplan.Asset
	visited := map[string]struct{}{}
	var visit func(importplan.Asset)
	visit = func(item importplan.Asset) {
		if _, ok := visited[item.SourceID]; ok {
			return
		}
		if parent, ok := bySourceID[item.ParentSourceID]; ok {
			visit(parent)
		}
		visited[item.SourceID] = struct{}{}
		sorted = append(sorted, item)
		for _, child := range children[item.SourceID] {
			visit(child)
		}
	}
	for _, root := range children[""] {
		visit(root)
	}
	var remaining []importplan.Asset
	for sourceID, item := range bySourceID {
		if _, ok := visited[sourceID]; !ok {
			remaining = append(remaining, item)
		}
	}
	sort.SliceStable(remaining, func(left, right int) bool {
		return remaining[left].Title < remaining[right].Title
	})
	for _, item := range remaining {
		visit(item)
	}
	return sorted
}

func stripAttachmentContent(plan *importplan.Plan) {
	for index := range plan.Attachments {
		plan.Attachments[index].Content = nil
	}
}

func safeImportError(err error) string {
	switch {
	case errors.Is(err, ErrAttachmentTooLarge):
		return "attachment is too large"
	case errors.Is(err, ErrAttachmentFileNameInvalid):
		return "attachment file name is invalid"
	case errors.Is(err, ErrAttachmentContentTypeUnsupported):
		return "attachment file type is unsupported"
	case errors.Is(err, ErrAttachmentContentMismatch):
		return "attachment content did not match its file type"
	case errors.Is(err, ErrAttachmentContentEmpty):
		return "attachment content was empty"
	default:
		return "import validation failed"
	}
}

func safeImportAttachmentUnavailableReason(reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return "attachment could not be downloaded"
	}
	return reason
}
