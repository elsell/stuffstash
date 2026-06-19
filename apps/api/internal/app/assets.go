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
	customFields, ok := asset.NewCustomFields(normalizeCustomFieldValues(input.CustomFields))
	if !ok {
		return asset.Asset{}, ErrInvalidInput
	}
	if !customFields.IsEmpty() {
		if a.customFields == nil {
			return asset.Asset{}, ErrInvalidInput
		}
		definitions, err := a.customFields.ListEffectiveCustomFieldDefinitions(ctx, input.TenantID, input.InventoryID)
		if err != nil {
			return asset.Asset{}, err
		}
		if !customfield.DefinitionSet(definitions).ValidateValues(customFields.Values()) {
			return asset.Asset{}, ErrInvalidInput
		}
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
