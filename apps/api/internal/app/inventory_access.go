package app

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type GrantInventoryAccessInput struct {
	Principal    identity.Principal
	Source       audit.Source
	RequestID    string
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	TargetUserID string
	Relationship string
}

type ListInventoryAccessGrantsInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Limit       int
	Cursor      string
}

type GetInventoryAccessGrantInput struct {
	Principal    identity.Principal
	Source       audit.Source
	RequestID    string
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	TargetUserID string
	Relationship string
}

type ListInventoryAccessGrantsResult struct {
	Items      []ports.InventoryAccessGrant
	Limit      int
	NextCursor *string
	HasMore    bool
}

type RevokeInventoryAccessInput struct {
	Principal    identity.Principal
	Source       audit.Source
	RequestID    string
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	TargetUserID string
	Relationship string
}

func (a App) GrantInventoryAccess(ctx context.Context, input GrantInventoryAccessInput) (ports.InventoryAccessGrant, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return ports.InventoryAccessGrant{}, err
	}

	targetPrincipalID, ok := identity.NewPrincipalID(input.TargetUserID)
	if !ok {
		return ports.InventoryAccessGrant{}, ErrInvalidInput
	}
	if targetPrincipalID == input.Principal.ID {
		return ports.InventoryAccessGrant{}, ErrInvalidInput
	}

	relationship, ok := inventoryAccessRelationship(input.Relationship)
	if !ok {
		return ports.InventoryAccessGrant{}, ErrInvalidInput
	}

	grant := ports.InventoryAccessGrant{
		TenantID:     input.TenantID,
		InventoryID:  input.InventoryID,
		PrincipalID:  targetPrincipalID,
		Relationship: relationship,
	}

	auditRecord, err := a.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionInventoryAccessGranted,
		TargetType:  audit.TargetInventoryAccessGrant,
		TargetID:    grant.CursorKey(),
		Metadata: map[string]string{
			"target_principal_id": targetPrincipalID.String(),
			"relationship":        string(relationship),
		},
	})
	if err != nil {
		return ports.InventoryAccessGrant{}, err
	}

	if err := a.inventoryAccessUnitOfWork.SaveInventoryAccessGrantAndEnqueue(ctx, a.ids.NewID(), grant, auditRecord); err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return ports.InventoryAccessGrant{}, ErrInvalidInput
		}
		return ports.InventoryAccessGrant{}, err
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventInventoryAccessGranted,
		Message: "inventory access granted",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"principal_id": input.Principal.ID.String(),
			"target_id":    targetPrincipalID.String(),
			"relationship": string(relationship),
		},
	})
	a.drainAuthorizationOutboxBestEffort(ctx, a.authorizationOutboxDrainLimit())

	return grant, nil
}

func (a App) RevokeInventoryAccess(ctx context.Context, input RevokeInventoryAccessInput) (bool, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return false, err
	}

	targetPrincipalID, ok := identity.NewPrincipalID(input.TargetUserID)
	if !ok {
		return false, ErrInvalidInput
	}
	relationship, ok := inventoryAccessRelationship(input.Relationship)
	if !ok {
		return false, ErrInvalidInput
	}

	grant := ports.InventoryAccessGrant{
		TenantID:     input.TenantID,
		InventoryID:  input.InventoryID,
		PrincipalID:  targetPrincipalID,
		Relationship: relationship,
	}

	auditRecord, err := a.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionInventoryAccessRevoked,
		TargetType:  audit.TargetInventoryAccessGrant,
		TargetID:    grant.CursorKey(),
		Metadata: map[string]string{
			"target_principal_id": targetPrincipalID.String(),
			"relationship":        string(relationship),
		},
	})
	if err != nil {
		return false, err
	}

	eventID := a.ids.NewID()
	claimID := a.ids.NewID()
	event, removed, err := a.inventoryAccessUnitOfWork.DeleteInventoryAccessGrantAndClaimRevoke(ctx, eventID, claimID, time.Now().Add(a.authorizationOutboxClaimLease()), grant, auditRecord)
	if err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return false, ErrInvalidInput
		}
		return false, err
	}
	if removed {
		a.observer.Record(ctx, ports.Event{
			Name:    ports.EventInventoryAccessRevoked,
			Message: "inventory access revoked",
			Fields: map[string]string{
				"tenant_id":    input.TenantID.String(),
				"inventory_id": input.InventoryID.String(),
				"principal_id": input.Principal.ID.String(),
				"target_id":    targetPrincipalID.String(),
				"relationship": string(relationship),
			},
		})
	}
	if err := a.processClaimedAuthorizationOutboxEvent(ctx, event, claimID); err != nil {
		return removed, err
	}

	return removed, nil
}

func (a App) ListInventoryAccessGrants(ctx context.Context, input ListInventoryAccessGrantsInput) (ListInventoryAccessGrantsResult, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return ListInventoryAccessGrantsResult{}, err
	}

	limit := pageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	afterGrantKey, err := decodeInventoryAccessGrantCursor(input.TenantID, input.InventoryID, input.Cursor)
	if err != nil {
		return ListInventoryAccessGrantsResult{}, ErrInvalidInput
	}

	items, err := a.inventoryAccess.ListInventoryAccessGrants(ctx, input.TenantID, input.InventoryID, ports.InventoryAccessGrantPageRequest{
		AfterGrantKey: afterGrantKey,
		Limit:         limit + 1,
	})
	if err != nil {
		return ListInventoryAccessGrantsResult{}, err
	}

	hasMore := len(items) > limit
	var nextCursor *string
	if hasMore {
		items = items[:limit]
		nextCursor = encodeInventoryAccessGrantCursor(input.TenantID, input.InventoryID, items[len(items)-1].CursorKey())
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventInventoryAccessListed,
		Message: "inventory access grants listed",
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
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionInventoryAccessGrantListed,
		TargetType:  audit.TargetInventory,
		TargetID:    input.InventoryID.String(),
		Metadata: map[string]string{
			"limit": strconv.Itoa(limit),
		},
	}); err != nil {
		return ListInventoryAccessGrantsResult{}, err
	}

	return ListInventoryAccessGrantsResult{
		Items:      items,
		Limit:      limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (a App) GetInventoryAccessGrant(ctx context.Context, input GetInventoryAccessGrantInput) (ports.InventoryAccessGrant, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return ports.InventoryAccessGrant{}, err
	}
	targetPrincipalID, ok := identity.NewPrincipalID(input.TargetUserID)
	if !ok {
		return ports.InventoryAccessGrant{}, ErrInvalidInput
	}
	relationship, ok := inventoryAccessRelationship(input.Relationship)
	if !ok {
		return ports.InventoryAccessGrant{}, ErrInvalidInput
	}
	grant, found, err := a.inventoryAccess.InventoryAccessGrantByID(ctx, input.TenantID, input.InventoryID, targetPrincipalID, relationship)
	if err != nil {
		return ports.InventoryAccessGrant{}, err
	}
	if !found {
		return ports.InventoryAccessGrant{}, ErrNotFound
	}
	if err := a.saveReadAuditRecord(ctx, auditRecordInput{
		PrincipalID: input.Principal.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionInventoryAccessGrantViewed,
		TargetType:  audit.TargetInventoryAccessGrant,
		TargetID:    grant.CursorKey(),
		Metadata: map[string]string{
			"target_principal_id": targetPrincipalID.String(),
			"relationship":        string(relationship),
		},
	}); err != nil {
		return ports.InventoryAccessGrant{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventInventoryAccessViewed,
		Message: "inventory access grant viewed",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"principal_id": input.Principal.ID.String(),
			"target_id":    targetPrincipalID.String(),
			"relationship": string(relationship),
		},
	})
	return grant, nil
}
