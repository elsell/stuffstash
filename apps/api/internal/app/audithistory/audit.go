package audithistory

import (
	"context"
	"strconv"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type ListAuditRecordsInput struct {
	Principal   identity.Principal
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Limit       int
	Cursor      string
}

type ListAssetAuditHistoryInput struct {
	Principal   identity.Principal
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	AssetID     string
	Limit       int
}

type ListAuditRecordsResult struct {
	Items              []audit.Record
	ResolvedPrincipals map[identity.PrincipalID]identity.User
	Limit              int
	NextCursor         *string
	HasMore            bool
}

type ListAssetAuditHistoryResult struct {
	Items              []audit.Record
	ResolvedPrincipals map[identity.PrincipalID]identity.User
	Limit              int
	HasMore            bool
}

type auditRecordInput = appsupport.AuditRecordInput

func (a Service) saveReadAuditRecord(ctx context.Context, input auditRecordInput) error {
	return appsupport.SaveReadAuditRecord(ctx, a.deps.Audit, a.deps.IDs, a.deps.Clock, input)
}

func (a Service) ListTenantAuditRecords(ctx context.Context, input ListAuditRecordsInput) (ListAuditRecordsResult, error) {
	if err := a.deps.Access.EnsureTenantExists(ctx, input.TenantID); err != nil {
		return ListAuditRecordsResult{}, err
	}
	if err := a.deps.Authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return ListAuditRecordsResult{}, err
	}

	limit := appsupport.PageLimit(a.deps.DefaultPageLimit, a.deps.MaxPageLimit, input.Limit)
	afterOccurredAt, afterRecordID, err := decodeAuditRecordCursor(input.TenantID, input.InventoryID, input.Cursor)
	if err != nil {
		return ListAuditRecordsResult{}, apperrors.ErrInvalidInput
	}
	items, err := a.deps.Audit.ListTenantAuditRecords(ctx, input.TenantID, ports.AuditRecordPageRequest{
		AfterOccurredAt: afterOccurredAt,
		AfterRecordID:   afterRecordID,
		Limit:           limit + 1,
	})
	if err != nil {
		return ListAuditRecordsResult{}, err
	}
	return a.auditRecordListResult(ctx, input, items, limit)
}

func (a Service) ListInventoryAuditRecords(ctx context.Context, input ListAuditRecordsInput) (ListAuditRecordsResult, error) {
	if err := a.deps.Access.EnsureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return ListAuditRecordsResult{}, err
	}

	limit := appsupport.PageLimit(a.deps.DefaultPageLimit, a.deps.MaxPageLimit, input.Limit)
	afterOccurredAt, afterRecordID, err := decodeAuditRecordCursor(input.TenantID, input.InventoryID, input.Cursor)
	if err != nil {
		return ListAuditRecordsResult{}, apperrors.ErrInvalidInput
	}
	items, err := a.deps.Audit.ListInventoryAuditRecords(ctx, input.TenantID, input.InventoryID, ports.AuditRecordPageRequest{
		AfterOccurredAt: afterOccurredAt,
		AfterRecordID:   afterRecordID,
		Limit:           limit + 1,
	})
	if err != nil {
		return ListAuditRecordsResult{}, err
	}
	return a.auditRecordListResult(ctx, input, items, limit)
}

func (a Service) ListAssetAuditHistory(ctx context.Context, input ListAssetAuditHistoryInput) (ListAssetAuditHistoryResult, error) {
	if err := a.deps.Access.EnsureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return ListAssetAuditHistoryResult{}, err
	}
	if strings.TrimSpace(input.AssetID) == "" {
		return ListAssetAuditHistoryResult{}, apperrors.ErrInvalidInput
	}

	limit := appsupport.PageLimit(a.deps.DefaultPageLimit, a.deps.MaxPageLimit, input.Limit)
	items, err := a.deps.Audit.ListAssetAuditRecords(ctx, input.TenantID, input.InventoryID, input.AssetID, ports.AssetAuditRecordListRequest{
		Limit: limit + 1,
	})
	if err != nil {
		return ListAssetAuditHistoryResult{}, err
	}
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	a.deps.Observer.Record(ctx, ports.Event{
		Name:    ports.EventAuditRecordsListed,
		Message: "asset audit history listed",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"principal_id": input.Principal.ID.String(),
			"target_type":  audit.TargetAsset.String(),
			"limit":        strconv.Itoa(limit),
		},
	})
	if err := a.saveReadAuditRecord(ctx, auditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      audit.SourceAPI,
		Action:      audit.ActionAuditRecordListed,
		TargetType:  audit.TargetAuditRecord,
		TargetID:    input.AssetID,
		Metadata: map[string]string{
			"limit":       strconv.Itoa(limit),
			"target_type": audit.TargetAsset.String(),
			"target_id":   input.AssetID,
		},
	}); err != nil {
		return ListAssetAuditHistoryResult{}, err
	}
	return ListAssetAuditHistoryResult{
		Items:              items,
		ResolvedPrincipals: a.resolveAuditPrincipals(ctx, ListAuditRecordsInput{Principal: input.Principal, TenantID: input.TenantID, InventoryID: input.InventoryID}, items),
		Limit:              limit,
		HasMore:            hasMore,
	}, nil
}

func (a Service) auditRecordListResult(ctx context.Context, input ListAuditRecordsInput, items []audit.Record, limit int) (ListAuditRecordsResult, error) {
	hasMore := len(items) > limit
	var nextCursor *string
	if hasMore {
		items = items[:limit]
		nextCursor = encodeAuditRecordCursor(input.TenantID, input.InventoryID, items[len(items)-1])
	}

	resolvedPrincipals := a.resolveAuditPrincipals(ctx, input, items)
	a.deps.Observer.Record(ctx, ports.Event{
		Name:    ports.EventAuditRecordsListed,
		Message: "audit records listed",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"principal_id": input.Principal.ID.String(),
			"limit":        strconv.Itoa(limit),
		},
	})
	if err := a.saveReadAuditRecord(ctx, auditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Action:      audit.ActionAuditRecordListed,
		TargetType:  audit.TargetAuditRecord,
		TargetID:    auditRecordCursorScope(input.TenantID, input.InventoryID),
		Metadata: map[string]string{
			"limit": strconv.Itoa(limit),
		},
	}); err != nil {
		return ListAuditRecordsResult{}, err
	}

	return ListAuditRecordsResult{
		Items:              items,
		ResolvedPrincipals: resolvedPrincipals,
		Limit:              limit,
		NextCursor:         nextCursor,
		HasMore:            hasMore,
	}, nil
}
