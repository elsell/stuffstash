package inventories

import (
	"context"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a Service) GetInventoryAccessInvitation(ctx context.Context, input GetInventoryAccessInvitationInput) (ports.InventoryAccessInvitation, error) {
	if err := a.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return ports.InventoryAccessInvitation{}, err
	}
	invitation, found, err := a.inventoryAccess.InventoryAccessInvitationByID(ctx, input.TenantID, input.InventoryID, input.InvitationID)
	if err != nil {
		return ports.InventoryAccessInvitation{}, err
	}
	if !found {
		return ports.InventoryAccessInvitation{}, apperrors.ErrNotFound
	}
	if err := appsupport.SaveReadAuditRecord(ctx, a.audit, a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionInventoryInvitationViewed,
		TargetType:  audit.TargetInventoryInvitation,
		TargetID:    invitation.ID,
		Metadata: map[string]string{
			"relationship": string(invitation.Relationship),
			"status":       string(invitation.Status),
		},
	}); err != nil {
		return ports.InventoryAccessInvitation{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventInventoryInvitationViewed,
		Message: "inventory invitation viewed",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"principal_id":  input.Principal.ID.String(),
			"invitation_id": invitation.ID,
			"status":        string(invitation.Status),
		},
	})
	return invitation, nil
}

func (a Service) ListInventoryAccessInvitations(ctx context.Context, input ListInventoryAccessInvitationsInput) (ListInventoryAccessInvitationsResult, error) {
	if err := a.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return ListInventoryAccessInvitationsResult{}, err
	}

	limit := appsupport.PageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	afterInvitationID, err := decodeInventoryAccessInvitationCursor(input.TenantID, input.InventoryID, input.Cursor)
	if err != nil {
		return ListInventoryAccessInvitationsResult{}, apperrors.ErrInvalidInput
	}
	statusFilter, ok := inventoryAccessInvitationStatusFilter(input.StatusFilter)
	if !ok {
		return ListInventoryAccessInvitationsResult{}, apperrors.ErrInvalidInput
	}

	now := a.clock.Now()
	items, err := a.inventoryAccess.ListInventoryAccessInvitations(ctx, input.TenantID, input.InventoryID, ports.InventoryAccessInvitationPageRequest{
		AfterInvitationID: afterInvitationID,
		Limit:             limit + 1,
		StatusFilter:      statusFilter,
		Now:               now,
	})
	if err != nil {
		return ListInventoryAccessInvitationsResult{}, err
	}

	hasMore := len(items) > limit
	var nextCursor *string
	if hasMore {
		items = items[:limit]
		nextCursor = encodeInventoryAccessInvitationCursor(input.TenantID, input.InventoryID, items[len(items)-1].CursorKey())
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventInventoryInvitationListed,
		Message: "inventory invitations listed",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"principal_id":  input.Principal.ID.String(),
			"limit":         strconv.Itoa(limit),
			"status_filter": string(statusFilter),
		},
	})
	if err := appsupport.SaveReadAuditRecord(ctx, a.audit, a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionInventoryInvitationListed,
		TargetType:  audit.TargetInventory,
		TargetID:    input.InventoryID.String(),
		Metadata: map[string]string{
			"limit":         strconv.Itoa(limit),
			"status_filter": string(statusFilter),
		},
	}); err != nil {
		return ListInventoryAccessInvitationsResult{}, err
	}

	return ListInventoryAccessInvitationsResult{
		Items:      items,
		Limit:      limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Now:        now,
	}, nil
}

func encodeInventoryAccessInvitationCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, key string) *string {
	return appsupport.EncodePageCursor("inventory_access_invitations", tenantID.String()+":"+inventoryID.String(), key)
}

func decodeInventoryAccessInvitationCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, cursor string) (string, error) {
	return appsupport.DecodePageCursor("inventory_access_invitations", tenantID.String()+":"+inventoryID.String(), cursor)
}

func inventoryAccessInvitationStatusFilter(value string) (ports.InventoryAccessInvitationStatusFilter, bool) {
	if value == "" {
		return ports.InventoryAccessInvitationStatusFilterAll, true
	}
	filter := ports.InventoryAccessInvitationStatusFilter(value)
	switch filter {
	case ports.InventoryAccessInvitationStatusFilterAll,
		ports.InventoryAccessInvitationStatusFilterPending,
		ports.InventoryAccessInvitationStatusFilterAccepted,
		ports.InventoryAccessInvitationStatusFilterRevoked,
		ports.InventoryAccessInvitationStatusFilterCancelled,
		ports.InventoryAccessInvitationStatusFilterExpired:
		return filter, true
	default:
		return "", false
	}
}
