package customfields

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s Service) GetTenantCustomFieldDefinition(ctx context.Context, input GetCustomFieldDefinitionInput) (customfield.Definition, error) {
	if err := s.ensureTenantExists(ctx, input.TenantID); err != nil {
		return customfield.Definition{}, err
	}
	if err := s.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		s.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return customfield.Definition{}, err
	}
	return s.getCustomFieldDefinition(ctx, input, customfield.ScopeTenant)
}

func (s Service) GetInventoryCustomFieldDefinition(ctx context.Context, input GetCustomFieldDefinitionInput) (customfield.Definition, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return customfield.Definition{}, err
	}
	return s.getCustomFieldDefinition(ctx, input, customfield.ScopeInventory)
}

func (s Service) getCustomFieldDefinition(ctx context.Context, input GetCustomFieldDefinitionInput, scope customfield.Scope) (customfield.Definition, error) {
	definitionID, ok := customfield.NewID(input.DefinitionID.String())
	if !ok {
		return customfield.Definition{}, apperrors.ErrInvalidInput
	}
	definition, found, err := s.customFields.CustomFieldDefinitionByID(ctx, input.TenantID, input.InventoryID, definitionID)
	if err != nil {
		return customfield.Definition{}, err
	}
	if !found || definition.Scope != scope {
		return customfield.Definition{}, apperrors.ErrNotFound
	}
	if scope == customfield.ScopeInventory && definition.InventoryID.String() != input.InventoryID.String() {
		return customfield.Definition{}, apperrors.ErrNotFound
	}
	if err := s.saveReadAuditRecord(ctx, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionCustomFieldDefinitionViewed,
		TargetType:  audit.TargetCustomFieldDefinition,
		TargetID:    definition.ID.String(),
		Metadata: map[string]string{
			"field_key":       definition.Key.String(),
			"scope":           definition.Scope.String(),
			"lifecycle_state": definition.LifecycleState.String(),
		},
	}); err != nil {
		return customfield.Definition{}, err
	}
	s.observer.Record(ctx, ports.Event{
		Name:    ports.EventCustomFieldDefinitionViewed,
		Message: "custom field definition viewed",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"principal_id":  input.Principal.ID.String(),
			"definition_id": definition.ID.String(),
			"scope":         definition.Scope.String(),
		},
	})
	return definition, nil
}

func (s Service) ArchiveTenantCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput) (customfield.Definition, error) {
	if err := s.ensureTenantExists(ctx, input.TenantID); err != nil {
		return customfield.Definition{}, err
	}
	if err := s.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		s.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return customfield.Definition{}, err
	}
	return s.updateCustomFieldDefinitionLifecycle(ctx, input, customfield.ScopeTenant, customfield.DefinitionLifecycleActive, customfield.DefinitionLifecycleArchived, audit.ActionCustomFieldDefinitionArchived, ports.EventCustomFieldDefinitionArchived, "custom field definition archived")
}

func (s Service) ArchiveInventoryCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput) (customfield.Definition, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return customfield.Definition{}, err
	}
	return s.updateCustomFieldDefinitionLifecycle(ctx, input, customfield.ScopeInventory, customfield.DefinitionLifecycleActive, customfield.DefinitionLifecycleArchived, audit.ActionCustomFieldDefinitionArchived, ports.EventCustomFieldDefinitionArchived, "custom field definition archived")
}

func (s Service) RestoreTenantCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput) (customfield.Definition, error) {
	if err := s.ensureTenantExists(ctx, input.TenantID); err != nil {
		return customfield.Definition{}, err
	}
	if err := s.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		s.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return customfield.Definition{}, err
	}
	return s.updateCustomFieldDefinitionLifecycle(ctx, input, customfield.ScopeTenant, customfield.DefinitionLifecycleArchived, customfield.DefinitionLifecycleActive, audit.ActionCustomFieldDefinitionRestored, ports.EventCustomFieldDefinitionRestored, "custom field definition restored")
}

func (s Service) RestoreInventoryCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput) (customfield.Definition, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return customfield.Definition{}, err
	}
	return s.updateCustomFieldDefinitionLifecycle(ctx, input, customfield.ScopeInventory, customfield.DefinitionLifecycleArchived, customfield.DefinitionLifecycleActive, audit.ActionCustomFieldDefinitionRestored, ports.EventCustomFieldDefinitionRestored, "custom field definition restored")
}

func (s Service) updateCustomFieldDefinitionLifecycle(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput, scope customfield.Scope, from customfield.DefinitionLifecycleState, to customfield.DefinitionLifecycleState, action audit.Action, eventName ports.EventName, eventMessage string) (customfield.Definition, error) {
	definitionID, ok := customfield.NewID(input.DefinitionID.String())
	if !ok {
		return customfield.Definition{}, apperrors.ErrInvalidInput
	}
	current, found, err := s.customFields.CustomFieldDefinitionByID(ctx, input.TenantID, input.InventoryID, definitionID)
	if err != nil {
		return customfield.Definition{}, err
	}
	if !found || current.Scope != scope {
		return customfield.Definition{}, apperrors.ErrNotFound
	}
	if scope == customfield.ScopeInventory && current.InventoryID.String() != input.InventoryID.String() {
		return customfield.Definition{}, apperrors.ErrNotFound
	}
	if current.LifecycleState != from {
		return customfield.Definition{}, apperrors.ErrInvalidInput
	}
	if to == customfield.DefinitionLifecycleActive {
		if err := s.validateCustomFieldTargetIDs(ctx, input.TenantID, input.InventoryID, scope, current.CustomAssetTypeIDs); err != nil {
			return customfield.Definition{}, err
		}
	}
	updated := current
	updated.LifecycleState = to
	auditRecord, err := s.newAuditRecord(appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      action,
		TargetType:  audit.TargetCustomFieldDefinition,
		TargetID:    updated.ID.String(),
		Metadata: map[string]string{
			"field_key":       updated.Key.String(),
			"scope":           updated.Scope.String(),
			"previous_state":  current.LifecycleState.String(),
			"lifecycle_state": updated.LifecycleState.String(),
		},
	})
	if err != nil {
		return customfield.Definition{}, err
	}
	if err := s.customFieldUnitOfWork.UpdateCustomFieldDefinitionLifecycle(ctx, updated, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) || errors.Is(err, ports.ErrForbidden) {
			return customfield.Definition{}, apperrors.ErrInvalidInput
		}
		return customfield.Definition{}, err
	}
	s.observer.Record(ctx, ports.Event{
		Name:    eventName,
		Message: eventMessage,
		Fields: map[string]string{
			"tenant_id":       input.TenantID.String(),
			"inventory_id":    input.InventoryID.String(),
			"principal_id":    input.Principal.ID.String(),
			"definition_id":   updated.ID.String(),
			"scope":           updated.Scope.String(),
			"lifecycle_state": updated.LifecycleState.String(),
		},
	})
	return updated, nil
}

func (s Service) DeleteTenantCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput) error {
	if err := s.ensureTenantExists(ctx, input.TenantID); err != nil {
		return err
	}
	if err := s.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		s.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return err
	}
	return s.deleteCustomFieldDefinition(ctx, input, customfield.ScopeTenant)
}

func (s Service) DeleteInventoryCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput) error {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return err
	}
	return s.deleteCustomFieldDefinition(ctx, input, customfield.ScopeInventory)
}

func (s Service) deleteCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput, scope customfield.Scope) error {
	definitionID, ok := customfield.NewID(input.DefinitionID.String())
	if !ok {
		return apperrors.ErrInvalidInput
	}
	current, found, err := s.customFields.CustomFieldDefinitionByID(ctx, input.TenantID, input.InventoryID, definitionID)
	if err != nil {
		return err
	}
	if !found || current.Scope != scope {
		return apperrors.ErrNotFound
	}
	if scope == customfield.ScopeInventory && current.InventoryID.String() != input.InventoryID.String() {
		return apperrors.ErrNotFound
	}
	hasValues, err := s.customFields.CustomFieldDefinitionHasActiveAssetValues(ctx, input.TenantID, input.InventoryID, current)
	if err != nil {
		return err
	}
	if hasValues {
		return apperrors.ErrInvalidInput
	}
	auditRecord, err := s.newAuditRecord(appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionCustomFieldDefinitionDeleted,
		TargetType:  audit.TargetCustomFieldDefinition,
		TargetID:    current.ID.String(),
		Metadata: map[string]string{
			"field_key":       current.Key.String(),
			"scope":           current.Scope.String(),
			"lifecycle_state": current.LifecycleState.String(),
		},
	})
	if err != nil {
		return err
	}
	if err := s.customFieldUnitOfWork.DeleteCustomFieldDefinition(ctx, input.TenantID, input.InventoryID, definitionID, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) || errors.Is(err, ports.ErrForbidden) {
			return apperrors.ErrInvalidInput
		}
		return err
	}
	s.observer.Record(ctx, ports.Event{
		Name:    ports.EventCustomFieldDefinitionDeleted,
		Message: "custom field definition deleted",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"principal_id":  input.Principal.ID.String(),
			"definition_id": current.ID.String(),
			"scope":         current.Scope.String(),
		},
	})
	return nil
}
