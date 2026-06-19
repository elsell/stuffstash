package app

import (
	"context"
	"errors"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type CreateCustomFieldDefinitionInput struct {
	Principal   identity.Principal
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Key         string
	DisplayName string
	Type        string
	EnumOptions []string
}

type ListCustomFieldDefinitionsInput struct {
	Principal   identity.Principal
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Limit       int
	Cursor      string
}

type ListCustomFieldDefinitionsResult struct {
	Items      []customfield.Definition
	Limit      int
	NextCursor *string
	HasMore    bool
}

func (a App) CreateTenantCustomFieldDefinition(ctx context.Context, input CreateCustomFieldDefinitionInput) (customfield.Definition, error) {
	if err := a.ensureTenantExists(ctx, input.TenantID); err != nil {
		return customfield.Definition{}, err
	}
	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return customfield.Definition{}, err
	}

	return a.createCustomFieldDefinition(ctx, input, customfield.ScopeTenant)
}

func (a App) CreateInventoryCustomFieldDefinition(ctx context.Context, input CreateCustomFieldDefinitionInput) (customfield.Definition, error) {
	if err := a.ensureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return customfield.Definition{}, err
	}

	return a.createCustomFieldDefinition(ctx, input, customfield.ScopeInventory)
}

func (a App) createCustomFieldDefinition(ctx context.Context, input CreateCustomFieldDefinitionInput, scope customfield.Scope) (customfield.Definition, error) {
	id, ok := customfield.NewID(a.ids.NewID())
	if !ok {
		return customfield.Definition{}, ErrInvalidInput
	}
	key, ok := customfield.NewKey(input.Key)
	if !ok {
		return customfield.Definition{}, ErrInvalidInput
	}
	displayName, ok := customfield.NewDisplayName(input.DisplayName)
	if !ok {
		return customfield.Definition{}, ErrInvalidInput
	}
	fieldType, ok := customfield.NewFieldType(input.Type)
	if !ok {
		return customfield.Definition{}, ErrInvalidInput
	}
	enumOptions, ok := customFieldEnumOptions(input.EnumOptions)
	if !ok {
		return customfield.Definition{}, ErrInvalidInput
	}

	inventoryID := customfield.InventoryID("")
	if scope == customfield.ScopeInventory {
		inventoryID = customfield.InventoryID(input.InventoryID.String())
	}
	definition, ok := customfield.NewDefinition(
		id,
		customfield.TenantID(input.TenantID.String()),
		inventoryID,
		scope,
		key,
		displayName,
		fieldType,
		enumOptions,
	)
	if !ok {
		return customfield.Definition{}, ErrInvalidInput
	}

	if err := a.customFields.SaveCustomFieldDefinition(ctx, definition); err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return customfield.Definition{}, ErrInvalidInput
		}
		return customfield.Definition{}, err
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventCustomFieldDefinitionCreated,
		Message: "custom field definition created",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"principal_id":  input.Principal.ID.String(),
			"definition_id": definition.ID.String(),
			"field_key":     definition.Key.String(),
			"scope":         definition.Scope.String(),
		},
	})

	return definition, nil
}

func (a App) ListTenantCustomFieldDefinitions(ctx context.Context, input ListCustomFieldDefinitionsInput) (ListCustomFieldDefinitionsResult, error) {
	if err := a.ensureTenantExists(ctx, input.TenantID); err != nil {
		return ListCustomFieldDefinitionsResult{}, err
	}
	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return ListCustomFieldDefinitionsResult{}, err
	}

	limit := pageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	afterDefinitionKey, err := decodeCustomFieldDefinitionCursor(input.TenantID, input.InventoryID, input.Cursor)
	if err != nil {
		return ListCustomFieldDefinitionsResult{}, ErrInvalidInput
	}
	items, err := a.customFields.ListTenantCustomFieldDefinitions(ctx, input.TenantID, ports.CustomFieldDefinitionPageRequest{
		AfterDefinitionKey: afterDefinitionKey,
		Limit:              limit + 1,
	})
	if err != nil {
		return ListCustomFieldDefinitionsResult{}, err
	}

	return a.customFieldDefinitionListResult(ctx, input, items, limit), nil
}

func (a App) ListInventoryCustomFieldDefinitions(ctx context.Context, input ListCustomFieldDefinitionsInput) (ListCustomFieldDefinitionsResult, error) {
	if err := a.ensureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return ListCustomFieldDefinitionsResult{}, err
	}

	limit := pageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	afterDefinitionKey, err := decodeCustomFieldDefinitionCursor(input.TenantID, input.InventoryID, input.Cursor)
	if err != nil {
		return ListCustomFieldDefinitionsResult{}, ErrInvalidInput
	}
	items, err := a.customFields.ListInventoryCustomFieldDefinitions(ctx, input.TenantID, input.InventoryID, ports.CustomFieldDefinitionPageRequest{
		AfterDefinitionKey: afterDefinitionKey,
		Limit:              limit + 1,
	})
	if err != nil {
		return ListCustomFieldDefinitionsResult{}, err
	}

	return a.customFieldDefinitionListResult(ctx, input, items, limit), nil
}

func (a App) customFieldDefinitionListResult(ctx context.Context, input ListCustomFieldDefinitionsInput, items []customfield.Definition, limit int) ListCustomFieldDefinitionsResult {
	hasMore := len(items) > limit
	var nextCursor *string
	if hasMore {
		items = items[:limit]
		nextCursor = encodeCustomFieldDefinitionCursor(input.TenantID, input.InventoryID, items[len(items)-1].CursorKey())
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventCustomFieldDefinitionsListed,
		Message: "custom field definitions listed",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"principal_id": input.Principal.ID.String(),
			"limit":        strconv.Itoa(limit),
		},
	})

	return ListCustomFieldDefinitionsResult{
		Items:      items,
		Limit:      limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}
}

func customFieldEnumOptions(values []string) ([]customfield.Key, bool) {
	options := make([]customfield.Key, 0, len(values))
	for _, value := range values {
		option, ok := customfield.NewKey(value)
		if !ok {
			return nil, false
		}
		options = append(options, option)
	}
	return options, true
}

func encodeCustomFieldDefinitionCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, key string) *string {
	return encodePageCursor("custom_field_definitions", customFieldDefinitionCursorScope(tenantID, inventoryID), key)
}

func decodeCustomFieldDefinitionCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, cursor string) (string, error) {
	decoded, err := decodePageCursor("custom_field_definitions", customFieldDefinitionCursorScope(tenantID, inventoryID), cursor)
	if err != nil {
		return "", err
	}
	if decoded == "" {
		return "", nil
	}
	return decoded, nil
}

func customFieldDefinitionCursorScope(tenantID tenant.ID, inventoryID inventory.InventoryID) string {
	if inventoryID.String() == "" {
		return tenantID.String()
	}
	return tenantID.String() + ":" + inventoryID.String()
}

func (a App) ensureTenantExists(ctx context.Context, tenantID tenant.ID) error {
	exists, err := a.tenants.TenantExists(ctx, tenantID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	return nil
}
