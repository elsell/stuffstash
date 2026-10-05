package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

func fieldDefinition(v generated.DefinitionResponse) ports.FieldDefinition {
	return ports.FieldDefinition{ID: v.Id, TenantID: v.TenantId, InventoryID: v.InventoryId, Scope: v.Scope, Key: v.Key, DisplayName: v.DisplayName, Type: v.Type, EnumOptions: v.EnumOptions.GetOrEmpty(), Applicability: v.Applicability, CustomAssetTypeIDs: v.CustomAssetTypeIds.GetOrEmpty(), Lifecycle: v.LifecycleState}
}
func fieldDefinitionResult(response *http.Response, err error) (ports.Result[ports.FieldDefinition], error) {
	r, err := read[generated.SuccessEnvelopeDefinitionResponse](response, err)
	if err != nil {
		return ports.Result[ports.FieldDefinition]{}, err
	}
	return ports.Result[ports.FieldDefinition]{Data: fieldDefinition(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) FieldDefinitions(ctx context.Context, s ports.DefinitionScope, pageInput ports.Page, lifecycle string) (ports.Result[[]ports.FieldDefinition], error) {
	if err := validateDefinitionScope(s); err != nil {
		return ports.Result[[]ports.FieldDefinition]{}, err
	}
	var response *http.Response
	var err error
	if s.Level == ports.HouseholdDefinition {
		p := &generated.GetTenantsByTenantIdCustomFieldDefinitionsParams{Limit: &pageInput.Limit, Cursor: &pageInput.Cursor}
		if lifecycle != "" {
			v := generated.GetTenantsByTenantIdCustomFieldDefinitionsParamsLifecycleState(lifecycle)
			p.LifecycleState = &v
		}
		response, err = c.sdk.GetTenantsByTenantIdCustomFieldDefinitions(ctx, s.Scope.Tenant, p)
	} else {
		p := &generated.GetTenantsByTenantIdInventoriesByInventoryIdCustomFieldDefinitionsParams{Limit: &pageInput.Limit, Cursor: &pageInput.Cursor}
		if lifecycle != "" {
			v := generated.GetTenantsByTenantIdInventoriesByInventoryIdCustomFieldDefinitionsParamsLifecycleState(lifecycle)
			p.LifecycleState = &v
		}
		response, err = c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdCustomFieldDefinitions(ctx, s.Scope.Tenant, s.Scope.Inventory, p)
	}
	r, err := read[generated.SuccessEnvelopeListDefinitionResponse](response, err)
	if err != nil {
		return ports.Result[[]ports.FieldDefinition]{}, err
	}
	var items []ports.FieldDefinition
	if r.Data.GetOrEmpty() != nil {
		items = make([]ports.FieldDefinition, 0, len(r.Data.GetOrEmpty()))
	}
	for _, v := range r.Data.GetOrEmpty() {
		items = append(items, fieldDefinition(v))
	}
	return ports.Result[[]ports.FieldDefinition]{Data: items, Schema: r.Schema, Meta: metadata(r.Meta), Pagination: page(r.Meta)}, nil
}
func (c *Client) FieldDefinition(ctx context.Context, s ports.DefinitionScope, id string) (ports.Result[ports.FieldDefinition], error) {
	if err := validateDefinitionScope(s); err != nil {
		return ports.Result[ports.FieldDefinition]{}, err
	}
	if s.Level == ports.HouseholdDefinition {
		return fieldDefinitionResult(c.sdk.GetTenantsByTenantIdCustomFieldDefinitionsByDefinitionId(ctx, s.Scope.Tenant, id, nil))
	}
	return fieldDefinitionResult(c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdCustomFieldDefinitionsByDefinitionId(ctx, s.Scope.Tenant, s.Scope.Inventory, id, nil))
}
func (c *Client) DeleteFieldDefinition(ctx context.Context, s ports.DefinitionScope, id string) error {
	if err := validateDefinitionScope(s); err != nil {
		return err
	}
	if s.Level == ports.HouseholdDefinition {
		return noContent(c.sdk.DeleteTenantsByTenantIdCustomFieldDefinitionsByDefinitionId(ctx, s.Scope.Tenant, id, nil))
	}
	return noContent(c.sdk.DeleteTenantsByTenantIdInventoriesByInventoryIdCustomFieldDefinitionsByDefinitionId(ctx, s.Scope.Tenant, s.Scope.Inventory, id, nil))
}
