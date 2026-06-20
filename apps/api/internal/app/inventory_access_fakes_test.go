package app

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (f *fakeInventoryRepository) SaveInventoryAccessInvitation(_ context.Context, invitation ports.InventoryAccessInvitation, auditRecord audit.Record) (ports.InventoryAccessInvitation, error) {
	for _, existing := range f.invitations {
		if existing.TenantID == invitation.TenantID && existing.InventoryID == invitation.InventoryID && existing.Email == invitation.Email && existing.Relationship == invitation.Relationship && existing.Status == ports.InventoryAccessInvitationPending {
			return ports.InventoryAccessInvitation{}, ports.ErrConflict
		}
	}
	if invitation.ExpiresAt.IsZero() {
		return ports.InventoryAccessInvitation{}, ports.ErrConflict
	}
	invitation.Status = ports.InventoryAccessInvitationPending
	f.invitations = append(f.invitations, invitation)
	f.auditRecords = append(f.auditRecords, auditRecord)
	return invitation, nil
}

func (f *fakeInventoryRepository) AcceptInventoryAccessInvitationAndEnqueue(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, tokenHash string, acceptor identity.Principal, eventID string, auditRecord audit.Record) (ports.InventoryAccessInvitation, ports.InventoryAccessGrant, error) {
	for index, invitation := range f.invitations {
		if invitation.ID != invitationID || invitation.TenantID != tenantID || invitation.InventoryID != inventoryID || invitation.Status != ports.InventoryAccessInvitationPending || invitation.Email != acceptor.Email || invitation.TokenHash != tokenHash || invitation.ExpiresAt.IsZero() || !invitation.ExpiresAt.After(time.Now()) {
			continue
		}
		grant := ports.InventoryAccessGrant{
			TenantID:     invitation.TenantID,
			InventoryID:  invitation.InventoryID,
			PrincipalID:  acceptor.ID,
			Relationship: invitation.Relationship,
		}
		invitation.Status = ports.InventoryAccessInvitationAccepted
		invitation.AcceptedPrincipalID = acceptor.ID
		f.invitations[index] = invitation
		grantExists := false
		for _, existingGrant := range f.accessGrants {
			if existingGrant.TenantID == grant.TenantID && existingGrant.InventoryID == grant.InventoryID && existingGrant.CursorKey() == grant.CursorKey() {
				grantExists = true
				break
			}
		}
		if !grantExists {
			f.accessGrants = append(f.accessGrants, grant)
		}
		f.auditRecords = append(f.auditRecords, auditRecord)
		if f.outbox != nil && !grantExists {
			eventKind, _ := invitation.Relationship.GrantOutboxKind()
			f.outbox.events = append(f.outbox.events, ports.AuthorizationOutboxEvent{
				ID:          eventID,
				Kind:        eventKind,
				PrincipalID: acceptor.ID,
				TenantID:    invitation.TenantID,
				InventoryID: invitation.InventoryID,
			})
		}
		return invitation, grant, nil
	}
	return ports.InventoryAccessInvitation{}, ports.InventoryAccessGrant{}, ports.ErrForbidden
}

func (f *fakeInventoryRepository) RevokeInventoryAccessInvitation(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, auditRecord audit.Record) (bool, error) {
	for index, invitation := range f.invitations {
		if invitation.ID == invitationID && invitation.TenantID == tenantID && invitation.InventoryID == inventoryID && invitation.Status == ports.InventoryAccessInvitationPending {
			invitation.Status = ports.InventoryAccessInvitationRevoked
			f.invitations[index] = invitation
			f.auditRecords = append(f.auditRecords, auditRecord)
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeInventoryRepository) InventoryAccessInvitationByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string) (ports.InventoryAccessInvitation, bool, error) {
	for _, invitation := range f.invitations {
		if invitation.ID == invitationID && invitation.TenantID == tenantID && invitation.InventoryID == inventoryID {
			return invitation, true, nil
		}
	}
	return ports.InventoryAccessInvitation{}, false, nil
}

func (f *fakeInventoryRepository) CancelInventoryAccessInvitation(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, auditRecord audit.Record) (bool, error) {
	for index, invitation := range f.invitations {
		if invitation.ID == invitationID && invitation.TenantID == tenantID && invitation.InventoryID == inventoryID && invitation.Status == ports.InventoryAccessInvitationPending {
			invitation.Status = ports.InventoryAccessInvitationCancelled
			f.invitations[index] = invitation
			f.auditRecords = append(f.auditRecords, auditRecord)
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeInventoryRepository) DeleteInventoryAccessInvitation(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, auditRecord audit.Record) (bool, error) {
	for index, invitation := range f.invitations {
		if invitation.ID == invitationID && invitation.TenantID == tenantID && invitation.InventoryID == inventoryID {
			f.invitations = append(f.invitations[:index], f.invitations[index+1:]...)
			f.auditRecords = append(f.auditRecords, auditRecord)
			return true, nil
		}
	}
	return false, nil
}
