package mapper

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/access/dto"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func GrantToResponse(grant ports.InventoryAccessGrant) dto.GrantResponse {
	return dto.GrantResponse{
		TenantID:     grant.TenantID.String(),
		InventoryID:  grant.InventoryID.String(),
		PrincipalID:  grant.PrincipalID.String(),
		Relationship: string(grant.Relationship),
	}
}

func GrantsToResponse(grants []ports.InventoryAccessGrant) []dto.GrantResponse {
	data := make([]dto.GrantResponse, 0, len(grants))
	for _, grant := range grants {
		data = append(data, GrantToResponse(grant))
	}
	return data
}

func InvitationToResponse(invitation ports.InventoryAccessInvitation) dto.InvitationResponse {
	return InvitationToResponseAt(invitation, time.Now())
}

func InvitationToResponseAt(invitation ports.InventoryAccessInvitation, now time.Time) dto.InvitationResponse {
	return dto.InvitationResponse{
		ID:                  invitation.ID,
		TenantID:            invitation.TenantID.String(),
		InventoryID:         invitation.InventoryID.String(),
		Email:               invitation.Email.String(),
		Relationship:        string(invitation.Relationship),
		Status:              string(invitation.Status),
		InviterPrincipalID:  invitation.InviterPrincipalID.String(),
		AcceptedPrincipalID: invitation.AcceptedPrincipalID.String(),
		ExpiresAt:           invitation.ExpiresAt.Format(time.RFC3339),
		IsExpired:           invitation.IsExpired(now),
	}
}

func InvitationsToResponse(invitations []ports.InventoryAccessInvitation) []dto.InvitationResponse {
	return InvitationsToResponseAt(invitations, time.Now())
}

func InvitationsToResponseAt(invitations []ports.InventoryAccessInvitation, now time.Time) []dto.InvitationResponse {
	data := make([]dto.InvitationResponse, 0, len(invitations))
	for _, invitation := range invitations {
		data = append(data, InvitationToResponseAt(invitation, now))
	}
	return data
}

func InvitationAcceptanceToResponse(invitation ports.InventoryAccessInvitation, grant ports.InventoryAccessGrant) dto.InvitationAcceptanceResponse {
	return dto.InvitationAcceptanceResponse{
		Invitation: InvitationToResponse(invitation),
		Grant:      GrantToResponse(grant),
	}
}
