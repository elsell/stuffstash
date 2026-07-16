package app

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type AssetActivityView string

const (
	AssetActivityViewChanges AssetActivityView = "changes"
	AssetActivityViewAll     AssetActivityView = "all"
)

type ListAssetActivityInput struct {
	Principal   identity.Principal
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	AssetID     asset.ID
	View        AssetActivityView
	Limit       int
	Cursor      string
}

type ListAssetActivityResult struct {
	Items              []audit.AssetActivityEntry
	ResolvedPrincipals map[identity.PrincipalID]identity.User
	Limit              int
	NextCursor         *string
	HasMore            bool
}

func (a App) ListAssetActivity(ctx context.Context, input ListAssetActivityInput) (ListAssetActivityResult, error) {
	if err := a.ensureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return ListAssetActivityResult{}, err
	}
	if input.AssetID.String() == "" || a.assets == nil {
		return ListAssetActivityResult{}, ErrInvalidInput
	}
	canUndo := a.ensureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset) == nil
	if _, found, err := a.assets.AssetByID(ctx, input.TenantID, input.InventoryID, input.AssetID); err != nil {
		return ListAssetActivityResult{}, err
	} else if !found {
		return ListAssetActivityResult{}, ErrNotFound
	}
	view := input.View
	if view == "" {
		view = AssetActivityViewChanges
	}
	if view != AssetActivityViewChanges && view != AssetActivityViewAll {
		return ListAssetActivityResult{}, ErrInvalidInput
	}
	beforeOccurredAt, beforeRecordID, err := decodeAssetActivityCursor(input.TenantID, input.InventoryID, input.AssetID, view, input.Cursor)
	if err != nil {
		return ListAssetActivityResult{}, ErrInvalidInput
	}
	limit := pageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	request := ports.AssetAuditRecordListRequest{BeforeOccurredAt: beforeOccurredAt, BeforeRecordID: beforeRecordID, Limit: limit + 1}
	if view == AssetActivityViewChanges {
		request.Actions = audit.AssetActivityChangeActions()
	}
	records, err := a.audit.ListAssetAuditRecords(ctx, input.TenantID, input.InventoryID, input.AssetID.String(), request)
	if err != nil {
		return ListAssetActivityResult{}, err
	}
	hasMore := len(records) > limit
	var nextCursor *string
	if hasMore {
		records = records[:limit]
		nextCursor = encodeAssetActivityCursor(input.TenantID, input.InventoryID, input.AssetID, view, records[len(records)-1])
	}
	entries := make([]audit.AssetActivityEntry, 0, len(records))
	for _, record := range records {
		entries = append(entries, a.projectAssetActivityEntry(ctx, input, record, canUndo))
	}
	if err := a.saveReadAuditRecord(ctx, auditRecordInput{
		Principal: input.Principal, TenantID: input.TenantID, InventoryID: input.InventoryID, Source: audit.SourceAPI,
		Action: audit.ActionAuditRecordListed, TargetType: audit.TargetAuditRecord, TargetID: input.AssetID.String(),
		Metadata: map[string]string{"limit": strconv.Itoa(limit), "target_type": audit.TargetAsset.String(), "target_id": input.AssetID.String(), "view": string(view)},
	}); err != nil {
		return ListAssetActivityResult{}, err
	}
	a.observer.Record(ctx, ports.Event{Name: ports.EventAuditRecordsListed, Message: "asset activity listed", Fields: map[string]string{
		"tenant_id": input.TenantID.String(), "inventory_id": input.InventoryID.String(), "asset_id": input.AssetID.String(), "principal_id": input.Principal.ID.String(), "view": string(view), "limit": strconv.Itoa(limit),
	}})
	return ListAssetActivityResult{
		Items: entries, ResolvedPrincipals: a.resolveAuditPrincipals(ctx, ListAuditRecordsInput{Principal: input.Principal, TenantID: input.TenantID, InventoryID: input.InventoryID}, records),
		Limit: limit, NextCursor: nextCursor, HasMore: hasMore,
	}, nil
}

func (a App) projectAssetActivityEntry(ctx context.Context, input ListAssetActivityInput, record audit.Record, canUndo bool) audit.AssetActivityEntry {
	entry := audit.AssetActivityEntry{
		ID: record.ID, PrincipalID: record.PrincipalID, Action: record.Action, Category: record.Action.AssetActivityCategory(), Source: record.Source,
		OccurredAt: record.OccurredAt, RequestID: record.RequestID, Changes: projectAssetActivityChanges(record), TechnicalMetadata: map[string]string{},
	}
	operationID := strings.TrimSpace(record.Metadata["operation_id"])
	if canUndo && operationID != "" && a.undoables != nil {
		operation, found, err := a.undoables.UndoableOperationByID(ctx, input.TenantID, input.InventoryID, operationID)
		if err == nil && found && operation.TargetType == audit.TargetAsset && operation.TargetID == input.AssetID.String() && operation.OriginalAction == record.Action {
			entry.Undo = &audit.AssetActivityUndo{OperationID: operation.ID, Status: string(operation.Status)}
		}
	}
	return entry
}

func projectAssetActivityChanges(record audit.Record) []audit.AssetActivityChange {
	metadata := record.Metadata
	changes := make([]audit.AssetActivityChange, 0, 4)
	appendValues := func(field audit.AssetActivityField, previousKey, currentKey string) {
		previous, previousOK := metadata[previousKey]
		current, currentOK := metadata[currentKey]
		if previousOK || currentOK {
			changes = append(changes, audit.AssetActivityChange{Field: field, PreviousValue: previous, CurrentValue: current})
		}
	}
	appendValues(audit.AssetActivityFieldTitle, "previous_title", "updated_title")
	if metadata["description_changed"] == "true" {
		changes = append(changes, audit.AssetActivityChange{Field: audit.AssetActivityFieldDescription})
	}
	appendValues(audit.AssetActivityFieldTags, "previous_tag_count", "updated_tag_count")
	appendValues(audit.AssetActivityFieldParent, "previous_parent", "new_parent")
	appendValues(audit.AssetActivityFieldLifecycleState, "previous_lifecycle_state", "new_lifecycle_state")
	appendValues(audit.AssetActivityFieldCheckoutState, "previous_checkout_state", "new_checkout_state")
	return changes
}

type assetActivityCursorPayload struct {
	Version    int    `json:"v"`
	Collection string `json:"collection"`
	Scope      string `json:"scope"`
	View       string `json:"view"`
	LastID     string `json:"lastId"`
	OccurredAt string `json:"occurredAt"`
}

func encodeAssetActivityCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, view AssetActivityView, record audit.Record) *string {
	payload, err := json.Marshal(assetActivityCursorPayload{Version: paginationCursorVersion, Collection: "asset_activity", Scope: assetActivityCursorScope(tenantID, inventoryID, assetID), View: string(view), LastID: record.ID.String(), OccurredAt: record.OccurredAt.UTC().Format(time.RFC3339Nano)})
	if err != nil {
		return nil
	}
	cursor := base64.RawURLEncoding.EncodeToString(payload)
	return &cursor
}

func decodeAssetActivityCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, view AssetActivityView, cursor string) (time.Time, audit.ID, error) {
	if strings.TrimSpace(cursor) == "" {
		return time.Time{}, "", nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, "", err
	}
	var payload assetActivityCursorPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return time.Time{}, "", err
	}
	if payload.Version != paginationCursorVersion || payload.Collection != "asset_activity" || payload.Scope != assetActivityCursorScope(tenantID, inventoryID, assetID) || payload.View != string(view) || strings.TrimSpace(payload.LastID) == "" || strings.TrimSpace(payload.OccurredAt) == "" {
		return time.Time{}, "", ErrInvalidInput
	}
	occurredAt, err := time.Parse(time.RFC3339Nano, payload.OccurredAt)
	if err != nil {
		return time.Time{}, "", err
	}
	id, ok := audit.NewID(payload.LastID)
	if !ok {
		return time.Time{}, "", ErrInvalidInput
	}
	return occurredAt, id, nil
}

func assetActivityCursorScope(tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) string {
	return tenantID.String() + ":" + inventoryID.String() + ":" + assetID.String()
}
