package inventories

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a Service) GrantInventoryAccess(ctx context.Context, input GrantInventoryAccessInput) (ports.InventoryAccessGrant, error) {
	if err := a.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return ports.InventoryAccessGrant{}, err
	}

	targetPrincipalID, ok := identity.NewPrincipalID(input.TargetUserID)
	if !ok {
		return ports.InventoryAccessGrant{}, apperrors.ErrInvalidInput
	}
	if targetPrincipalID == input.Principal.ID {
		return ports.InventoryAccessGrant{}, apperrors.ErrInvalidInput
	}

	relationship, ok := inventoryAccessRelationship(input.Relationship)
	if !ok {
		return ports.InventoryAccessGrant{}, apperrors.ErrInvalidInput
	}

	grant := ports.InventoryAccessGrant{
		TenantID:     input.TenantID,
		InventoryID:  input.InventoryID,
		PrincipalID:  targetPrincipalID,
		Relationship: relationship,
	}

	auditRecord, err := appsupport.NewAuditRecord(a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:   input.Principal,
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
			return ports.InventoryAccessGrant{}, apperrors.ErrInvalidInput
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
	a.DrainAuthorizationOutboxBestEffort(ctx, a.AuthorizationOutboxDrainLimit())

	return grant, nil
}

func (a Service) RevokeInventoryAccess(ctx context.Context, input RevokeInventoryAccessInput) (bool, error) {
	if err := a.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return false, err
	}

	targetPrincipalID, ok := identity.NewPrincipalID(input.TargetUserID)
	if !ok {
		return false, apperrors.ErrInvalidInput
	}
	relationship, ok := inventoryAccessRelationship(input.Relationship)
	if !ok {
		return false, apperrors.ErrInvalidInput
	}

	grant := ports.InventoryAccessGrant{
		TenantID:     input.TenantID,
		InventoryID:  input.InventoryID,
		PrincipalID:  targetPrincipalID,
		Relationship: relationship,
	}

	auditRecord, err := appsupport.NewAuditRecord(a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:   input.Principal,
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
	event, removed, err := a.inventoryAccessUnitOfWork.DeleteInventoryAccessGrantAndClaimRevoke(ctx, eventID, claimID, a.clock.Now().Add(a.AuthorizationOutboxClaimLease()), grant, auditRecord)
	if err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return false, apperrors.ErrInvalidInput
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
	if err := a.ProcessClaimedAuthorizationOutboxEvent(ctx, event, claimID); err != nil {
		return removed, err
	}

	return removed, nil
}

func inventoryAccessRelationship(value string) (ports.InventoryAccessRelationship, bool) {
	relationship := ports.InventoryAccessRelationship(value)
	switch relationship {
	case ports.InventoryAccessViewer, ports.InventoryAccessEditor:
		return relationship, true
	default:
		return "", false
	}
}
