package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

func assetType(v generated.AssetTypeResponse) ports.AssetType {
	return ports.AssetType{ID: v.Id, TenantID: v.TenantId, InventoryID: v.InventoryId, Scope: v.Scope, Key: v.Key, DisplayName: v.DisplayName, Description: v.Description, ExpirationEnabled: v.ExpirationEnabled, Lifecycle: v.LifecycleState}
}
func assetTypeResult(response *http.Response, err error) (ports.Result[ports.AssetType], error) {
	r, err := read[generated.SuccessEnvelopeAssetTypeResponse](response, err)
	if err != nil {
		return ports.Result[ports.AssetType]{}, err
	}
	return ports.Result[ports.AssetType]{Data: assetType(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) AssetTypes(ctx context.Context, s ports.DefinitionScope, pageInput ports.Page, lifecycle string) (ports.Result[[]ports.AssetType], error) {
	if err := validateDefinitionScope(s); err != nil {
		return ports.Result[[]ports.AssetType]{}, err
	}
	var response *http.Response
	var err error
	if s.Level == ports.HouseholdDefinition {
		p := &generated.GetTenantsByTenantIdCustomAssetTypesParams{Limit: &pageInput.Limit, Cursor: &pageInput.Cursor}
		if lifecycle != "" {
			v := generated.GetTenantsByTenantIdCustomAssetTypesParamsLifecycleState(lifecycle)
			p.LifecycleState = &v
		}
		response, err = c.sdk.GetTenantsByTenantIdCustomAssetTypes(ctx, s.Scope.Tenant, p)
	} else {
		p := &generated.GetTenantsByTenantIdInventoriesByInventoryIdCustomAssetTypesParams{Limit: &pageInput.Limit, Cursor: &pageInput.Cursor}
		if lifecycle != "" {
			v := generated.GetTenantsByTenantIdInventoriesByInventoryIdCustomAssetTypesParamsLifecycleState(lifecycle)
			p.LifecycleState = &v
		}
		response, err = c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdCustomAssetTypes(ctx, s.Scope.Tenant, s.Scope.Inventory, p)
	}
	r, err := read[generated.SuccessEnvelopeListAssetTypeResponse](response, err)
	if err != nil {
		return ports.Result[[]ports.AssetType]{}, err
	}
	var items []ports.AssetType
	if r.Data.GetOrEmpty() != nil {
		items = make([]ports.AssetType, 0, len(r.Data.GetOrEmpty()))
	}
	for _, v := range r.Data.GetOrEmpty() {
		items = append(items, assetType(v))
	}
	return ports.Result[[]ports.AssetType]{Data: items, Schema: r.Schema, Meta: metadata(r.Meta), Pagination: page(r.Meta)}, nil
}
func (c *Client) AssetType(ctx context.Context, s ports.DefinitionScope, id string) (ports.Result[ports.AssetType], error) {
	if err := validateDefinitionScope(s); err != nil {
		return ports.Result[ports.AssetType]{}, err
	}
	if s.Level == ports.HouseholdDefinition {
		return assetTypeResult(c.sdk.GetTenantsByTenantIdCustomAssetTypesByCustomAssetTypeId(ctx, s.Scope.Tenant, id, nil))
	}
	return assetTypeResult(c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdCustomAssetTypesByCustomAssetTypeId(ctx, s.Scope.Tenant, s.Scope.Inventory, id, nil))
}
func (c *Client) DeleteAssetType(ctx context.Context, s ports.DefinitionScope, id string) error {
	if err := validateDefinitionScope(s); err != nil {
		return err
	}
	if s.Level == ports.HouseholdDefinition {
		return noContent(c.sdk.DeleteTenantsByTenantIdCustomAssetTypesByCustomAssetTypeId(ctx, s.Scope.Tenant, id, nil))
	}
	return noContent(c.sdk.DeleteTenantsByTenantIdInventoriesByInventoryIdCustomAssetTypesByCustomAssetTypeId(ctx, s.Scope.Tenant, s.Scope.Inventory, id, nil))
}
