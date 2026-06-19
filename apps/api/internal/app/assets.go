package app

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type CreateAssetInput struct {
	Principal     identity.Principal
	TenantID      tenant.ID
	InventoryID   inventory.InventoryID
	Kind          string
	Title         string
	Description   string
	ParentAssetID string
	CustomFields  map[string]any
}

type ListAssetsInput struct {
	Principal   identity.Principal
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Limit       int
	Cursor      string
}

type AssetParentUpdate struct {
	Present bool
	Null    bool
	Value   string
}

type UpdateAssetInput struct {
	Principal     identity.Principal
	TenantID      tenant.ID
	InventoryID   inventory.InventoryID
	AssetID       asset.ID
	Title         *string
	Description   *string
	ParentAssetID AssetParentUpdate
	CustomFields  map[string]any
}

type ListAssetsResult struct {
	Items      []asset.Asset
	Limit      int
	NextCursor *string
	HasMore    bool
}

func (a App) CreateAsset(ctx context.Context, input CreateAssetInput) (asset.Asset, error) {
	if err := a.ensureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionCreateAsset); err != nil {
		return asset.Asset{}, err
	}

	kind, ok := asset.NewKind(input.Kind)
	if !ok {
		return asset.Asset{}, ErrInvalidInput
	}
	title, ok := asset.NewTitle(input.Title)
	if !ok {
		return asset.Asset{}, ErrInvalidInput
	}
	customFields, err := a.validatedCustomFields(ctx, input.TenantID, input.InventoryID, input.CustomFields)
	if err != nil {
		return asset.Asset{}, err
	}

	id, ok := asset.NewID(a.ids.NewID())
	if !ok {
		return asset.Asset{}, ErrInvalidInput
	}

	parentAssetID := asset.ID("")
	if strings.TrimSpace(input.ParentAssetID) != "" {
		parsedParentID, ok := asset.NewID(input.ParentAssetID)
		if !ok {
			return asset.Asset{}, ErrInvalidInput
		}
		parent, found, err := a.assets.AssetByID(ctx, input.TenantID, input.InventoryID, parsedParentID)
		if err != nil {
			return asset.Asset{}, err
		}
		if !found {
			return asset.Asset{}, ErrNotFound
		}
		if !parent.Kind.CanContainChildren() || parent.LifecycleState != asset.LifecycleStateActive {
			return asset.Asset{}, ErrInvalidInput
		}
		parentAssetID = parsedParentID
	}

	item := asset.Asset{
		ID:             id,
		TenantID:       asset.TenantID(input.TenantID.String()),
		InventoryID:    asset.InventoryID(input.InventoryID.String()),
		ParentAssetID:  parentAssetID,
		Kind:           kind,
		Title:          title,
		Description:    asset.NewDescription(input.Description),
		CustomFields:   customFields,
		LifecycleState: asset.LifecycleStateActive,
	}

	if err := a.assets.CreateAsset(ctx, item); err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return asset.Asset{}, ErrInvalidInput
		}
		return asset.Asset{}, err
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAssetCreated,
		Message: "asset created",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"asset_id":     item.ID.String(),
			"asset_kind":   item.Kind.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})

	return item, nil
}

func (a App) UpdateAsset(ctx context.Context, input UpdateAssetInput) (asset.Asset, error) {
	if err := a.ensureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return asset.Asset{}, err
	}
	if input.AssetID.String() == "" {
		return asset.Asset{}, ErrInvalidInput
	}

	current, found, err := a.assets.AssetByID(ctx, input.TenantID, input.InventoryID, input.AssetID)
	if err != nil {
		return asset.Asset{}, err
	}
	if !found {
		return asset.Asset{}, ErrNotFound
	}
	updated := current

	if input.Title != nil {
		title, ok := asset.NewTitle(*input.Title)
		if !ok {
			return asset.Asset{}, ErrInvalidInput
		}
		updated.Title = title
	}
	if input.Description != nil {
		updated.Description = asset.NewDescription(*input.Description)
	}
	if input.CustomFields != nil {
		customFields, err := a.validatedCustomFields(ctx, input.TenantID, input.InventoryID, input.CustomFields)
		if err != nil {
			return asset.Asset{}, err
		}
		updated.CustomFields = customFields
	}
	if input.ParentAssetID.Present {
		parentAssetID := asset.ID("")
		if !input.ParentAssetID.Null {
			parentAsset, err := a.validatedParentAssetID(ctx, input.TenantID, input.InventoryID, input.AssetID, input.ParentAssetID.Value)
			if err != nil {
				return asset.Asset{}, err
			}
			parentAssetID = parentAsset
		}
		updated.ParentAssetID = parentAssetID
	}

	if err := a.assets.UpdateAsset(ctx, updated); err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return asset.Asset{}, ErrInvalidInput
		}
		return asset.Asset{}, err
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAssetUpdated,
		Message: "asset updated",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"asset_id":     updated.ID.String(),
			"asset_kind":   updated.Kind.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})

	return updated, nil
}

func (a App) validatedCustomFields(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, values map[string]any) (asset.CustomFields, error) {
	customFields, ok := asset.NewCustomFields(normalizeCustomFieldValues(values))
	if !ok {
		return asset.CustomFields{}, ErrInvalidInput
	}
	if customFields.IsEmpty() {
		return customFields, nil
	}
	if a.customFields == nil {
		return asset.CustomFields{}, ErrInvalidInput
	}
	definitions, err := a.customFields.ListEffectiveCustomFieldDefinitions(ctx, tenantID, inventoryID)
	if err != nil {
		return asset.CustomFields{}, err
	}
	if !customfield.DefinitionSet(definitions).ValidateValues(customFields.Values()) {
		return asset.CustomFields{}, ErrInvalidInput
	}
	return customFields, nil
}

func (a App) validatedParentAssetID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, rawParentAssetID string) (asset.ID, error) {
	if strings.TrimSpace(rawParentAssetID) == "" {
		return asset.ID(""), ErrInvalidInput
	}
	parentAssetID, ok := asset.NewID(rawParentAssetID)
	if !ok || parentAssetID == assetID {
		return asset.ID(""), ErrInvalidInput
	}
	parent, found, err := a.assets.AssetByID(ctx, tenantID, inventoryID, parentAssetID)
	if err != nil {
		return asset.ID(""), err
	}
	if !found {
		return asset.ID(""), ErrNotFound
	}
	if !parent.Kind.CanContainChildren() || parent.LifecycleState != asset.LifecycleStateActive {
		return asset.ID(""), ErrInvalidInput
	}
	for current := parent; current.ParentAssetID.String() != ""; {
		if current.ParentAssetID == assetID {
			return asset.ID(""), ErrInvalidInput
		}
		next, found, err := a.assets.AssetByID(ctx, tenantID, inventoryID, current.ParentAssetID)
		if err != nil {
			return asset.ID(""), err
		}
		if !found {
			return asset.ID(""), ErrInvalidInput
		}
		current = next
	}
	return parentAssetID, nil
}

func normalizeCustomFieldValues(values map[string]any) map[string]any {
	normalized := map[string]any{}
	for key, value := range values {
		normalized[key] = customfield.NormalizeJSONNumber(value)
	}
	return normalized
}

func (a App) ListAssets(ctx context.Context, input ListAssetsInput) (ListAssetsResult, error) {
	if err := a.ensureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return ListAssetsResult{}, err
	}

	limit := pageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	afterAssetID, err := decodeAssetCursor(input.TenantID, input.InventoryID, input.Cursor)
	if err != nil {
		return ListAssetsResult{}, ErrInvalidInput
	}

	items, err := a.assets.ListAssetsByInventory(ctx, input.TenantID, input.InventoryID, ports.AssetListPageRequest{
		AfterAssetID: afterAssetID,
		Limit:        limit + 1,
	})
	if err != nil {
		return ListAssetsResult{}, err
	}

	hasMore := len(items) > limit
	var nextCursor *string
	if hasMore {
		items = items[:limit]
		nextCursor = encodeAssetCursor(input.TenantID, input.InventoryID, items[len(items)-1].ID)
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAssetsListed,
		Message: "assets listed",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"principal_id": input.Principal.ID.String(),
			"limit":        strings.TrimSpace(strconv.Itoa(limit)),
		},
	})

	return ListAssetsResult{
		Items:      items,
		Limit:      limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func encodeAssetCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, id asset.ID) *string {
	return encodePageCursor("assets", tenantID.String()+":"+inventoryID.String(), id.String())
}

func decodeAssetCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, cursor string) (asset.ID, error) {
	decoded, err := decodePageCursor("assets", tenantID.String()+":"+inventoryID.String(), cursor)
	if err != nil {
		return asset.ID(""), err
	}
	if decoded == "" {
		return asset.ID(""), nil
	}
	id, ok := asset.NewID(decoded)
	if !ok {
		return asset.ID(""), ErrInvalidInput
	}
	return id, nil
}

func (a App) ensureInventoryAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID, permission ports.InventoryPermission) error {
	exists, err := a.tenants.TenantExists(ctx, tenantID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}

	_, found, err := a.inventories.InventoryByID(ctx, tenantID, inventoryID)
	if err != nil {
		return err
	}
	if !found {
		return ErrNotFound
	}

	if err := a.authorizer.CheckInventory(ctx, principal, permission, inventoryID); err != nil {
		a.recordAuthorizationDenied(ctx, principal, tenantID)
		return err
	}
	return nil
}
