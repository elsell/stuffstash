package app

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) GetTenantCustomFieldDefinition(ctx context.Context, input GetCustomFieldDefinitionInput) (customfield.Definition, error) {
	if err := a.ensureTenantExists(ctx, input.TenantID); err != nil {
		return customfield.Definition{}, err
	}
	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return customfield.Definition{}, err
	}
	return a.getCustomFieldDefinition(ctx, input, customfield.ScopeTenant)
}

func (a App) GetInventoryCustomFieldDefinition(ctx context.Context, input GetCustomFieldDefinitionInput) (customfield.Definition, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return customfield.Definition{}, err
	}
	return a.getCustomFieldDefinition(ctx, input, customfield.ScopeInventory)
}

func (a App) getCustomFieldDefinition(ctx context.Context, input GetCustomFieldDefinitionInput, scope customfield.Scope) (customfield.Definition, error) {
	definitionID, ok := customfield.NewID(input.DefinitionID.String())
	if !ok {
		return customfield.Definition{}, ErrInvalidInput
	}
	definition, found, err := a.customFields.CustomFieldDefinitionByID(ctx, input.TenantID, input.InventoryID, definitionID)
	if err != nil {
		return customfield.Definition{}, err
	}
	if !found || definition.Scope != scope {
		return customfield.Definition{}, ErrNotFound
	}
	if scope == customfield.ScopeInventory && definition.InventoryID.String() != input.InventoryID.String() {
		return customfield.Definition{}, ErrNotFound
	}
	if err := a.saveReadAuditRecord(ctx, auditRecordInput{
		PrincipalID: input.Principal.ID,
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
	a.observer.Record(ctx, ports.Event{
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

func (a App) ArchiveTenantCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput) (customfield.Definition, error) {
	if err := a.ensureTenantExists(ctx, input.TenantID); err != nil {
		return customfield.Definition{}, err
	}
	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return customfield.Definition{}, err
	}
	return a.updateCustomFieldDefinitionLifecycle(ctx, input, customfield.ScopeTenant, customfield.DefinitionLifecycleActive, customfield.DefinitionLifecycleArchived, audit.ActionCustomFieldDefinitionArchived, ports.EventCustomFieldDefinitionArchived, "custom field definition archived")
}

func (a App) ArchiveInventoryCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput) (customfield.Definition, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return customfield.Definition{}, err
	}
	return a.updateCustomFieldDefinitionLifecycle(ctx, input, customfield.ScopeInventory, customfield.DefinitionLifecycleActive, customfield.DefinitionLifecycleArchived, audit.ActionCustomFieldDefinitionArchived, ports.EventCustomFieldDefinitionArchived, "custom field definition archived")
}

func (a App) RestoreTenantCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput) (customfield.Definition, error) {
	if err := a.ensureTenantExists(ctx, input.TenantID); err != nil {
		return customfield.Definition{}, err
	}
	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return customfield.Definition{}, err
	}
	return a.updateCustomFieldDefinitionLifecycle(ctx, input, customfield.ScopeTenant, customfield.DefinitionLifecycleArchived, customfield.DefinitionLifecycleActive, audit.ActionCustomFieldDefinitionRestored, ports.EventCustomFieldDefinitionRestored, "custom field definition restored")
}

func (a App) RestoreInventoryCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput) (customfield.Definition, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return customfield.Definition{}, err
	}
	return a.updateCustomFieldDefinitionLifecycle(ctx, input, customfield.ScopeInventory, customfield.DefinitionLifecycleArchived, customfield.DefinitionLifecycleActive, audit.ActionCustomFieldDefinitionRestored, ports.EventCustomFieldDefinitionRestored, "custom field definition restored")
}

func (a App) updateCustomFieldDefinitionLifecycle(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput, scope customfield.Scope, from customfield.DefinitionLifecycleState, to customfield.DefinitionLifecycleState, action audit.Action, eventName ports.EventName, eventMessage string) (customfield.Definition, error) {
	definitionID, ok := customfield.NewID(input.DefinitionID.String())
	if !ok {
		return customfield.Definition{}, ErrInvalidInput
	}
	current, found, err := a.customFields.CustomFieldDefinitionByID(ctx, input.TenantID, input.InventoryID, definitionID)
	if err != nil {
		return customfield.Definition{}, err
	}
	if !found || current.Scope != scope {
		return customfield.Definition{}, ErrNotFound
	}
	if scope == customfield.ScopeInventory && current.InventoryID.String() != input.InventoryID.String() {
		return customfield.Definition{}, ErrNotFound
	}
	if current.LifecycleState != from {
		return customfield.Definition{}, ErrInvalidInput
	}
	if to == customfield.DefinitionLifecycleActive {
		if err := a.validateCustomFieldTargetIDs(ctx, input.TenantID, input.InventoryID, scope, current.CustomAssetTypeIDs); err != nil {
			return customfield.Definition{}, err
		}
	}
	updated := current
	updated.LifecycleState = to
	auditRecord, err := a.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
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
	if err := a.customFields.UpdateCustomFieldDefinitionLifecycle(ctx, updated, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) || errors.Is(err, ports.ErrForbidden) {
			return customfield.Definition{}, ErrInvalidInput
		}
		return customfield.Definition{}, err
	}
	a.observer.Record(ctx, ports.Event{
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

func (a App) DeleteTenantCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput) error {
	if err := a.ensureTenantExists(ctx, input.TenantID); err != nil {
		return err
	}
	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return err
	}
	return a.deleteCustomFieldDefinition(ctx, input, customfield.ScopeTenant)
}

func (a App) DeleteInventoryCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput) error {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return err
	}
	return a.deleteCustomFieldDefinition(ctx, input, customfield.ScopeInventory)
}

func (a App) deleteCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput, scope customfield.Scope) error {
	definitionID, ok := customfield.NewID(input.DefinitionID.String())
	if !ok {
		return ErrInvalidInput
	}
	current, found, err := a.customFields.CustomFieldDefinitionByID(ctx, input.TenantID, input.InventoryID, definitionID)
	if err != nil {
		return err
	}
	if !found || current.Scope != scope {
		return ErrNotFound
	}
	if scope == customfield.ScopeInventory && current.InventoryID.String() != input.InventoryID.String() {
		return ErrNotFound
	}
	hasValues, err := a.customFields.CustomFieldDefinitionHasActiveAssetValues(ctx, input.TenantID, input.InventoryID, current)
	if err != nil {
		return err
	}
	if hasValues {
		return ErrInvalidInput
	}
	auditRecord, err := a.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
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
	if err := a.customFields.DeleteCustomFieldDefinition(ctx, input.TenantID, input.InventoryID, definitionID, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) || errors.Is(err, ports.ErrForbidden) {
			return ErrInvalidInput
		}
		return err
	}
	a.observer.Record(ctx, ports.Event{
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
