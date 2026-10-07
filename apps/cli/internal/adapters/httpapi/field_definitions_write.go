package httpapi

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

func (c *Client) ChangeFieldDefinition(ctx context.Context, s ports.DefinitionScope, id string, action ports.FieldAction, body []byte) (ports.Result[ports.FieldDefinition], error) {
	if err := validateDefinitionScope(s); err != nil {
		return ports.Result[ports.FieldDefinition]{}, err
	}
	var response *http.Response
	var err error
	if s.Level == ports.HouseholdDefinition {
		switch action {
		case ports.CreateField:
			response, err = c.sdk.PostTenantsByTenantIdCustomFieldDefinitionsWithBody(ctx, s.Scope.Tenant, nil, "application/json", bytes.NewReader(body))
		case ports.UpdateField:
			response, err = c.sdk.PatchTenantsByTenantIdCustomFieldDefinitionsByDefinitionIdWithBody(ctx, s.Scope.Tenant, id, nil, "application/json", bytes.NewReader(body))
		case ports.ArchiveField:
			response, err = c.sdk.PatchTenantsByTenantIdCustomFieldDefinitionsByDefinitionIdArchive(ctx, s.Scope.Tenant, id, nil)
		case ports.RestoreField:
			response, err = c.sdk.PatchTenantsByTenantIdCustomFieldDefinitionsByDefinitionIdRestore(ctx, s.Scope.Tenant, id, nil)
		default:
			return ports.Result[ports.FieldDefinition]{}, ports.Failure("usage", "Unknown field definition action. Use --help to select a command.")
		}
	} else {
		switch action {
		case ports.CreateField:
			response, err = c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdCustomFieldDefinitionsWithBody(ctx, s.Scope.Tenant, s.Scope.Inventory, nil, "application/json", bytes.NewReader(body))
		case ports.UpdateField:
			response, err = c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdCustomFieldDefinitionsByDefinitionIdWithBody(ctx, s.Scope.Tenant, s.Scope.Inventory, id, nil, "application/json", bytes.NewReader(body))
		case ports.ArchiveField:
			response, err = c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdCustomFieldDefinitionsByDefinitionIdArchive(ctx, s.Scope.Tenant, s.Scope.Inventory, id, nil)
		case ports.RestoreField:
			response, err = c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdCustomFieldDefinitionsByDefinitionIdRestore(ctx, s.Scope.Tenant, s.Scope.Inventory, id, nil)
		default:
			return ports.Result[ports.FieldDefinition]{}, ports.Failure("usage", "Unknown field definition action. Use --help to select a command.")
		}
	}
	return fieldDefinitionResult(response, err)
}
