package assets

import (
	"context"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s Service) GetAsset(ctx context.Context, input GetAssetInput) (asset.Asset, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return asset.Asset{}, err
	}
	if err := s.ensureAssetRepository(); err != nil {
		return asset.Asset{}, err
	}
	if input.AssetID.String() == "" {
		return asset.Asset{}, apperrors.ErrInvalidInput
	}
	item, found, err := s.assets.AssetByID(ctx, input.TenantID, input.InventoryID, input.AssetID)
	if err != nil {
		return asset.Asset{}, err
	}
	if !found {
		return asset.Asset{}, apperrors.ErrNotFound
	}
	if err := s.saveReadAuditRecord(ctx, auditRecordInput{
		PrincipalID: input.Principal.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionAssetViewed,
		TargetType:  audit.TargetAsset,
		TargetID:    item.ID.String(),
		Metadata: map[string]string{
			"asset_kind":      item.Kind.String(),
			"lifecycle_state": item.LifecycleState.String(),
		},
	}); err != nil {
		return asset.Asset{}, err
	}
	s.observer.Record(ctx, ports.Event{
		Name:    ports.EventAssetViewed,
		Message: "asset viewed",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"asset_id":     item.ID.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})
	return item, nil
}

func (s Service) ListAssets(ctx context.Context, input ListAssetsInput) (ListAssetsResult, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return ListAssetsResult{}, err
	}
	if err := s.ensureAssetRepository(); err != nil {
		return ListAssetsResult{}, err
	}

	limit := pageLimit(s.defaultPageLimit, s.maxPageLimit, input.Limit)
	lifecycleFilter, err := LifecycleFilter(input.LifecycleState)
	if err != nil {
		return ListAssetsResult{}, apperrors.ErrInvalidInput
	}
	sort, err := AssetSort(input.Sort)
	if err != nil {
		return ListAssetsResult{}, apperrors.ErrInvalidInput
	}
	cursorPosition, err := decodeAssetCursor(input.TenantID, input.InventoryID, lifecycleFilter, sort, input.Cursor)
	if err != nil {
		return ListAssetsResult{}, apperrors.ErrInvalidInput
	}

	items, err := s.assets.ListAssetsByInventory(ctx, input.TenantID, input.InventoryID, ports.AssetListPageRequest{
		AfterAssetID:    cursorPosition.AssetID,
		AfterUpdatedAt:  cursorPosition.UpdatedAt,
		Limit:           limit + 1,
		LifecycleFilter: lifecycleFilter,
		Sort:            sort,
	})
	if err != nil {
		return ListAssetsResult{}, err
	}

	hasMore := len(items) > limit
	var nextCursor *string
	if hasMore {
		items = items[:limit]
		nextCursor = encodeAssetCursor(input.TenantID, input.InventoryID, lifecycleFilter, sort, items[len(items)-1])
	}

	s.observer.Record(ctx, ports.Event{
		Name:    ports.EventAssetsListed,
		Message: "assets listed",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"principal_id": input.Principal.ID.String(),
			"limit":        strconv.Itoa(limit),
			"lifecycle":    string(lifecycleFilter),
			"sort":         string(sort),
		},
	})
	if err := s.saveReadAuditRecord(ctx, auditRecordInput{
		PrincipalID: input.Principal.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionAssetListed,
		TargetType:  audit.TargetInventory,
		TargetID:    input.InventoryID.String(),
		Metadata: map[string]string{
			"limit":     strconv.Itoa(limit),
			"lifecycle": string(lifecycleFilter),
			"sort":      string(sort),
		},
	}); err != nil {
		return ListAssetsResult{}, err
	}

	return ListAssetsResult{
		Items:      items,
		Limit:      limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}
