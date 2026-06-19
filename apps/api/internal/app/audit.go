package app

import (
	"context"
	"strconv"
	"time"

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

type ListAuditRecordsResult struct {
	Items      []audit.Record
	Limit      int
	NextCursor *string
	HasMore    bool
}

type auditRecordInput struct {
	PrincipalID identity.PrincipalID
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Source      audit.Source
	Action      audit.Action
	TargetType  audit.TargetType
	TargetID    string
	Metadata    map[string]string
}

func (a App) newAuditRecord(input auditRecordInput) (audit.Record, error) {
	id, ok := audit.NewID(a.ids.NewID())
	if !ok {
		return audit.Record{}, ErrInvalidInput
	}
	source := input.Source
	if source.String() == "" {
		source = audit.SourceAPI
	}
	record, ok := audit.NewRecord(
		id,
		audit.TenantID(input.TenantID.String()),
		audit.InventoryID(input.InventoryID.String()),
		audit.PrincipalID(input.PrincipalID.String()),
		input.Action,
		source,
		input.TargetType,
		input.TargetID,
		time.Now(),
		"",
		input.Metadata,
	)
	if !ok {
		return audit.Record{}, ErrInvalidInput
	}
	return record, nil
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
	afterRecordID, err := decodeAuditRecordCursor(input.TenantID, input.InventoryID, input.Cursor)
	if err != nil {
		return ListAuditRecordsResult{}, ErrInvalidInput
	}
	items, err := a.audit.ListTenantAuditRecords(ctx, input.TenantID, ports.AuditRecordPageRequest{
		AfterRecordID: afterRecordID,
		Limit:         limit + 1,
	})
	if err != nil {
		return ListAuditRecordsResult{}, err
	}
	return a.auditRecordListResult(ctx, input, items, limit), nil
}

func (a App) ListInventoryAuditRecords(ctx context.Context, input ListAuditRecordsInput) (ListAuditRecordsResult, error) {
	if err := a.ensureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return ListAuditRecordsResult{}, err
	}

	limit := pageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	afterRecordID, err := decodeAuditRecordCursor(input.TenantID, input.InventoryID, input.Cursor)
	if err != nil {
		return ListAuditRecordsResult{}, ErrInvalidInput
	}
	items, err := a.audit.ListInventoryAuditRecords(ctx, input.TenantID, input.InventoryID, ports.AuditRecordPageRequest{
		AfterRecordID: afterRecordID,
		Limit:         limit + 1,
	})
	if err != nil {
		return ListAuditRecordsResult{}, err
	}
	return a.auditRecordListResult(ctx, input, items, limit), nil
}

func (a App) auditRecordListResult(ctx context.Context, input ListAuditRecordsInput, items []audit.Record, limit int) ListAuditRecordsResult {
	hasMore := len(items) > limit
	var nextCursor *string
	if hasMore {
		items = items[:limit]
		nextCursor = encodeAuditRecordCursor(input.TenantID, input.InventoryID, items[len(items)-1].ID)
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

	return ListAuditRecordsResult{
		Items:      items,
		Limit:      limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}
}

func encodeAuditRecordCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, id audit.ID) *string {
	return encodePageCursor("audit_records", auditRecordCursorScope(tenantID, inventoryID), id.String())
}

func decodeAuditRecordCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, cursor string) (audit.ID, error) {
	decoded, err := decodePageCursor("audit_records", auditRecordCursorScope(tenantID, inventoryID), cursor)
	if err != nil {
		return audit.ID(""), err
	}
	if decoded == "" {
		return audit.ID(""), nil
	}
	id, ok := audit.NewID(decoded)
	if !ok {
		return audit.ID(""), ErrInvalidInput
	}
	return id, nil
}

func auditRecordCursorScope(tenantID tenant.ID, inventoryID inventory.InventoryID) string {
	if inventoryID.String() == "" {
		return tenantID.String()
	}
	return tenantID.String() + ":" + inventoryID.String()
}
