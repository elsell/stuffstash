package audithistory

import (
	"context"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
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

func (a Service) ListAssetActivity(ctx context.Context, input ListAssetActivityInput) (ListAssetActivityResult, error) {
	if err := a.deps.Access.EnsureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return ListAssetActivityResult{}, err
	}
	if input.AssetID.String() == "" || a.deps.Assets == nil {
		return ListAssetActivityResult{}, apperrors.ErrInvalidInput
	}
	canUndo := a.deps.Access.EnsureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset) == nil
	if _, found, err := a.deps.Assets.AssetByID(ctx, input.TenantID, input.InventoryID, input.AssetID); err != nil {
		return ListAssetActivityResult{}, err
	} else if !found {
		return ListAssetActivityResult{}, apperrors.ErrNotFound
	}
	view := input.View
	if view == "" {
		view = AssetActivityViewChanges
	}
	if view != AssetActivityViewChanges && view != AssetActivityViewAll {
		return ListAssetActivityResult{}, apperrors.ErrInvalidInput
	}
	beforeOccurredAt, beforeRecordID, err := DecodeAssetActivityCursor(input.TenantID, input.InventoryID, input.AssetID, view, input.Cursor)
	if err != nil {
		return ListAssetActivityResult{}, apperrors.ErrInvalidInput
	}
	limit := appsupport.PageLimit(a.deps.DefaultPageLimit, a.deps.MaxPageLimit, input.Limit)
	request := ports.AssetAuditRecordListRequest{BeforeOccurredAt: beforeOccurredAt, BeforeRecordID: beforeRecordID, Limit: limit + 1}
	if view == AssetActivityViewChanges {
		request.Actions = audit.AssetActivityChangeActions()
	}
	records, err := a.deps.Audit.ListAssetAuditRecords(ctx, input.TenantID, input.InventoryID, input.AssetID.String(), request)
	if err != nil {
		return ListAssetActivityResult{}, err
	}
	hasMore := len(records) > limit
	var nextCursor *string
	if hasMore {
		records = records[:limit]
		nextCursor = EncodeAssetActivityCursor(input.TenantID, input.InventoryID, input.AssetID, view, records[len(records)-1])
	}
	entries := make([]audit.AssetActivityEntry, 0, len(records))
	for _, record := range records {
		entries = append(entries, a.ProjectAssetActivityEntry(ctx, input, record, canUndo))
	}
	if err := a.saveReadAuditRecord(ctx, auditRecordInput{
		Principal: input.Principal, TenantID: input.TenantID, InventoryID: input.InventoryID, Source: audit.SourceAPI,
		Action: audit.ActionAuditRecordListed, TargetType: audit.TargetAuditRecord, TargetID: input.AssetID.String(),
		Metadata: map[string]string{"limit": strconv.Itoa(limit), "target_type": audit.TargetAsset.String(), "target_id": input.AssetID.String(), "view": string(view)},
	}); err != nil {
		return ListAssetActivityResult{}, err
	}
	a.deps.Observer.Record(ctx, ports.Event{Name: ports.EventAuditRecordsListed, Message: "asset activity listed", Fields: map[string]string{
		"tenant_id": input.TenantID.String(), "inventory_id": input.InventoryID.String(), "asset_id": input.AssetID.String(), "principal_id": input.Principal.ID.String(), "view": string(view), "limit": strconv.Itoa(limit),
	}})
	return ListAssetActivityResult{
		Items: entries, ResolvedPrincipals: a.resolveAuditPrincipals(ctx, ListAuditRecordsInput{Principal: input.Principal, TenantID: input.TenantID, InventoryID: input.InventoryID}, records),
		Limit: limit, NextCursor: nextCursor, HasMore: hasMore,
	}, nil
}
