package search

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// WithAncestorPaths enriches an already authorized page. Every lookup remains
// scoped to that result's tenant and inventory, including archived ancestors.
func WithAncestorPaths(ctx context.Context, repository ports.AssetRepository, results []ports.AssetSearchResult) ([]ports.AssetSearchResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	type key struct {
		tenant    tenant.ID
		inventory inventory.InventoryID
		asset     asset.ID
	}
	known := make(map[key]asset.Asset)
	for _, result := range results {
		known[key{result.TenantID, result.Inventory.ID, result.Asset.ID}] = result.Asset
	}
	enriched := make([]ports.AssetSearchResult, 0, len(results))
	for _, result := range results {
		if result.Asset.TenantID.String() != result.TenantID.String() || result.Asset.InventoryID.String() != result.Inventory.ID.String() {
			return nil, apperrors.ErrInvalidInput
		}
		path := make([]ports.AssetSearchAncestor, 0)
		visited := map[asset.ID]bool{result.Asset.ID: true}
		parentID := result.Asset.ParentAssetID
		for parentID != "" {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			if visited[parentID] {
				return nil, apperrors.ErrInvalidInput
			}
			visited[parentID] = true
			ref := key{result.TenantID, result.Inventory.ID, parentID}
			parent, exists := known[ref]
			if !exists {
				if repository == nil {
					return nil, apperrors.ErrInvalidInput
				}
				var err error
				parent, exists, err = repository.AssetByID(ctx, ref.tenant, ref.inventory, ref.asset)
				if err != nil {
					return nil, err
				}
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				if !exists {
					return nil, apperrors.ErrNotFound
				}
				known[ref] = parent
			}
			if parent.ID != parentID || parent.TenantID.String() != result.TenantID.String() || parent.InventoryID.String() != result.Inventory.ID.String() {
				return nil, apperrors.ErrInvalidInput
			}
			path = append(path, ports.AssetSearchAncestor{ID: parent.ID, Title: parent.Title})
			parentID = parent.ParentAssetID
		}
		for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
			path[left], path[right] = path[right], path[left]
		}
		result.AncestorPath = path
		enriched = append(enriched, result)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return enriched, nil
}
