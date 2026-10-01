package search

import (
	"context"
	"sort"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

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

func (a Service) inventoryIDsForTenant(ctx context.Context, tenantID tenant.ID) ([]inventory.InventoryID, error) {
	items, err := a.deps.Inventories.ListInventoriesByTenant(ctx, inventory.TenantID(tenantID.String()), ports.InventoryListPageRequest{})
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
		return "", apperrors.ErrInvalidInput
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
		return "", apperrors.ErrInvalidInput
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
