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

type CreateCustomAssetTypeInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Key         string
	DisplayName string
	Description string
}

type ListCustomAssetTypesInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Limit       int
	Cursor      string
}

type GetCustomAssetTypeInput struct {
	Principal         identity.Principal
	Source            audit.Source
	RequestID         string
	TenantID          tenant.ID
	InventoryID       inventory.InventoryID
	CustomAssetTypeID customfield.AssetTypeID
}

type UpdateCustomAssetTypeInput struct {
	Principal         identity.Principal
	Source            audit.Source
	RequestID         string
	TenantID          tenant.ID
	InventoryID       inventory.InventoryID
	CustomAssetTypeID customfield.AssetTypeID
	DisplayName       *string
	Description       *string
}

type ArchiveCustomAssetTypeInput struct {
	Principal         identity.Principal
	Source            audit.Source
	RequestID         string
	TenantID          tenant.ID
	InventoryID       inventory.InventoryID
	CustomAssetTypeID customfield.AssetTypeID
}

type ListCustomAssetTypesResult struct {
	Items      []customfield.AssetType
	Limit      int
	NextCursor *string
	HasMore    bool
}

func (s Service) CreateTenantCustomAssetType(ctx context.Context, input CreateCustomAssetTypeInput) (customfield.AssetType, error) {
	if err := s.ensureTenantExists(ctx, input.TenantID); err != nil {
		return customfield.AssetType{}, err
	}
	if err := s.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		s.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return customfield.AssetType{}, err
	}
	return s.createCustomAssetType(ctx, input, customfield.ScopeTenant)
}

func (s Service) CreateInventoryCustomAssetType(ctx context.Context, input CreateCustomAssetTypeInput) (customfield.AssetType, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return customfield.AssetType{}, err
	}
	return s.createCustomAssetType(ctx, input, customfield.ScopeInventory)
}

func (s Service) UpdateTenantCustomAssetType(ctx context.Context, input UpdateCustomAssetTypeInput) (customfield.AssetType, error) {
	if err := s.ensureTenantExists(ctx, input.TenantID); err != nil {
		return customfield.AssetType{}, err
	}
	if err := s.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		s.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return customfield.AssetType{}, err
	}
	return s.updateCustomAssetType(ctx, input, customfield.ScopeTenant)
}

func (s Service) UpdateInventoryCustomAssetType(ctx context.Context, input UpdateCustomAssetTypeInput) (customfield.AssetType, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return customfield.AssetType{}, err
	}
	return s.updateCustomAssetType(ctx, input, customfield.ScopeInventory)
}

func (s Service) ArchiveTenantCustomAssetType(ctx context.Context, input ArchiveCustomAssetTypeInput) (customfield.AssetType, error) {
	if err := s.ensureTenantExists(ctx, input.TenantID); err != nil {
		return customfield.AssetType{}, err
	}
	if err := s.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		s.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return customfield.AssetType{}, err
	}
	return s.archiveCustomAssetType(ctx, input, customfield.ScopeTenant)
}

func (s Service) ArchiveInventoryCustomAssetType(ctx context.Context, input ArchiveCustomAssetTypeInput) (customfield.AssetType, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return customfield.AssetType{}, err
	}
	return s.archiveCustomAssetType(ctx, input, customfield.ScopeInventory)
}

func (s Service) GetTenantCustomAssetType(ctx context.Context, input GetCustomAssetTypeInput) (customfield.AssetType, error) {
	if err := s.ensureTenantExists(ctx, input.TenantID); err != nil {
		return customfield.AssetType{}, err
	}
	if err := s.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		s.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return customfield.AssetType{}, err
	}
	return s.getCustomAssetType(ctx, input, customfield.ScopeTenant)
}

func (s Service) GetInventoryCustomAssetType(ctx context.Context, input GetCustomAssetTypeInput) (customfield.AssetType, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return customfield.AssetType{}, err
	}
	return s.getCustomAssetType(ctx, input, customfield.ScopeInventory)
}

func (s Service) getCustomAssetType(ctx context.Context, input GetCustomAssetTypeInput, scope customfield.Scope) (customfield.AssetType, error) {
	assetTypeID, ok := customfield.NewAssetTypeID(input.CustomAssetTypeID.String())
	if !ok {
		return customfield.AssetType{}, apperrors.ErrInvalidInput
	}
	item, found, err := s.customAssetTypes.CustomAssetTypeByID(ctx, input.TenantID, input.InventoryID, assetTypeID)
	if err != nil {
		return customfield.AssetType{}, err
	}
	if !found || item.Scope != scope {
		return customfield.AssetType{}, apperrors.ErrNotFound
	}
	if scope == customfield.ScopeInventory && item.InventoryID.String() != input.InventoryID.String() {
		return customfield.AssetType{}, apperrors.ErrNotFound
	}
	if err := s.saveReadAuditRecord(ctx, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionCustomAssetTypeViewed,
		TargetType:  audit.TargetCustomAssetType,
		TargetID:    item.ID.String(),
		Metadata: map[string]string{
			"type_key":        item.Key.String(),
			"scope":           item.Scope.String(),
			"lifecycle_state": item.LifecycleState.String(),
		},
	}); err != nil {
		return customfield.AssetType{}, err
	}
	s.observer.Record(ctx, ports.Event{
		Name:    ports.EventCustomAssetTypeViewed,
		Message: "custom asset type viewed",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"principal_id":  input.Principal.ID.String(),
			"asset_type_id": item.ID.String(),
			"scope":         item.Scope.String(),
		},
	})
	return item, nil
}

func (s Service) RestoreTenantCustomAssetType(ctx context.Context, input ArchiveCustomAssetTypeInput) (customfield.AssetType, error) {
	if err := s.ensureTenantExists(ctx, input.TenantID); err != nil {
		return customfield.AssetType{}, err
	}
	if err := s.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		s.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return customfield.AssetType{}, err
	}
	return s.restoreCustomAssetType(ctx, input, customfield.ScopeTenant)
}

func (s Service) RestoreInventoryCustomAssetType(ctx context.Context, input ArchiveCustomAssetTypeInput) (customfield.AssetType, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return customfield.AssetType{}, err
	}
	return s.restoreCustomAssetType(ctx, input, customfield.ScopeInventory)
}

func (s Service) DeleteTenantCustomAssetType(ctx context.Context, input ArchiveCustomAssetTypeInput) error {
	if err := s.ensureTenantExists(ctx, input.TenantID); err != nil {
		return err
	}
	if err := s.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		s.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return err
	}
	return s.deleteCustomAssetType(ctx, input, customfield.ScopeTenant)
}

func (s Service) DeleteInventoryCustomAssetType(ctx context.Context, input ArchiveCustomAssetTypeInput) error {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return err
	}
	return s.deleteCustomAssetType(ctx, input, customfield.ScopeInventory)
}

func (s Service) createCustomAssetType(ctx context.Context, input CreateCustomAssetTypeInput, scope customfield.Scope) (customfield.AssetType, error) {
	id, ok := customfield.NewAssetTypeID(s.ids.NewID())
	if !ok {
		return customfield.AssetType{}, apperrors.ErrInvalidInput
	}
	key, ok := customfield.NewKey(input.Key)
	if !ok {
		return customfield.AssetType{}, apperrors.ErrInvalidInput
	}
	displayName, ok := customfield.NewDisplayName(input.DisplayName)
	if !ok {
		return customfield.AssetType{}, apperrors.ErrInvalidInput
	}
	description, ok := customfield.NewDescription(input.Description)
	if !ok {
		return customfield.AssetType{}, apperrors.ErrInvalidInput
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
		return customfield.AssetType{}, apperrors.ErrInvalidInput
	}

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
		return customfield.AssetType{}, err
	}

	if err := s.customAssetTypeUnitOfWork.SaveCustomAssetType(ctx, assetType, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return customfield.AssetType{}, apperrors.ErrInvalidInput
		}
		return customfield.AssetType{}, err
	}

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

	return assetType, nil
}

func (s Service) updateCustomAssetType(ctx context.Context, input UpdateCustomAssetTypeInput, scope customfield.Scope) (customfield.AssetType, error) {
	assetTypeID, ok := customfield.NewAssetTypeID(input.CustomAssetTypeID.String())
	if !ok || (input.DisplayName == nil && input.Description == nil) {
		return customfield.AssetType{}, apperrors.ErrInvalidInput
	}
	current, found, err := s.customAssetTypes.CustomAssetTypeByID(ctx, input.TenantID, input.InventoryID, assetTypeID)
	if err != nil {
		return customfield.AssetType{}, err
	}
	if !found {
		return customfield.AssetType{}, apperrors.ErrNotFound
	}
	if current.Scope != scope {
		return customfield.AssetType{}, apperrors.ErrNotFound
	}
	if scope == customfield.ScopeInventory && current.InventoryID.String() != input.InventoryID.String() {
		return customfield.AssetType{}, apperrors.ErrNotFound
	}
	if !current.IsActive() {
		return customfield.AssetType{}, apperrors.ErrInvalidInput
	}

	updated := current
	changedFields := map[string]string{}
	if input.DisplayName != nil {
		displayName, ok := customfield.NewDisplayName(*input.DisplayName)
		if !ok {
			return customfield.AssetType{}, apperrors.ErrInvalidInput
		}
		if displayName != current.DisplayName {
			updated.DisplayName = displayName
			changedFields["display_name"] = "true"
		}
	}
	if input.Description != nil {
		description, ok := customfield.NewDescription(*input.Description)
		if !ok {
			return customfield.AssetType{}, apperrors.ErrInvalidInput
		}
		if description != current.Description {
			updated.Description = description
			changedFields["description"] = "true"
		}
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
		Action:      audit.ActionCustomAssetTypeUpdated,
		TargetType:  audit.TargetCustomAssetType,
		TargetID:    updated.ID.String(),
		Metadata: map[string]string{
			"type_key": updated.Key.String(),
			"scope":    updated.Scope.String(),
		},
	})
	if err != nil {
		return customfield.AssetType{}, err
	}
	for key, value := range changedFields {
		auditRecord.Metadata[key] = value
	}

	if err := s.customAssetTypeUnitOfWork.UpdateCustomAssetType(ctx, updated, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) || errors.Is(err, ports.ErrForbidden) {
			return customfield.AssetType{}, apperrors.ErrInvalidInput
		}
		return customfield.AssetType{}, err
	}

	s.observer.Record(ctx, ports.Event{
		Name:    ports.EventCustomAssetTypeUpdated,
		Message: "custom asset type updated",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"principal_id":  input.Principal.ID.String(),
			"asset_type_id": updated.ID.String(),
			"type_key":      updated.Key.String(),
			"scope":         updated.Scope.String(),
		},
	})

	return updated, nil
}

func (s Service) archiveCustomAssetType(ctx context.Context, input ArchiveCustomAssetTypeInput, scope customfield.Scope) (customfield.AssetType, error) {
	assetTypeID, ok := customfield.NewAssetTypeID(input.CustomAssetTypeID.String())
	if !ok {
		return customfield.AssetType{}, apperrors.ErrInvalidInput
	}
	current, found, err := s.customAssetTypes.CustomAssetTypeByID(ctx, input.TenantID, input.InventoryID, assetTypeID)
	if err != nil {
		return customfield.AssetType{}, err
	}
	if !found {
		return customfield.AssetType{}, apperrors.ErrNotFound
	}
	if current.Scope != scope {
		return customfield.AssetType{}, apperrors.ErrNotFound
	}
	if scope == customfield.ScopeInventory && current.InventoryID.String() != input.InventoryID.String() {
		return customfield.AssetType{}, apperrors.ErrNotFound
	}
	archived, ok := current.Archive()
	if !ok {
		return customfield.AssetType{}, apperrors.ErrInvalidInput
	}

	auditRecord, err := s.newAuditRecord(appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionCustomAssetTypeArchived,
		TargetType:  audit.TargetCustomAssetType,
		TargetID:    archived.ID.String(),
		Metadata: map[string]string{
			"type_key": archived.Key.String(),
			"scope":    archived.Scope.String(),
		},
	})
	if err != nil {
		return customfield.AssetType{}, err
	}

	if err := s.customAssetTypeUnitOfWork.ArchiveCustomAssetType(ctx, archived, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) || errors.Is(err, ports.ErrForbidden) {
			return customfield.AssetType{}, apperrors.ErrInvalidInput
		}
		return customfield.AssetType{}, err
	}

	s.observer.Record(ctx, ports.Event{
		Name:    ports.EventCustomAssetTypeArchived,
		Message: "custom asset type archived",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"principal_id":  input.Principal.ID.String(),
			"asset_type_id": archived.ID.String(),
			"type_key":      archived.Key.String(),
			"scope":         archived.Scope.String(),
		},
	})

	return archived, nil
}

func (s Service) restoreCustomAssetType(ctx context.Context, input ArchiveCustomAssetTypeInput, scope customfield.Scope) (customfield.AssetType, error) {
	assetTypeID, ok := customfield.NewAssetTypeID(input.CustomAssetTypeID.String())
	if !ok {
		return customfield.AssetType{}, apperrors.ErrInvalidInput
	}
	current, found, err := s.customAssetTypes.CustomAssetTypeByID(ctx, input.TenantID, input.InventoryID, assetTypeID)
	if err != nil {
		return customfield.AssetType{}, err
	}
	if !found || current.Scope != scope {
		return customfield.AssetType{}, apperrors.ErrNotFound
	}
	if scope == customfield.ScopeInventory && current.InventoryID.String() != input.InventoryID.String() {
		return customfield.AssetType{}, apperrors.ErrNotFound
	}
	if current.LifecycleState != customfield.AssetTypeLifecycleArchived {
		return customfield.AssetType{}, apperrors.ErrInvalidInput
	}
	restored := current
	restored.LifecycleState = customfield.AssetTypeLifecycleActive
	auditRecord, err := s.newAuditRecord(appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionCustomAssetTypeRestored,
		TargetType:  audit.TargetCustomAssetType,
		TargetID:    restored.ID.String(),
		Metadata: map[string]string{
			"type_key": restored.Key.String(),
			"scope":    restored.Scope.String(),
		},
	})
	if err != nil {
		return customfield.AssetType{}, err
	}
	if err := s.customAssetTypeUnitOfWork.RestoreCustomAssetType(ctx, restored, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) || errors.Is(err, ports.ErrForbidden) {
			return customfield.AssetType{}, apperrors.ErrInvalidInput
		}
		return customfield.AssetType{}, err
	}
	s.observer.Record(ctx, ports.Event{
		Name:    ports.EventCustomAssetTypeRestored,
		Message: "custom asset type restored",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"principal_id":  input.Principal.ID.String(),
			"asset_type_id": restored.ID.String(),
			"scope":         restored.Scope.String(),
		},
	})
	return restored, nil
}

func (s Service) deleteCustomAssetType(ctx context.Context, input ArchiveCustomAssetTypeInput, scope customfield.Scope) error {
	assetTypeID, ok := customfield.NewAssetTypeID(input.CustomAssetTypeID.String())
	if !ok {
		return apperrors.ErrInvalidInput
	}
	current, found, err := s.customAssetTypes.CustomAssetTypeByID(ctx, input.TenantID, input.InventoryID, assetTypeID)
	if err != nil {
		return err
	}
	if !found || current.Scope != scope {
		return apperrors.ErrNotFound
	}
	if scope == customfield.ScopeInventory && current.InventoryID.String() != input.InventoryID.String() {
		return apperrors.ErrNotFound
	}
	hasReferences, err := s.customAssetTypes.CustomAssetTypeHasActiveReferences(ctx, input.TenantID, input.InventoryID, assetTypeID)
	if err != nil {
		return err
	}
	if hasReferences {
		return apperrors.ErrInvalidInput
	}
	auditRecord, err := s.newAuditRecord(appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionCustomAssetTypeDeleted,
		TargetType:  audit.TargetCustomAssetType,
		TargetID:    current.ID.String(),
		Metadata: map[string]string{
			"type_key":        current.Key.String(),
			"scope":           current.Scope.String(),
			"lifecycle_state": current.LifecycleState.String(),
		},
	})
	if err != nil {
		return err
	}
	if err := s.customAssetTypeUnitOfWork.DeleteCustomAssetType(ctx, input.TenantID, input.InventoryID, assetTypeID, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) || errors.Is(err, ports.ErrForbidden) {
			return apperrors.ErrInvalidInput
		}
		return err
	}
	s.observer.Record(ctx, ports.Event{
		Name:    ports.EventCustomAssetTypeDeleted,
		Message: "custom asset type deleted",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"principal_id":  input.Principal.ID.String(),
			"asset_type_id": current.ID.String(),
			"scope":         current.Scope.String(),
		},
	})
	return nil
}

func (s Service) ListTenantCustomAssetTypes(ctx context.Context, input ListCustomAssetTypesInput) (ListCustomAssetTypesResult, error) {
	if err := s.ensureTenantExists(ctx, input.TenantID); err != nil {
		return ListCustomAssetTypesResult{}, err
	}
	if err := s.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		s.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return ListCustomAssetTypesResult{}, err
	}

	limit := appsupport.PageLimit(s.defaultPageLimit, s.maxPageLimit, input.Limit)
	afterAssetTypeKey, err := decodeCustomAssetTypeCursor(input.TenantID, input.InventoryID, input.Cursor)
	if err != nil {
		return ListCustomAssetTypesResult{}, apperrors.ErrInvalidInput
	}
	items, err := s.customAssetTypes.ListTenantCustomAssetTypes(ctx, input.TenantID, ports.CustomAssetTypePageRequest{
		AfterAssetTypeKey: afterAssetTypeKey,
		Limit:             limit + 1,
	})
	if err != nil {
		return ListCustomAssetTypesResult{}, err
	}
	return s.customAssetTypeListResult(ctx, input, items, limit)
}

func (s Service) ListInventoryCustomAssetTypes(ctx context.Context, input ListCustomAssetTypesInput) (ListCustomAssetTypesResult, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return ListCustomAssetTypesResult{}, err
	}

	limit := appsupport.PageLimit(s.defaultPageLimit, s.maxPageLimit, input.Limit)
	afterAssetTypeKey, err := decodeCustomAssetTypeCursor(input.TenantID, input.InventoryID, input.Cursor)
	if err != nil {
		return ListCustomAssetTypesResult{}, apperrors.ErrInvalidInput
	}
	items, err := s.customAssetTypes.ListInventoryCustomAssetTypes(ctx, input.TenantID, input.InventoryID, ports.CustomAssetTypePageRequest{
		AfterAssetTypeKey: afterAssetTypeKey,
		Limit:             limit + 1,
	})
	if err != nil {
		return ListCustomAssetTypesResult{}, err
	}
	return s.customAssetTypeListResult(ctx, input, items, limit)
}

func (s Service) customAssetTypeListResult(ctx context.Context, input ListCustomAssetTypesInput, items []customfield.AssetType, limit int) (ListCustomAssetTypesResult, error) {
	hasMore := len(items) > limit
	var nextCursor *string
	if hasMore {
		items = items[:limit]
		nextCursor = encodeCustomAssetTypeCursor(input.TenantID, input.InventoryID, items[len(items)-1].CursorKey())
	}

	s.observer.Record(ctx, ports.Event{
		Name:    ports.EventCustomAssetTypesListed,
		Message: "custom asset types listed",
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
		Action:      audit.ActionCustomAssetTypeListed,
		TargetType:  audit.TargetTenant,
		TargetID:    input.TenantID.String(),
		Metadata: map[string]string{
			"limit": strconv.Itoa(limit),
		},
	}); err != nil {
		return ListCustomAssetTypesResult{}, err
	}

	return ListCustomAssetTypesResult{
		Items:      items,
		Limit:      limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func encodeCustomAssetTypeCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, key string) *string {
	return appsupport.EncodePageCursor("custom_asset_types", customFieldDefinitionCursorScope(tenantID, inventoryID), key)
}

func decodeCustomAssetTypeCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, cursor string) (string, error) {
	decoded, err := appsupport.DecodePageCursor("custom_asset_types", customFieldDefinitionCursorScope(tenantID, inventoryID), cursor)
	if err != nil {
		return "", err
	}
	if decoded == "" {
		return "", nil
	}
	return decoded, nil
}
