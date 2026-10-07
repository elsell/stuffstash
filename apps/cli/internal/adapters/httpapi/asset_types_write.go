package httpapi

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

func (c *Client) ChangeAssetType(ctx context.Context, s ports.DefinitionScope, id string, action ports.TypeAction, body []byte) (ports.Result[ports.AssetType], error) {
	if err := validateDefinitionScope(s); err != nil {
		return ports.Result[ports.AssetType]{}, err
	}
	var response *http.Response
	var err error
	if s.Level == ports.HouseholdDefinition {
		switch action {
		case ports.CreateType:
			response, err = c.sdk.PostTenantsByTenantIdCustomAssetTypesWithBody(ctx, s.Scope.Tenant, nil, "application/json", bytes.NewReader(body))
		case ports.UpdateType:
			response, err = c.sdk.PatchTenantsByTenantIdCustomAssetTypesByCustomAssetTypeIdWithBody(ctx, s.Scope.Tenant, id, nil, "application/json", bytes.NewReader(body))
		case ports.ArchiveType:
			response, err = c.sdk.PatchTenantsByTenantIdCustomAssetTypesByCustomAssetTypeIdArchive(ctx, s.Scope.Tenant, id, nil)
		case ports.RestoreType:
			response, err = c.sdk.PatchTenantsByTenantIdCustomAssetTypesByCustomAssetTypeIdRestore(ctx, s.Scope.Tenant, id, nil)
		default:
			return ports.Result[ports.AssetType]{}, ports.Failure("usage", "Unknown asset type action. Use --help to select a command.")
		}
	} else {
		switch action {
		case ports.CreateType:
			response, err = c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdCustomAssetTypesWithBody(ctx, s.Scope.Tenant, s.Scope.Inventory, nil, "application/json", bytes.NewReader(body))
		case ports.UpdateType:
			response, err = c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdCustomAssetTypesByCustomAssetTypeIdWithBody(ctx, s.Scope.Tenant, s.Scope.Inventory, id, nil, "application/json", bytes.NewReader(body))
		case ports.ArchiveType:
			response, err = c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdCustomAssetTypesByCustomAssetTypeIdArchive(ctx, s.Scope.Tenant, s.Scope.Inventory, id, nil)
		case ports.RestoreType:
			response, err = c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdCustomAssetTypesByCustomAssetTypeIdRestore(ctx, s.Scope.Tenant, s.Scope.Inventory, id, nil)
		default:
			return ports.Result[ports.AssetType]{}, ports.Failure("usage", "Unknown asset type action. Use --help to select a command.")
		}
	}
	return assetTypeResult(response, err)
}
