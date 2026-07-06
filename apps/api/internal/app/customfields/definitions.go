package customfields

import (
	"context"
	"errors"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type CreateCustomFieldDefinitionInput struct {
	Principal          identity.Principal
	Source             audit.Source
	RequestID          string
	TenantID           tenant.ID
	InventoryID        inventory.InventoryID
	Key                string
	DisplayName        string
	Type               string
	EnumOptions        []string
	Applicability      string
	CustomAssetTypeIDs []string
}

type ListCustomFieldDefinitionsInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Limit       int
	Cursor      string
}

type GetCustomFieldDefinitionInput struct {
	Principal    identity.Principal
	Source       audit.Source
	RequestID    string
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	DefinitionID customfield.ID
}

type UpdateCustomFieldDefinitionLifecycleInput struct {
	Principal    identity.Principal
	Source       audit.Source
	RequestID    string
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	DefinitionID customfield.ID
}

type UpdateCustomFieldDefinitionInput struct {
	Principal          identity.Principal
	Source             audit.Source
	RequestID          string
	TenantID           tenant.ID
	InventoryID        inventory.InventoryID
	DefinitionID       customfield.ID
	DisplayName        *string
	Key                *string
	Type               *string
	EnumOptions        *[]string
	Applicability      *string
	CustomAssetTypeIDs *[]string
}

type ListCustomFieldDefinitionsResult struct {
	Items      []customfield.Definition
	Limit      int
	NextCursor *string
	HasMore    bool
}

func (s Service) CreateTenantCustomFieldDefinition(ctx context.Context, input CreateCustomFieldDefinitionInput) (customfield.Definition, error) {
	if err := s.ensureTenantExists(ctx, input.TenantID); err != nil {
		return customfield.Definition{}, err
	}
	if err := s.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		s.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return customfield.Definition{}, err
	}

	return s.createCustomFieldDefinition(ctx, input, customfield.ScopeTenant)
}

func (s Service) CreateInventoryCustomFieldDefinition(ctx context.Context, input CreateCustomFieldDefinitionInput) (customfield.Definition, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return customfield.Definition{}, err
	}

	return s.createCustomFieldDefinition(ctx, input, customfield.ScopeInventory)
}

func (s Service) UpdateTenantCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionInput) (customfield.Definition, error) {
	if err := s.ensureTenantExists(ctx, input.TenantID); err != nil {
		return customfield.Definition{}, err
	}
	if err := s.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		s.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return customfield.Definition{}, err
	}
	return s.updateCustomFieldDefinition(ctx, input, customfield.ScopeTenant)
}

func (s Service) UpdateInventoryCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionInput) (customfield.Definition, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return customfield.Definition{}, err
	}
	return s.updateCustomFieldDefinition(ctx, input, customfield.ScopeInventory)
}

func (s Service) createCustomFieldDefinition(ctx context.Context, input CreateCustomFieldDefinitionInput, scope customfield.Scope) (customfield.Definition, error) {
	id, ok := customfield.NewID(s.ids.NewID())
	if !ok {
		return customfield.Definition{}, apperrors.ErrInvalidInput
	}
	key, ok := customfield.NewKey(input.Key)
	if !ok {
		return customfield.Definition{}, apperrors.ErrInvalidInput
	}
	displayName, ok := customfield.NewDisplayName(input.DisplayName)
	if !ok {
		return customfield.Definition{}, apperrors.ErrInvalidInput
	}
	fieldType, ok := customfield.NewFieldType(input.Type)
	if !ok {
		return customfield.Definition{}, apperrors.ErrInvalidInput
	}
	enumOptions, ok := customFieldEnumOptions(input.EnumOptions)
	if !ok {
		return customfield.Definition{}, apperrors.ErrInvalidInput
	}
	applicability, ok := customfield.NewApplicability(input.Applicability)
	if !ok {
		return customfield.Definition{}, apperrors.ErrInvalidInput
	}
	customAssetTypeIDs, err := s.validatedCustomFieldTargetIDs(ctx, input, scope, applicability)
	if err != nil {
		return customfield.Definition{}, err
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
		applicability,
		customAssetTypeIDs,
	)
	if !ok {
		return customfield.Definition{}, apperrors.ErrInvalidInput
	}

	auditRecord, err := s.newAuditRecord(appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionCustomFieldDefinitionCreated,
		TargetType:  audit.TargetCustomFieldDefinition,
		TargetID:    definition.ID.String(),
		Metadata: map[string]string{
			"field_key":     definition.Key.String(),
			"scope":         definition.Scope.String(),
			"applicability": definition.Applicability.String(),
			"target_count":  strconv.Itoa(len(definition.CustomAssetTypeIDs)),
		},
	})
	if err != nil {
		return customfield.Definition{}, err
	}

	if err := s.customFieldUnitOfWork.SaveCustomFieldDefinition(ctx, definition, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return customfield.Definition{}, apperrors.ErrInvalidInput
		}
		return customfield.Definition{}, err
	}

	s.observer.Record(ctx, ports.Event{
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

func (s Service) updateCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionInput, scope customfield.Scope) (customfield.Definition, error) {
	definitionID, ok := customfield.NewID(input.DefinitionID.String())
	if !ok || updateCustomFieldDefinitionInputIsEmpty(input) {
		return customfield.Definition{}, apperrors.ErrInvalidInput
	}
	current, found, err := s.customFields.CustomFieldDefinitionByID(ctx, input.TenantID, input.InventoryID, definitionID)
	if err != nil {
		return customfield.Definition{}, err
	}
	if !found {
		return customfield.Definition{}, apperrors.ErrNotFound
	}
	if current.Scope != scope {
		return customfield.Definition{}, apperrors.ErrNotFound
	}
	if scope == customfield.ScopeInventory && current.InventoryID.String() != input.InventoryID.String() {
		return customfield.Definition{}, apperrors.ErrNotFound
	}

	updated := current
	changedFields := map[string]string{}
	if input.DisplayName != nil {
		displayName, ok := customfield.NewDisplayName(*input.DisplayName)
		if !ok {
			return customfield.Definition{}, apperrors.ErrInvalidInput
		}
		if displayName != current.DisplayName {
			updated.DisplayName = displayName
			changedFields["display_name"] = "true"
		}
	}
	if input.Key != nil {
		key, ok := customfield.NewKey(*input.Key)
		if !ok || key != current.Key {
			return customfield.Definition{}, apperrors.ErrInvalidInput
		}
	}
	if input.Type != nil {
		fieldType, ok := customfield.NewFieldType(*input.Type)
		if !ok || fieldType != current.Type {
			return customfield.Definition{}, apperrors.ErrInvalidInput
		}
	}
	if input.EnumOptions != nil {
		enumOptions, ok := customFieldEnumOptions(*input.EnumOptions)
		if !ok {
			return customfield.Definition{}, apperrors.ErrInvalidInput
		}
		updated.EnumOptions = enumOptions
	}
	if input.Applicability != nil {
		applicability, ok := customfield.NewApplicability(*input.Applicability)
		if !ok {
			return customfield.Definition{}, apperrors.ErrInvalidInput
		}
		updated.Applicability = applicability
		if applicability == customfield.ApplicabilityAllAssets {
			updated.CustomAssetTypeIDs = nil
		}
	}
	if input.CustomAssetTypeIDs != nil {
		targetIDs, err := customFieldTargetIDs(*input.CustomAssetTypeIDs)
		if err != nil {
			return customfield.Definition{}, err
		}
		updated.CustomAssetTypeIDs = targetIDs
	}
	updated, ok = customfield.NewDefinitionWithLifecycle(
		updated.ID,
		updated.TenantID,
		updated.InventoryID,
		updated.Scope,
		updated.Key,
		updated.DisplayName,
		updated.Type,
		updated.EnumOptions,
		updated.Applicability,
		updated.CustomAssetTypeIDs,
		updated.LifecycleState,
	)
	if !ok {
		return customfield.Definition{}, apperrors.ErrInvalidInput
	}
	if !current.IsActive() {
		return customfield.Definition{}, apperrors.ErrInvalidInput
	}
	schemaChange, ok := current.CompatibleSchemaChange(updated)
	if !ok {
		return customfield.Definition{}, apperrors.ErrInvalidInput
	}
	if len(schemaChange.AddedCustomAssetTypeIDs) != 0 {
		if err := s.validateCustomFieldTargetIDs(ctx, input.TenantID, input.InventoryID, scope, schemaChange.AddedCustomAssetTypeIDs); err != nil {
			return customfield.Definition{}, err
		}
		changedFields["custom_asset_type_targets_added"] = strconv.Itoa(len(schemaChange.AddedCustomAssetTypeIDs))
	}
	if len(schemaChange.AddedEnumOptions) != 0 {
		changedFields["enum_options_added"] = strconv.Itoa(len(schemaChange.AddedEnumOptions))
	}
	if schemaChange.ExpandedToAllAssets {
		changedFields["applicability"] = customfield.ApplicabilityAllAssets.String()
	}
	if len(changedFields) == 0 {
		return current, nil
	}

	auditRecord, err := s.newAuditRecord(appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionCustomFieldDefinitionUpdated,
		TargetType:  audit.TargetCustomFieldDefinition,
		TargetID:    updated.ID.String(),
		Metadata: map[string]string{
			"field_key": updated.Key.String(),
			"scope":     updated.Scope.String(),
		},
	})
	if err != nil {
		return customfield.Definition{}, err
	}
	for key, value := range changedFields {
		auditRecord.Metadata[key] = value
	}

	if err := s.customFieldUnitOfWork.UpdateCustomFieldDefinition(ctx, updated, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) || errors.Is(err, ports.ErrForbidden) {
			return customfield.Definition{}, apperrors.ErrInvalidInput
		}
		return customfield.Definition{}, err
	}

	s.observer.Record(ctx, ports.Event{
		Name:    ports.EventCustomFieldDefinitionUpdated,
		Message: "custom field definition updated",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"principal_id":  input.Principal.ID.String(),
			"definition_id": updated.ID.String(),
			"field_key":     updated.Key.String(),
			"scope":         updated.Scope.String(),
		},
	})

	return updated, nil
}

func (s Service) ListTenantCustomFieldDefinitions(ctx context.Context, input ListCustomFieldDefinitionsInput) (ListCustomFieldDefinitionsResult, error) {
	if err := s.ensureTenantExists(ctx, input.TenantID); err != nil {
		return ListCustomFieldDefinitionsResult{}, err
	}
	if err := s.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		s.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return ListCustomFieldDefinitionsResult{}, err
	}

	limit := appsupport.PageLimit(s.defaultPageLimit, s.maxPageLimit, input.Limit)
	afterDefinitionKey, err := decodeCustomFieldDefinitionCursor(input.TenantID, input.InventoryID, input.Cursor)
	if err != nil {
		return ListCustomFieldDefinitionsResult{}, apperrors.ErrInvalidInput
	}
	items, err := s.customFields.ListTenantCustomFieldDefinitions(ctx, input.TenantID, ports.CustomFieldDefinitionPageRequest{
		AfterDefinitionKey: afterDefinitionKey,
		Limit:              limit + 1,
	})
	if err != nil {
		return ListCustomFieldDefinitionsResult{}, err
	}

	return s.customFieldDefinitionListResult(ctx, input, items, limit)
}

func (s Service) ListInventoryCustomFieldDefinitions(ctx context.Context, input ListCustomFieldDefinitionsInput) (ListCustomFieldDefinitionsResult, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return ListCustomFieldDefinitionsResult{}, err
	}

	limit := appsupport.PageLimit(s.defaultPageLimit, s.maxPageLimit, input.Limit)
	afterDefinitionKey, err := decodeCustomFieldDefinitionCursor(input.TenantID, input.InventoryID, input.Cursor)
	if err != nil {
		return ListCustomFieldDefinitionsResult{}, apperrors.ErrInvalidInput
	}
	items, err := s.customFields.ListInventoryCustomFieldDefinitions(ctx, input.TenantID, input.InventoryID, ports.CustomFieldDefinitionPageRequest{
		AfterDefinitionKey: afterDefinitionKey,
		Limit:              limit + 1,
	})
	if err != nil {
		return ListCustomFieldDefinitionsResult{}, err
	}

	return s.customFieldDefinitionListResult(ctx, input, items, limit)
}

func (s Service) customFieldDefinitionListResult(ctx context.Context, input ListCustomFieldDefinitionsInput, items []customfield.Definition, limit int) (ListCustomFieldDefinitionsResult, error) {
	hasMore := len(items) > limit
	var nextCursor *string
	if hasMore {
		items = items[:limit]
		nextCursor = encodeCustomFieldDefinitionCursor(input.TenantID, input.InventoryID, items[len(items)-1].CursorKey())
	}

	s.observer.Record(ctx, ports.Event{
		Name:    ports.EventCustomFieldDefinitionsListed,
		Message: "custom field definitions listed",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"principal_id": input.Principal.ID.String(),
			"limit":        strconv.Itoa(limit),
		},
	})
	if err := s.saveReadAuditRecord(ctx, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionCustomFieldDefinitionListed,
		TargetType:  audit.TargetTenant,
		TargetID:    input.TenantID.String(),
		Metadata: map[string]string{
			"limit": strconv.Itoa(limit),
		},
	}); err != nil {
		return ListCustomFieldDefinitionsResult{}, err
	}

	return ListCustomFieldDefinitionsResult{
		Items:      items,
		Limit:      limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
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

func updateCustomFieldDefinitionInputIsEmpty(input UpdateCustomFieldDefinitionInput) bool {
	return input.DisplayName == nil &&
		input.Key == nil &&
		input.Type == nil &&
		input.EnumOptions == nil &&
		input.Applicability == nil &&
		input.CustomAssetTypeIDs == nil
}

func customFieldTargetIDs(values []string) ([]customfield.AssetTypeID, error) {
	targetIDs := make([]customfield.AssetTypeID, 0, len(values))
	seen := map[customfield.AssetTypeID]struct{}{}
	for _, raw := range values {
		id, ok := customfield.NewAssetTypeID(raw)
		if !ok {
			return nil, apperrors.ErrInvalidInput
		}
		if _, exists := seen[id]; exists {
			return nil, apperrors.ErrInvalidInput
		}
		seen[id] = struct{}{}
		targetIDs = append(targetIDs, id)
	}
	return targetIDs, nil
}

func (s Service) validatedCustomFieldTargetIDs(ctx context.Context, input CreateCustomFieldDefinitionInput, scope customfield.Scope, applicability customfield.Applicability) ([]customfield.AssetTypeID, error) {
	if applicability == customfield.ApplicabilityAllAssets {
		if len(input.CustomAssetTypeIDs) != 0 {
			return nil, apperrors.ErrInvalidInput
		}
		return nil, nil
	}
	targetIDs, err := customFieldTargetIDs(input.CustomAssetTypeIDs)
	if err != nil {
		return nil, err
	}
	if len(targetIDs) == 0 {
		return nil, apperrors.ErrInvalidInput
	}
	if err := s.validateCustomFieldTargetIDs(ctx, input.TenantID, input.InventoryID, scope, targetIDs); err != nil {
		return nil, err
	}
	return targetIDs, nil
}

func (s Service) validateCustomFieldTargetIDs(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, scope customfield.Scope, targetIDs []customfield.AssetTypeID) error {
	if s.customAssetTypes == nil {
		return apperrors.ErrInvalidInput
	}
	targets, err := s.customAssetTypes.CustomAssetTypesByID(ctx, tenantID, inventoryID, targetIDs)
	if err != nil {
		return err
	}
	if len(targets) != len(targetIDs) {
		return apperrors.ErrNotFound
	}
	targetByID := map[customfield.AssetTypeID]customfield.AssetType{}
	for _, target := range targets {
		targetByID[target.ID] = target
	}
	for _, id := range targetIDs {
		target, found := targetByID[id]
		if !found {
			return apperrors.ErrNotFound
		}
		if target.TenantID.String() != tenantID.String() {
			return apperrors.ErrNotFound
		}
		if scope == customfield.ScopeInventory && target.Scope == customfield.ScopeInventory && target.InventoryID.String() != inventoryID.String() {
			return apperrors.ErrNotFound
		}
		if scope == customfield.ScopeTenant && target.Scope != customfield.ScopeTenant {
			return apperrors.ErrInvalidInput
		}
	}
	return nil
}

func encodeCustomFieldDefinitionCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, key string) *string {
	return appsupport.EncodePageCursor("custom_field_definitions", customFieldDefinitionCursorScope(tenantID, inventoryID), key)
}

func decodeCustomFieldDefinitionCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, cursor string) (string, error) {
	decoded, err := appsupport.DecodePageCursor("custom_field_definitions", customFieldDefinitionCursorScope(tenantID, inventoryID), cursor)
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
