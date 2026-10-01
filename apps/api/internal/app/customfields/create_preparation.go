package customfields

import (
	"context"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type PreparedCustomFieldDefinition = ports.PreparedCustomFieldDefinition

func (s Service) PrepareInventoryCustomFieldDefinition(ctx context.Context, input CreateCustomFieldDefinitionInput) (PreparedCustomFieldDefinition, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return PreparedCustomFieldDefinition{}, err
	}
	return s.prepareCustomFieldDefinition(ctx, input, customfield.ScopeInventory)
}

func (s Service) prepareCustomFieldDefinition(ctx context.Context, input CreateCustomFieldDefinitionInput, scope customfield.Scope) (PreparedCustomFieldDefinition, error) {
	id, ok := customfield.NewID(s.ids.NewID())
	if !ok {
		return PreparedCustomFieldDefinition{}, apperrors.ErrInvalidInput
	}
	key, ok := customfield.NewKey(input.Key)
	if !ok {
		return PreparedCustomFieldDefinition{}, apperrors.ErrInvalidInput
	}
	displayName, ok := customfield.NewDisplayName(input.DisplayName)
	if !ok {
		return PreparedCustomFieldDefinition{}, apperrors.ErrInvalidInput
	}
	fieldType, ok := customfield.NewFieldType(input.Type)
	if !ok {
		return PreparedCustomFieldDefinition{}, apperrors.ErrInvalidInput
	}
	enumOptions, ok := customFieldEnumOptions(input.EnumOptions)
	if !ok {
		return PreparedCustomFieldDefinition{}, apperrors.ErrInvalidInput
	}
	applicability, ok := customfield.NewApplicability(input.Applicability)
	if !ok {
		return PreparedCustomFieldDefinition{}, apperrors.ErrInvalidInput
	}
	customAssetTypeIDs, err := s.validatedCustomFieldTargetIDs(ctx, input, scope, applicability)
	if err != nil {
		return PreparedCustomFieldDefinition{}, err
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
		return PreparedCustomFieldDefinition{}, apperrors.ErrInvalidInput
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
		return PreparedCustomFieldDefinition{}, err
	}

	return PreparedCustomFieldDefinition{Item: definition, AuditRecord: auditRecord}, nil
}

func (s Service) RecordCustomFieldDefinitionCreated(ctx context.Context, input CreateCustomFieldDefinitionInput, definition customfield.Definition) {
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

}

type PreparedCustomAssetType = ports.PreparedCustomAssetType

func (s Service) PrepareInventoryCustomAssetType(ctx context.Context, input CreateCustomAssetTypeInput) (PreparedCustomAssetType, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return PreparedCustomAssetType{}, err
	}
	return s.prepareCustomAssetType(ctx, input, customfield.ScopeInventory)
}

func (s Service) prepareCustomAssetType(ctx context.Context, input CreateCustomAssetTypeInput, scope customfield.Scope) (PreparedCustomAssetType, error) {
	id, ok := customfield.NewAssetTypeID(s.ids.NewID())
	if !ok {
		return PreparedCustomAssetType{}, apperrors.ErrInvalidInput
	}
	key, ok := customfield.NewKey(input.Key)
	if !ok {
		return PreparedCustomAssetType{}, apperrors.ErrInvalidInput
	}
	displayName, ok := customfield.NewDisplayName(input.DisplayName)
	if !ok {
		return PreparedCustomAssetType{}, apperrors.ErrInvalidInput
	}
	description, ok := customfield.NewDescription(input.Description)
	if !ok {
		return PreparedCustomAssetType{}, apperrors.ErrInvalidInput
	}

	inventoryID := customfield.InventoryID("")
	if scope == customfield.ScopeInventory {
		inventoryID = customfield.InventoryID(input.InventoryID.String())
	}
	assetType, ok := customfield.NewAssetType(
		id,
		customfield.TenantID(input.TenantID.String()),
		inventoryID,
		scope,
		key,
		displayName,
		description,
	)
	if !ok {
		return PreparedCustomAssetType{}, apperrors.ErrInvalidInput
	}

	assetType.ExpirationEnabled = input.ExpirationEnabled

	auditRecord, err := s.newAuditRecord(appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionCustomAssetTypeCreated,
		TargetType:  audit.TargetCustomAssetType,
		TargetID:    assetType.ID.String(),
		Metadata: map[string]string{
			"type_key": assetType.Key.String(),
			"scope":    assetType.Scope.String(),
		},
	})
	if err != nil {
		return PreparedCustomAssetType{}, err
	}

	return PreparedCustomAssetType{Item: assetType, AuditRecord: auditRecord}, nil
}

func (s Service) RecordCustomAssetTypeCreated(ctx context.Context, input CreateCustomAssetTypeInput, assetType customfield.AssetType) {
	s.observer.Record(ctx, ports.Event{
		Name:    ports.EventCustomAssetTypeCreated,
		Message: "custom asset type created",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"principal_id":  input.Principal.ID.String(),
			"asset_type_id": assetType.ID.String(),
			"type_key":      assetType.Key.String(),
			"scope":         assetType.Scope.String(),
		},
	})

}
