package inventories

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a Service) RevokeInventoryAccessInvitation(ctx context.Context, input RevokeInventoryAccessInvitationInput) (bool, error) {
	if err := a.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return false, err
	}
	auditRecord, err := appsupport.NewAuditRecord(a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionInventoryInvitationRevoked,
		TargetType:  audit.TargetInventoryInvitation,
		TargetID:    input.InvitationID,
		Metadata:    map[string]string{},
	})
	if err != nil {
		return false, err
	}
	revoked, err := a.inventoryAccessUnitOfWork.RevokeInventoryAccessInvitation(ctx, input.TenantID, input.InventoryID, input.InvitationID, auditRecord)
	if err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return false, apperrors.ErrInvalidInput
		}
		return false, err
	}
	if revoked {
		a.observer.Record(ctx, ports.Event{
			Name:    ports.EventInventoryInvitationRevoked,
			Message: "inventory invitation revoked",
			Fields: map[string]string{
				"tenant_id":     input.TenantID.String(),
				"inventory_id":  input.InventoryID.String(),
				"principal_id":  input.Principal.ID.String(),
				"invitation_id": input.InvitationID,
				"result_status": string(ports.InventoryAccessInvitationRevoked),
			},
		})
	}
	return revoked, nil
}

func (a Service) UpdateInventoryAccessInvitationExpiration(ctx context.Context, input UpdateInventoryAccessInvitationExpirationInput) (ports.InventoryAccessInvitation, error) {
	if err := a.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return ports.InventoryAccessInvitation{}, err
	}
	if input.ExpiresAt.IsZero() {
		return ports.InventoryAccessInvitation{}, apperrors.ErrInvalidInput
	}
	auditRecord, err := appsupport.NewAuditRecord(a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionInventoryInvitationExpirationUpdated,
		TargetType:  audit.TargetInventoryInvitation,
		TargetID:    input.InvitationID,
		Metadata: map[string]string{
			"expires_at": input.ExpiresAt.UTC().Format(time.RFC3339),
		},
	})
	if err != nil {
		return ports.InventoryAccessInvitation{}, err
	}
	invitation, updated, err := a.inventoryAccessUnitOfWork.UpdateInventoryAccessInvitationExpiration(ctx, input.TenantID, input.InventoryID, input.InvitationID, input.ExpiresAt, auditRecord)
	if err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return ports.InventoryAccessInvitation{}, apperrors.ErrInvalidInput
		}
		return ports.InventoryAccessInvitation{}, err
	}
	if !updated {
		return ports.InventoryAccessInvitation{}, apperrors.ErrNotFound
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventInventoryInvitationExpirationUpdated,
		Message: "inventory invitation expiration updated",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"principal_id":  input.Principal.ID.String(),
			"invitation_id": invitation.ID,
			"expires_at":    invitation.ExpiresAt.UTC().Format(time.RFC3339),
		},
	})
	return invitation, nil
}

func (a Service) CancelInventoryAccessInvitation(ctx context.Context, input RevokeInventoryAccessInvitationInput) (bool, error) {
	if err := a.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return false, err
	}
	auditRecord, err := appsupport.NewAuditRecord(a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionInventoryInvitationCancelled,
		TargetType:  audit.TargetInventoryInvitation,
		TargetID:    input.InvitationID,
		Metadata:    map[string]string{},
	})
	if err != nil {
		return false, err
	}
	cancelled, err := a.inventoryAccessUnitOfWork.CancelInventoryAccessInvitation(ctx, input.TenantID, input.InventoryID, input.InvitationID, auditRecord)
	if err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return false, apperrors.ErrInvalidInput
		}
		return false, err
	}
	if cancelled {
		a.observer.Record(ctx, ports.Event{
			Name:    ports.EventInventoryInvitationCancelled,
			Message: "inventory invitation cancelled",
			Fields: map[string]string{
				"tenant_id":     input.TenantID.String(),
				"inventory_id":  input.InventoryID.String(),
				"principal_id":  input.Principal.ID.String(),
				"invitation_id": input.InvitationID,
				"result_status": string(ports.InventoryAccessInvitationCancelled),
			},
		})
	}
	return cancelled, nil
}

func (a Service) DeleteInventoryAccessInvitation(ctx context.Context, input RevokeInventoryAccessInvitationInput) (bool, error) {
	if err := a.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return false, err
	}
	auditRecord, err := appsupport.NewAuditRecord(a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionInventoryInvitationDeleted,
		TargetType:  audit.TargetInventoryInvitation,
		TargetID:    input.InvitationID,
		Metadata:    map[string]string{},
	})
	if err != nil {
		return false, err
	}
	deleted, err := a.inventoryAccessUnitOfWork.DeleteInventoryAccessInvitation(ctx, input.TenantID, input.InventoryID, input.InvitationID, auditRecord)
	if err != nil {
		return false, err
	}
	if deleted {
		a.observer.Record(ctx, ports.Event{
			Name:    ports.EventInventoryInvitationDeleted,
			Message: "inventory invitation deleted",
			Fields: map[string]string{
				"tenant_id":     input.TenantID.String(),
				"inventory_id":  input.InventoryID.String(),
				"principal_id":  input.Principal.ID.String(),
				"invitation_id": input.InvitationID,
			},
		})
	}
	return deleted, nil
}
