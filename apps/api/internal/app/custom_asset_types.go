package app

import (
	"context"
	"errors"
	"strconv"

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
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Limit       int
	Cursor      string
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

func (a App) CreateTenantCustomAssetType(ctx context.Context, input CreateCustomAssetTypeInput) (customfield.AssetType, error) {
	if err := a.ensureTenantExists(ctx, input.TenantID); err != nil {
		return customfield.AssetType{}, err
	}
	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return customfield.AssetType{}, err
	}
	return a.createCustomAssetType(ctx, input, customfield.ScopeTenant)
}

func (a App) CreateInventoryCustomAssetType(ctx context.Context, input CreateCustomAssetTypeInput) (customfield.AssetType, error) {
	if err := a.ensureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return customfield.AssetType{}, err
	}
	return a.createCustomAssetType(ctx, input, customfield.ScopeInventory)
}

func (a App) UpdateTenantCustomAssetType(ctx context.Context, input UpdateCustomAssetTypeInput) (customfield.AssetType, error) {
	if err := a.ensureTenantExists(ctx, input.TenantID); err != nil {
		return customfield.AssetType{}, err
	}
	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return customfield.AssetType{}, err
	}
	return a.updateCustomAssetType(ctx, input, customfield.ScopeTenant)
}

func (a App) UpdateInventoryCustomAssetType(ctx context.Context, input UpdateCustomAssetTypeInput) (customfield.AssetType, error) {
	if err := a.ensureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return customfield.AssetType{}, err
	}
	return a.updateCustomAssetType(ctx, input, customfield.ScopeInventory)
}

func (a App) ArchiveTenantCustomAssetType(ctx context.Context, input ArchiveCustomAssetTypeInput) (customfield.AssetType, error) {
	if err := a.ensureTenantExists(ctx, input.TenantID); err != nil {
		return customfield.AssetType{}, err
	}
	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return customfield.AssetType{}, err
	}
	return a.archiveCustomAssetType(ctx, input, customfield.ScopeTenant)
}

func (a App) ArchiveInventoryCustomAssetType(ctx context.Context, input ArchiveCustomAssetTypeInput) (customfield.AssetType, error) {
	if err := a.ensureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return customfield.AssetType{}, err
	}
	return a.archiveCustomAssetType(ctx, input, customfield.ScopeInventory)
}

func (a App) createCustomAssetType(ctx context.Context, input CreateCustomAssetTypeInput, scope customfield.Scope) (customfield.AssetType, error) {
	id, ok := customfield.NewAssetTypeID(a.ids.NewID())
	if !ok {
		return customfield.AssetType{}, ErrInvalidInput
	}
	key, ok := customfield.NewKey(input.Key)
	if !ok {
		return customfield.AssetType{}, ErrInvalidInput
	}
	displayName, ok := customfield.NewDisplayName(input.DisplayName)
	if !ok {
		return customfield.AssetType{}, ErrInvalidInput
	}
	description, ok := customfield.NewDescription(input.Description)
	if !ok {
		return customfield.AssetType{}, ErrInvalidInput
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
		return customfield.AssetType{}, ErrInvalidInput
	}

	auditRecord, err := a.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
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

	if err := a.customAssetTypes.SaveCustomAssetType(ctx, assetType, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return customfield.AssetType{}, ErrInvalidInput
		}
		return customfield.AssetType{}, err
	}

	a.observer.Record(ctx, ports.Event{
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

func (a App) updateCustomAssetType(ctx context.Context, input UpdateCustomAssetTypeInput, scope customfield.Scope) (customfield.AssetType, error) {
	assetTypeID, ok := customfield.NewAssetTypeID(input.CustomAssetTypeID.String())
	if !ok || (input.DisplayName == nil && input.Description == nil) {
		return customfield.AssetType{}, ErrInvalidInput
	}
	current, found, err := a.customAssetTypes.CustomAssetTypeByID(ctx, input.TenantID, input.InventoryID, assetTypeID)
	if err != nil {
		return customfield.AssetType{}, err
	}
	if !found {
		return customfield.AssetType{}, ErrNotFound
	}
	if current.Scope != scope {
		return customfield.AssetType{}, ErrNotFound
	}
	if scope == customfield.ScopeInventory && current.InventoryID.String() != input.InventoryID.String() {
		return customfield.AssetType{}, ErrNotFound
	}
	if !current.IsActive() {
		return customfield.AssetType{}, ErrInvalidInput
	}

	updated := current
	changedFields := map[string]string{}
	if input.DisplayName != nil {
		displayName, ok := customfield.NewDisplayName(*input.DisplayName)
		if !ok {
			return customfield.AssetType{}, ErrInvalidInput
		}
		if displayName != current.DisplayName {
			updated.DisplayName = displayName
			changedFields["display_name"] = "true"
		}
	}
	if input.Description != nil {
		description, ok := customfield.NewDescription(*input.Description)
		if !ok {
			return customfield.AssetType{}, ErrInvalidInput
		}
		if description != current.Description {
			updated.Description = description
			changedFields["description"] = "true"
		}
	}
	if len(changedFields) == 0 {
		return current, nil
	}

	auditRecord, err := a.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
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

	if err := a.customAssetTypes.UpdateCustomAssetType(ctx, updated, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) || errors.Is(err, ports.ErrForbidden) {
			return customfield.AssetType{}, ErrInvalidInput
		}
		return customfield.AssetType{}, err
	}

	a.observer.Record(ctx, ports.Event{
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

func (a App) archiveCustomAssetType(ctx context.Context, input ArchiveCustomAssetTypeInput, scope customfield.Scope) (customfield.AssetType, error) {
	assetTypeID, ok := customfield.NewAssetTypeID(input.CustomAssetTypeID.String())
	if !ok {
		return customfield.AssetType{}, ErrInvalidInput
	}
	current, found, err := a.customAssetTypes.CustomAssetTypeByID(ctx, input.TenantID, input.InventoryID, assetTypeID)
	if err != nil {
		return customfield.AssetType{}, err
	}
	if !found {
		return customfield.AssetType{}, ErrNotFound
	}
	if current.Scope != scope {
		return customfield.AssetType{}, ErrNotFound
	}
	if scope == customfield.ScopeInventory && current.InventoryID.String() != input.InventoryID.String() {
		return customfield.AssetType{}, ErrNotFound
	}
	archived, ok := current.Archive()
	if !ok {
		return customfield.AssetType{}, ErrInvalidInput
	}

	auditRecord, err := a.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
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

	if err := a.customAssetTypes.ArchiveCustomAssetType(ctx, archived, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) || errors.Is(err, ports.ErrForbidden) {
			return customfield.AssetType{}, ErrInvalidInput
		}
		return customfield.AssetType{}, err
	}

	a.observer.Record(ctx, ports.Event{
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

func (a App) ListTenantCustomAssetTypes(ctx context.Context, input ListCustomAssetTypesInput) (ListCustomAssetTypesResult, error) {
	if err := a.ensureTenantExists(ctx, input.TenantID); err != nil {
		return ListCustomAssetTypesResult{}, err
	}
	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return ListCustomAssetTypesResult{}, err
	}

	limit := pageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	afterAssetTypeKey, err := decodeCustomAssetTypeCursor(input.TenantID, input.InventoryID, input.Cursor)
	if err != nil {
		return ListCustomAssetTypesResult{}, ErrInvalidInput
	}
	items, err := a.customAssetTypes.ListTenantCustomAssetTypes(ctx, input.TenantID, ports.CustomAssetTypePageRequest{
		AfterAssetTypeKey: afterAssetTypeKey,
		Limit:             limit + 1,
	})
	if err != nil {
		return ListCustomAssetTypesResult{}, err
	}
	return a.customAssetTypeListResult(ctx, input, items, limit), nil
}

func (a App) ListInventoryCustomAssetTypes(ctx context.Context, input ListCustomAssetTypesInput) (ListCustomAssetTypesResult, error) {
	if err := a.ensureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return ListCustomAssetTypesResult{}, err
	}

	limit := pageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	afterAssetTypeKey, err := decodeCustomAssetTypeCursor(input.TenantID, input.InventoryID, input.Cursor)
	if err != nil {
		return ListCustomAssetTypesResult{}, ErrInvalidInput
	}
	items, err := a.customAssetTypes.ListInventoryCustomAssetTypes(ctx, input.TenantID, input.InventoryID, ports.CustomAssetTypePageRequest{
		AfterAssetTypeKey: afterAssetTypeKey,
		Limit:             limit + 1,
	})
	if err != nil {
		return ListCustomAssetTypesResult{}, err
	}
	return a.customAssetTypeListResult(ctx, input, items, limit), nil
}

func (a App) customAssetTypeListResult(ctx context.Context, input ListCustomAssetTypesInput, items []customfield.AssetType, limit int) ListCustomAssetTypesResult {
	hasMore := len(items) > limit
	var nextCursor *string
	if hasMore {
		items = items[:limit]
		nextCursor = encodeCustomAssetTypeCursor(input.TenantID, input.InventoryID, items[len(items)-1].CursorKey())
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventCustomAssetTypesListed,
		Message: "custom asset types listed",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"principal_id": input.Principal.ID.String(),
			"limit":        strconv.Itoa(limit),
		},
	})

	return ListCustomAssetTypesResult{
		Items:      items,
		Limit:      limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}
}

func encodeCustomAssetTypeCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, key string) *string {
	return encodePageCursor("custom_asset_types", customFieldDefinitionCursorScope(tenantID, inventoryID), key)
}

func decodeCustomAssetTypeCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, cursor string) (string, error) {
	decoded, err := decodePageCursor("custom_asset_types", customFieldDefinitionCursorScope(tenantID, inventoryID), cursor)
	if err != nil {
		return "", err
	}
	if decoded == "" {
		return "", nil
	}
	return decoded, nil
}
