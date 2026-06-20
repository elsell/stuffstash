package app

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type CreateAssetInput struct {
	Principal         identity.Principal
	Source            audit.Source
	RequestID         string
	TenantID          tenant.ID
	InventoryID       inventory.InventoryID
	Kind              string
	Title             string
	Description       string
	ParentAssetID     string
	CustomAssetTypeID string
	CustomFields      map[string]any
}

type ListAssetsInput struct {
	Principal      identity.Principal
	Source         audit.Source
	RequestID      string
	TenantID       tenant.ID
	InventoryID    inventory.InventoryID
	Limit          int
	Cursor         string
	LifecycleState string
}

type GetAssetInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	AssetID     asset.ID
}

type AssetParentUpdate struct {
	Present bool
	Null    bool
	Value   string
}

type UpdateAssetInput struct {
	Principal     identity.Principal
	Source        audit.Source
	RequestID     string
	TenantID      tenant.ID
	InventoryID   inventory.InventoryID
	AssetID       asset.ID
	Title         *string
	Description   *string
	ParentAssetID AssetParentUpdate
	CustomFields  map[string]any
}

type UpdateAssetLifecycleInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	AssetID     asset.ID
}

type ListAssetsResult struct {
	Items      []asset.Asset
	Limit      int
	NextCursor *string
	HasMore    bool
}

func (a App) CreateAsset(ctx context.Context, input CreateAssetInput) (asset.Asset, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionCreateAsset); err != nil {
		return asset.Asset{}, err
	}

	kind, ok := asset.NewKind(input.Kind)
	if !ok {
		return asset.Asset{}, ErrInvalidInput
	}
	title, ok := asset.NewTitle(input.Title)
	if !ok {
		return asset.Asset{}, ErrInvalidInput
	}
	customAssetTypeID, err := a.validatedAssetCustomAssetTypeID(ctx, input.TenantID, input.InventoryID, input.CustomAssetTypeID)
	if err != nil {
		return asset.Asset{}, err
	}
	customFields, err := a.validatedCustomFields(ctx, input.TenantID, input.InventoryID, customAssetTypeID, input.CustomFields)
	if err != nil {
		return asset.Asset{}, err
	}

	id, ok := asset.NewID(a.ids.NewID())
	if !ok {
		return asset.Asset{}, ErrInvalidInput
	}

	parentAssetID := asset.ID("")
	if strings.TrimSpace(input.ParentAssetID) != "" {
		parsedParentID, ok := asset.NewID(input.ParentAssetID)
		if !ok {
			return asset.Asset{}, ErrInvalidInput
		}
		parent, found, err := a.assets.AssetByID(ctx, input.TenantID, input.InventoryID, parsedParentID)
		if err != nil {
			return asset.Asset{}, err
		}
		if !found {
			return asset.Asset{}, ErrNotFound
		}
		if !parent.Kind.CanContainChildren() || parent.LifecycleState != asset.LifecycleStateActive {
			return asset.Asset{}, ErrInvalidInput
		}
		parentAssetID = parsedParentID
	}

	item := asset.Asset{
		ID:                id,
		TenantID:          asset.TenantID(input.TenantID.String()),
		InventoryID:       asset.InventoryID(input.InventoryID.String()),
		ParentAssetID:     parentAssetID,
		CustomAssetTypeID: customAssetTypeID,
		Kind:              kind,
		Title:             title,
		Description:       asset.NewDescription(input.Description),
		CustomFields:      customFields,
		LifecycleState:    asset.LifecycleStateActive,
	}

	undoableOperation, err := a.newAssetUndoableOperation(input.Principal.ID, input.Source, input.TenantID, input.InventoryID, audit.ActionAssetCreated, nil, item)
	if err != nil {
		return asset.Asset{}, err
	}

	auditRecord, err := a.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionAssetCreated,
		TargetType:  audit.TargetAsset,
		TargetID:    item.ID.String(),
		Metadata: map[string]string{
			"asset_kind":   item.Kind.String(),
			"operation_id": undoableOperation.ID,
			"title":        item.Title.String(),
		},
	})
	if err != nil {
		return asset.Asset{}, err
	}
	if item.CustomAssetTypeID.String() != "" {
		auditRecord.Metadata["custom_asset_type_id"] = item.CustomAssetTypeID.String()
	}

	if a.assetUnitOfWork == nil {
		return asset.Asset{}, ErrInvalidInput
	}
	if err := a.assetUnitOfWork.CreateAsset(ctx, item, auditRecord, &undoableOperation); err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return asset.Asset{}, ErrInvalidInput
		}
		return asset.Asset{}, err
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAssetCreated,
		Message: "asset created",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"asset_id":     item.ID.String(),
			"asset_kind":   item.Kind.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})

	return item, nil
}

func (a App) UpdateAsset(ctx context.Context, input UpdateAssetInput) (asset.Asset, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return asset.Asset{}, err
	}
	if input.AssetID.String() == "" {
		return asset.Asset{}, ErrInvalidInput
	}

	current, found, err := a.assets.AssetByID(ctx, input.TenantID, input.InventoryID, input.AssetID)
	if err != nil {
		return asset.Asset{}, err
	}
	if !found {
		return asset.Asset{}, ErrNotFound
	}
	if current.LifecycleState != asset.LifecycleStateActive {
		return asset.Asset{}, ErrInvalidInput
	}
	updated := current
	parentChanged := false
	fieldsChanged := false

	if input.Title != nil {
		title, ok := asset.NewTitle(*input.Title)
		if !ok {
			return asset.Asset{}, ErrInvalidInput
		}
		updated.Title = title
		if updated.Title != current.Title {
			fieldsChanged = true
		}
	}
	if input.Description != nil {
		updated.Description = asset.NewDescription(*input.Description)
		if updated.Description != current.Description {
			fieldsChanged = true
		}
	}
	if input.CustomFields != nil {
		customFields, err := a.validatedCustomFields(ctx, input.TenantID, input.InventoryID, current.CustomAssetTypeID, input.CustomFields)
		if err != nil {
			return asset.Asset{}, err
		}
		updated.CustomFields = customFields
		fieldsChanged = true
	}
	if input.ParentAssetID.Present {
		parentAssetID := asset.ID("")
		if !input.ParentAssetID.Null {
			parentAsset, err := a.validatedParentAssetID(ctx, input.TenantID, input.InventoryID, input.AssetID, input.ParentAssetID.Value)
			if err != nil {
				return asset.Asset{}, err
			}
			parentAssetID = parentAsset
		}
		updated.ParentAssetID = parentAssetID
		if updated.ParentAssetID != current.ParentAssetID {
			parentChanged = true
		}
	}

	var undoableOperation *ports.UndoableOperation
	if fieldsChanged || parentChanged {
		operationAction := audit.ActionAssetMoved
		if fieldsChanged {
			operationAction = audit.ActionAssetUpdated
		}
		operation, err := a.newAssetUndoableOperation(input.Principal.ID, input.Source, input.TenantID, input.InventoryID, operationAction, &current, updated)
		if err != nil {
			return asset.Asset{}, err
		}
		undoableOperation = &operation
	}

	auditRecords := []audit.Record{}
	if fieldsChanged {
		metadata := map[string]string{
			"asset_kind": updated.Kind.String(),
		}
		if undoableOperation != nil {
			metadata["operation_id"] = undoableOperation.ID
		}
		auditRecord, err := a.newAuditRecord(auditRecordInput{
			PrincipalID: input.Principal.ID,
			TenantID:    input.TenantID,
			InventoryID: input.InventoryID,
			Source:      input.Source,
			RequestID:   input.RequestID,
			Action:      audit.ActionAssetUpdated,
			TargetType:  audit.TargetAsset,
			TargetID:    updated.ID.String(),
			Metadata:    metadata,
		})
		if err != nil {
			return asset.Asset{}, err
		}
		auditRecords = append(auditRecords, auditRecord)
	}
	if parentChanged {
		metadata := map[string]string{
			"asset_kind":      updated.Kind.String(),
			"previous_parent": current.ParentAssetID.String(),
			"new_parent":      updated.ParentAssetID.String(),
		}
		if undoableOperation != nil {
			metadata["operation_id"] = undoableOperation.ID
		}
		auditRecord, err := a.newAuditRecord(auditRecordInput{
			PrincipalID: input.Principal.ID,
			TenantID:    input.TenantID,
			InventoryID: input.InventoryID,
			Source:      input.Source,
			RequestID:   input.RequestID,
			Action:      audit.ActionAssetMoved,
			TargetType:  audit.TargetAsset,
			TargetID:    updated.ID.String(),
			Metadata:    metadata,
		})
		if err != nil {
			return asset.Asset{}, err
		}
		auditRecords = append(auditRecords, auditRecord)
	}

	if a.assetUnitOfWork == nil {
		return asset.Asset{}, ErrInvalidInput
	}
	if err := a.assetUnitOfWork.UpdateAsset(ctx, updated, auditRecords, undoableOperation); err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return asset.Asset{}, ErrInvalidInput
		}
		return asset.Asset{}, err
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAssetUpdated,
		Message: "asset updated",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"asset_id":     updated.ID.String(),
			"asset_kind":   updated.Kind.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})

	return updated, nil
}

func (a App) ArchiveAsset(ctx context.Context, input UpdateAssetLifecycleInput) (asset.Asset, error) {
	return a.updateAssetLifecycle(ctx, input, asset.LifecycleStateActive, asset.LifecycleStateArchived, audit.ActionAssetArchived, ports.EventAssetArchived, "asset archived")
}

func (a App) RestoreAsset(ctx context.Context, input UpdateAssetLifecycleInput) (asset.Asset, error) {
	return a.updateAssetLifecycle(ctx, input, asset.LifecycleStateArchived, asset.LifecycleStateActive, audit.ActionAssetRestored, ports.EventAssetRestored, "asset restored")
}

func (a App) GetAsset(ctx context.Context, input GetAssetInput) (asset.Asset, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return asset.Asset{}, err
	}
	if input.AssetID.String() == "" {
		return asset.Asset{}, ErrInvalidInput
	}
	item, found, err := a.assets.AssetByID(ctx, input.TenantID, input.InventoryID, input.AssetID)
	if err != nil {
		return asset.Asset{}, err
	}
	if !found {
		return asset.Asset{}, ErrNotFound
	}
	if err := a.saveReadAuditRecord(ctx, auditRecordInput{
		PrincipalID: input.Principal.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionAssetViewed,
		TargetType:  audit.TargetAsset,
		TargetID:    item.ID.String(),
		Metadata: map[string]string{
			"asset_kind":      item.Kind.String(),
			"lifecycle_state": item.LifecycleState.String(),
		},
	}); err != nil {
		return asset.Asset{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAssetViewed,
		Message: "asset viewed",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"asset_id":     item.ID.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})
	return item, nil
}

func (a App) DeleteAsset(ctx context.Context, input UpdateAssetLifecycleInput) error {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return err
	}
	if input.AssetID.String() == "" {
		return ErrInvalidInput
	}
	item, found, err := a.assets.AssetByID(ctx, input.TenantID, input.InventoryID, input.AssetID)
	if err != nil {
		return err
	}
	if !found {
		return ErrNotFound
	}
	auditRecord, err := a.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionAssetDeleted,
		TargetType:  audit.TargetAsset,
		TargetID:    item.ID.String(),
		Metadata: map[string]string{
			"asset_kind":      item.Kind.String(),
			"lifecycle_state": item.LifecycleState.String(),
		},
	})
	if err != nil {
		return err
	}
	if a.assetUnitOfWork == nil {
		return ErrInvalidInput
	}
	if err := a.assetUnitOfWork.DeleteAsset(ctx, input.TenantID, input.InventoryID, input.AssetID, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return ErrInvalidInput
		}
		return err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAssetDeleted,
		Message: "asset deleted",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"asset_id":     input.AssetID.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})
	return nil
}

func (a App) updateAssetLifecycle(ctx context.Context, input UpdateAssetLifecycleInput, from asset.LifecycleState, to asset.LifecycleState, action audit.Action, eventName ports.EventName, eventMessage string) (asset.Asset, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return asset.Asset{}, err
	}
	if input.AssetID.String() == "" {
		return asset.Asset{}, ErrInvalidInput
	}

	current, found, err := a.assets.AssetByID(ctx, input.TenantID, input.InventoryID, input.AssetID)
	if err != nil {
		return asset.Asset{}, err
	}
	if !found {
		return asset.Asset{}, ErrNotFound
	}
	if current.LifecycleState != from {
		return asset.Asset{}, ErrInvalidInput
	}
	if to == asset.LifecycleStateArchived {
		hasActiveChildren, err := a.assets.AssetHasActiveChildren(ctx, input.TenantID, input.InventoryID, input.AssetID)
		if err != nil {
			return asset.Asset{}, err
		}
		if hasActiveChildren {
			return asset.Asset{}, ErrInvalidInput
		}
	}
	if to == asset.LifecycleStateActive && current.ParentAssetID.String() != "" {
		parent, found, err := a.assets.AssetByID(ctx, input.TenantID, input.InventoryID, current.ParentAssetID)
		if err != nil {
			return asset.Asset{}, err
		}
		if !found {
			return asset.Asset{}, ErrInvalidInput
		}
		if parent.LifecycleState != asset.LifecycleStateActive {
			return asset.Asset{}, ErrInvalidInput
		}
	}

	updated := current
	updated.LifecycleState = to
	undoableOperation, err := a.newAssetUndoableOperation(input.Principal.ID, input.Source, input.TenantID, input.InventoryID, action, &current, updated)
	if err != nil {
		return asset.Asset{}, err
	}
	auditRecord, err := a.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      action,
		TargetType:  audit.TargetAsset,
		TargetID:    updated.ID.String(),
		Metadata: map[string]string{
			"asset_kind":      updated.Kind.String(),
			"operation_id":    undoableOperation.ID,
			"previous_state":  current.LifecycleState.String(),
			"lifecycle_state": updated.LifecycleState.String(),
		},
	})
	if err != nil {
		return asset.Asset{}, err
	}

	if a.assetUnitOfWork == nil {
		return asset.Asset{}, ErrInvalidInput
	}
	if err := a.assetUnitOfWork.UpdateAssetLifecycle(ctx, updated, auditRecord, &undoableOperation); err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return asset.Asset{}, ErrInvalidInput
		}
		return asset.Asset{}, err
	}

	a.observer.Record(ctx, ports.Event{
		Name:    eventName,
		Message: eventMessage,
		Fields: map[string]string{
			"tenant_id":       input.TenantID.String(),
			"inventory_id":    input.InventoryID.String(),
			"asset_id":        updated.ID.String(),
			"asset_kind":      updated.Kind.String(),
			"principal_id":    input.Principal.ID.String(),
			"lifecycle_state": updated.LifecycleState.String(),
			"previous_state":  current.LifecycleState.String(),
		},
	})

	return updated, nil
}

func (a App) newAssetUndoableOperation(principalID identity.PrincipalID, source audit.Source, tenantID tenant.ID, inventoryID inventory.InventoryID, originalAction audit.Action, before *asset.Asset, after asset.Asset) (ports.UndoableOperation, error) {
	if a.undoables == nil {
		return ports.UndoableOperation{}, ErrInvalidInput
	}
	id := a.ids.NewID()
	if strings.TrimSpace(id) == "" {
		return ports.UndoableOperation{}, ErrInvalidInput
	}
	var beforeCopy *asset.Asset
	if before != nil {
		copied := *before
		beforeCopy = &copied
	}
	return ports.UndoableOperation{
		ID:             id,
		TenantID:       tenantID,
		InventoryID:    inventoryID,
		PrincipalID:    principalID,
		Source:         source,
		TargetType:     audit.TargetAsset,
		TargetID:       after.ID.String(),
		OriginalAction: originalAction,
		Status:         ports.UndoableOperationAvailable,
		CreatedAt:      time.Now().UTC(),
		BeforeAsset:    beforeCopy,
		AfterAsset:     after,
	}, nil
}

func (a App) validatedAssetCustomAssetTypeID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, rawCustomAssetTypeID string) (asset.CustomAssetTypeID, error) {
	if strings.TrimSpace(rawCustomAssetTypeID) == "" {
		return "", nil
	}
	customAssetTypeID, ok := customfield.NewAssetTypeID(rawCustomAssetTypeID)
	if !ok {
		return "", ErrInvalidInput
	}
	if a.customAssetTypes == nil {
		return "", ErrInvalidInput
	}
	types, err := a.customAssetTypes.CustomAssetTypesByID(ctx, tenantID, inventoryID, []customfield.AssetTypeID{customAssetTypeID})
	if err != nil {
		return "", err
	}
	if len(types) != 1 {
		return "", ErrNotFound
	}
	return asset.CustomAssetTypeID(customAssetTypeID.String()), nil
}

func (a App) validatedCustomFields(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, customAssetTypeID asset.CustomAssetTypeID, values map[string]any) (asset.CustomFields, error) {
	customFields, ok := asset.NewCustomFields(normalizeCustomFieldValues(values))
	if !ok {
		return asset.CustomFields{}, ErrInvalidInput
	}
	if customFields.IsEmpty() {
		return customFields, nil
	}
	if a.customFields == nil {
		return asset.CustomFields{}, ErrInvalidInput
	}
	definitions, err := a.customFields.ListEffectiveCustomFieldDefinitions(ctx, tenantID, inventoryID)
	if err != nil {
		return asset.CustomFields{}, err
	}
	if !customfield.DefinitionSet(definitions).ValidateValuesForAssetType(customFields.Values(), customfield.AssetTypeID(customAssetTypeID.String())) {
		return asset.CustomFields{}, ErrInvalidInput
	}
	return customFields, nil
}

func (a App) validatedParentAssetID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, rawParentAssetID string) (asset.ID, error) {
	if strings.TrimSpace(rawParentAssetID) == "" {
		return asset.ID(""), ErrInvalidInput
	}
	parentAssetID, ok := asset.NewID(rawParentAssetID)
	if !ok || parentAssetID == assetID {
		return asset.ID(""), ErrInvalidInput
	}
	parent, found, err := a.assets.AssetByID(ctx, tenantID, inventoryID, parentAssetID)
	if err != nil {
		return asset.ID(""), err
	}
	if !found {
		return asset.ID(""), ErrNotFound
	}
	if !parent.Kind.CanContainChildren() || parent.LifecycleState != asset.LifecycleStateActive {
		return asset.ID(""), ErrInvalidInput
	}
	for current := parent; current.ParentAssetID.String() != ""; {
		if current.ParentAssetID == assetID {
			return asset.ID(""), ErrInvalidInput
		}
		next, found, err := a.assets.AssetByID(ctx, tenantID, inventoryID, current.ParentAssetID)
		if err != nil {
			return asset.ID(""), err
		}
		if !found {
			return asset.ID(""), ErrInvalidInput
		}
		current = next
	}
	return parentAssetID, nil
}

func normalizeCustomFieldValues(values map[string]any) map[string]any {
	normalized := map[string]any{}
	for key, value := range values {
		normalized[key] = customfield.NormalizeJSONNumber(value)
	}
	return normalized
}

func (a App) ListAssets(ctx context.Context, input ListAssetsInput) (ListAssetsResult, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return ListAssetsResult{}, err
	}

	limit := pageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	lifecycleFilter, err := assetLifecycleFilter(input.LifecycleState)
	if err != nil {
		return ListAssetsResult{}, ErrInvalidInput
	}
	afterAssetID, err := decodeAssetCursor(input.TenantID, input.InventoryID, lifecycleFilter, input.Cursor)
	if err != nil {
		return ListAssetsResult{}, ErrInvalidInput
	}

	items, err := a.assets.ListAssetsByInventory(ctx, input.TenantID, input.InventoryID, ports.AssetListPageRequest{
		AfterAssetID:    afterAssetID,
		Limit:           limit + 1,
		LifecycleFilter: lifecycleFilter,
	})
	if err != nil {
		return ListAssetsResult{}, err
	}

	hasMore := len(items) > limit
	var nextCursor *string
	if hasMore {
		items = items[:limit]
		nextCursor = encodeAssetCursor(input.TenantID, input.InventoryID, lifecycleFilter, items[len(items)-1].ID)
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAssetsListed,
		Message: "assets listed",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"principal_id": input.Principal.ID.String(),
			"limit":        strings.TrimSpace(strconv.Itoa(limit)),
			"lifecycle":    string(lifecycleFilter),
		},
	})
	if err := a.saveReadAuditRecord(ctx, auditRecordInput{
		PrincipalID: input.Principal.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionAssetListed,
		TargetType:  audit.TargetInventory,
		TargetID:    input.InventoryID.String(),
		Metadata: map[string]string{
			"limit":     strconv.Itoa(limit),
			"lifecycle": string(lifecycleFilter),
		},
	}); err != nil {
		return ListAssetsResult{}, err
	}

	return ListAssetsResult{
		Items:      items,
		Limit:      limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func assetLifecycleFilter(value string) (ports.AssetLifecycleFilter, error) {
	switch strings.TrimSpace(value) {
	case "":
		return ports.AssetLifecycleFilterActive, nil
	case string(ports.AssetLifecycleFilterActive):
		return ports.AssetLifecycleFilterActive, nil
	case string(ports.AssetLifecycleFilterArchived):
		return ports.AssetLifecycleFilterArchived, nil
	case string(ports.AssetLifecycleFilterAll):
		return ports.AssetLifecycleFilterAll, nil
	default:
		return "", ErrInvalidInput
	}
}

func encodeAssetCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, lifecycleFilter ports.AssetLifecycleFilter, id asset.ID) *string {
	return encodePageCursor("assets", tenantID.String()+":"+inventoryID.String()+":"+string(lifecycleFilter), id.String())
}

func decodeAssetCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, lifecycleFilter ports.AssetLifecycleFilter, cursor string) (asset.ID, error) {
	decoded, err := decodePageCursor("assets", tenantID.String()+":"+inventoryID.String()+":"+string(lifecycleFilter), cursor)
	if err != nil {
		return asset.ID(""), err
	}
	if decoded == "" {
		return asset.ID(""), nil
	}
	id, ok := asset.NewID(decoded)
	if !ok {
		return asset.ID(""), ErrInvalidInput
	}
	return id, nil
}

func (a App) ensureActiveInventoryAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID, permission ports.InventoryPermission) error {
	item, err := a.ensureInventoryAccessItem(ctx, principal, tenantID, inventoryID, permission)
	if err != nil {
		return err
	}
	if !item.IsActive() {
		return ErrNotFound
	}
	return nil
}

func (a App) ensureInventoryAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID, permission ports.InventoryPermission) error {
	_, err := a.ensureInventoryAccessItem(ctx, principal, tenantID, inventoryID, permission)
	return err
}

func (a App) ensureInventoryAccessItem(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID, permission ports.InventoryPermission) (inventory.Inventory, error) {
	exists, err := a.tenants.TenantExists(ctx, tenantID)
	if err != nil {
		return inventory.Inventory{}, err
	}
	if !exists {
		return inventory.Inventory{}, ErrNotFound
	}

	item, found, err := a.inventories.InventoryByID(ctx, tenantID, inventoryID)
	if err != nil {
		return inventory.Inventory{}, err
	}
	if !found {
		return inventory.Inventory{}, ErrNotFound
	}

	if err := a.authorizer.CheckInventory(ctx, principal, permission, inventoryID); err != nil {
		a.recordAuthorizationDenied(ctx, principal, tenantID)
		return inventory.Inventory{}, err
	}
	return item, nil
}
