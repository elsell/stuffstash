package app

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"time"

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
	Items      []audit.Record
	Limit      int
	NextCursor *string
	HasMore    bool
}

type ListAssetAuditHistoryResult struct {
	Items []audit.Record
	Limit int
}

type auditRecordInput = appsupport.AuditRecordInput

func (a App) newAuditRecord(input auditRecordInput) (audit.Record, error) {
	return appsupport.NewAuditRecord(a.ids, a.clock, input)
}

func (a App) saveReadAuditRecord(ctx context.Context, input auditRecordInput) error {
	return appsupport.SaveReadAuditRecord(ctx, a.audit, a.ids, a.clock, input)
}

func (a App) ListTenantAuditRecords(ctx context.Context, input ListAuditRecordsInput) (ListAuditRecordsResult, error) {
	if err := a.ensureTenantExists(ctx, input.TenantID); err != nil {
		return ListAuditRecordsResult{}, err
	}
	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return ListAuditRecordsResult{}, err
	}

	limit := pageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	afterOccurredAt, afterRecordID, err := decodeAuditRecordCursor(input.TenantID, input.InventoryID, input.Cursor)
	if err != nil {
		return ListAuditRecordsResult{}, ErrInvalidInput
	}
	items, err := a.audit.ListTenantAuditRecords(ctx, input.TenantID, ports.AuditRecordPageRequest{
		AfterOccurredAt: afterOccurredAt,
		AfterRecordID:   afterRecordID,
		Limit:           limit + 1,
	})
	if err != nil {
		return ListAuditRecordsResult{}, err
	}
	return a.auditRecordListResult(ctx, input, items, limit)
}

func (a App) ListInventoryAuditRecords(ctx context.Context, input ListAuditRecordsInput) (ListAuditRecordsResult, error) {
	if err := a.ensureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return ListAuditRecordsResult{}, err
	}

	limit := pageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	afterOccurredAt, afterRecordID, err := decodeAuditRecordCursor(input.TenantID, input.InventoryID, input.Cursor)
	if err != nil {
		return ListAuditRecordsResult{}, ErrInvalidInput
	}
	items, err := a.audit.ListInventoryAuditRecords(ctx, input.TenantID, input.InventoryID, ports.AuditRecordPageRequest{
		AfterOccurredAt: afterOccurredAt,
		AfterRecordID:   afterRecordID,
		Limit:           limit + 1,
	})
	if err != nil {
		return ListAuditRecordsResult{}, err
	}
	return a.auditRecordListResult(ctx, input, items, limit)
}

func (a App) ListAssetAuditHistory(ctx context.Context, input ListAssetAuditHistoryInput) (ListAssetAuditHistoryResult, error) {
	if err := a.ensureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return ListAssetAuditHistoryResult{}, err
	}
	if strings.TrimSpace(input.AssetID) == "" {
		return ListAssetAuditHistoryResult{}, ErrInvalidInput
	}

	limit := pageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	items, err := a.audit.ListAssetAuditRecords(ctx, input.TenantID, input.InventoryID, input.AssetID, ports.AssetAuditRecordListRequest{
		Limit: limit,
	})
	if err != nil {
		return ListAssetAuditHistoryResult{}, err
	}
	a.observer.Record(ctx, ports.Event{
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
		PrincipalID: input.Principal.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      audit.SourceAPI,
		Action:      audit.ActionAuditRecordListed,
		TargetType:  audit.TargetAuditRecord,
		TargetID:    audit.TargetAsset.String(),
		Metadata: map[string]string{
			"limit":       strconv.Itoa(limit),
			"target_type": audit.TargetAsset.String(),
		},
	}); err != nil {
		return ListAssetAuditHistoryResult{}, err
	}
	return ListAssetAuditHistoryResult{Items: items, Limit: limit}, nil
}

func (a App) auditRecordListResult(ctx context.Context, input ListAuditRecordsInput, items []audit.Record, limit int) (ListAuditRecordsResult, error) {
	hasMore := len(items) > limit
	var nextCursor *string
	if hasMore {
		items = items[:limit]
		nextCursor = encodeAuditRecordCursor(input.TenantID, input.InventoryID, items[len(items)-1])
	}

	a.observer.Record(ctx, ports.Event{
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
		PrincipalID: input.Principal.ID,
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
		Items:      items,
		Limit:      limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

type auditRecordCursorPayload struct {
	Version    int    `json:"v"`
	Collection string `json:"collection"`
	Scope      string `json:"scope"`
	LastID     string `json:"lastId"`
	OccurredAt string `json:"occurredAt"`
}

func encodeAuditRecordCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, record audit.Record) *string {
	payload, err := json.Marshal(auditRecordCursorPayload{
		Version:    paginationCursorVersion,
		Collection: "audit_records",
		Scope:      auditRecordCursorScope(tenantID, inventoryID),
		LastID:     record.ID.String(),
		OccurredAt: record.OccurredAt.UTC().Format(time.RFC3339Nano),
	})
	if err != nil {
		return nil
	}
	cursor := base64.RawURLEncoding.EncodeToString(payload)
	return &cursor
}

func decodeAuditRecordCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, cursor string) (time.Time, audit.ID, error) {
	cursor = strings.TrimSpace(cursor)
	if cursor == "" {
		return time.Time{}, audit.ID(""), nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, audit.ID(""), err
	}
	var payload auditRecordCursorPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return time.Time{}, audit.ID(""), err
	}
	if payload.Version != paginationCursorVersion || payload.Collection != "audit_records" || payload.Scope != auditRecordCursorScope(tenantID, inventoryID) || strings.TrimSpace(payload.LastID) == "" || strings.TrimSpace(payload.OccurredAt) == "" {
		return time.Time{}, audit.ID(""), ErrInvalidInput
	}
	occurredAt, err := time.Parse(time.RFC3339Nano, payload.OccurredAt)
	if err != nil {
		return time.Time{}, audit.ID(""), err
	}
	id, ok := audit.NewID(payload.LastID)
	if !ok {
		return time.Time{}, audit.ID(""), ErrInvalidInput
	}
	return occurredAt, id, nil
}

func auditRecordCursorScope(tenantID tenant.ID, inventoryID inventory.InventoryID) string {
	if inventoryID.String() == "" {
		return tenantID.String()
	}
	return tenantID.String() + ":" + inventoryID.String()
}
