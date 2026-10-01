package inventories

import (
	"context"
	"crypto/subtle"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a Service) PreviewInventoryAccessInvitation(ctx context.Context, input PreviewInventoryAccessInvitationInput) (InventoryAccessInvitationPreview, error) {
	if !isValidInventoryInvitationToken(input.Token) {
		return InventoryAccessInvitationPreview{}, ErrInvitationInvalid
	}
	invitation, found, err := a.inventoryAccess.InventoryAccessInvitationByID(ctx, input.TenantID, input.InventoryID, input.InvitationID)
	if err != nil {
		return InventoryAccessInvitationPreview{}, err
	}
	if !found || !invitationTokenMatches(invitation.TokenHash, input.Token) {
		return InventoryAccessInvitationPreview{}, ErrInvitationInvalid
	}
	if input.Principal.Email.String() == "" || !strings.EqualFold(input.Principal.Email.String(), invitation.Email.String()) {
		return InventoryAccessInvitationPreview{}, ErrInvitationEmailMismatch
	}
	if invitation.Status == ports.InventoryAccessInvitationAccepted && invitation.AcceptedPrincipalID != input.Principal.ID {
		return InventoryAccessInvitationPreview{}, ErrInvitationInvalid
	}
	item, found, err := a.inventories.InventoryByID(ctx, input.TenantID, input.InventoryID)
	if err != nil {
		return InventoryAccessInvitationPreview{}, err
	}
	if !found || !item.IsActive() {
		return InventoryAccessInvitationPreview{}, ErrInvitationInvalid
	}

	now := a.clock.Now()
	preview := InventoryAccessInvitationPreview{
		InventoryID:   item.ID,
		InventoryName: item.Name.String(),
		Relationship:  invitation.Relationship,
		Status:        invitation.Status,
		ExpiresAt:     invitation.ExpiresAt,
		IsExpired:     invitation.IsExpired(now),
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventInventoryInvitationPreviewed,
		Message: "inventory invitation previewed",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"principal_id":  input.Principal.ID.String(),
			"invitation_id": invitation.ID,
			"status":        string(invitation.Status),
		},
	})
	return preview, nil
}

func invitationTokenMatches(expectedHash string, token string) bool {
	actualHash := HashInventoryInvitationToken(token)
	return len(expectedHash) == len(actualHash) && subtle.ConstantTimeCompare([]byte(expectedHash), []byte(actualHash)) == 1
}

func isValidInventoryInvitationToken(token string) bool {
	if len(token) != 43 {
		return false
	}
	for _, char := range token {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_' {
			continue
		}
		return false
	}
	return true
}

func (a Service) AcceptInventoryAccessInvitation(ctx context.Context, input AcceptInventoryAccessInvitationInput) (ports.InventoryAccessInvitation, ports.InventoryAccessGrant, error) {
	if input.Principal.Email.String() == "" {
		return ports.InventoryAccessInvitation{}, ports.InventoryAccessGrant{}, apperrors.ErrUnauthorized
	}
	if input.Token == "" {
		return ports.InventoryAccessInvitation{}, ports.InventoryAccessGrant{}, apperrors.ErrUnauthorized
	}
	item, found, err := a.inventories.InventoryByID(ctx, input.TenantID, input.InventoryID)
	if err != nil {
		return ports.InventoryAccessInvitation{}, ports.InventoryAccessGrant{}, err
	}
	if !found || !item.IsActive() {
		return ports.InventoryAccessInvitation{}, ports.InventoryAccessGrant{}, apperrors.ErrUnauthorized
	}

	auditRecord, err := appsupport.NewAuditRecord(a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionInventoryInvitationAccepted,
		TargetType:  audit.TargetInventoryInvitation,
		TargetID:    input.InvitationID,
		Metadata: map[string]string{
			"accepted_principal_id": input.Principal.ID.String(),
		},
	})
	if err != nil {
		return ports.InventoryAccessInvitation{}, ports.InventoryAccessGrant{}, err
	}

	invitation, grant, err := a.inventoryAccessUnitOfWork.AcceptInventoryAccessInvitationAndEnqueue(ctx, input.TenantID, input.InventoryID, input.InvitationID, HashInventoryInvitationToken(input.Token), input.Principal, a.ids.NewID(), a.clock.Now().UTC(), auditRecord)
	if err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return ports.InventoryAccessInvitation{}, ports.InventoryAccessGrant{}, apperrors.ErrUnauthorized
		}
		if errors.Is(err, ports.ErrConflict) {
			return ports.InventoryAccessInvitation{}, ports.InventoryAccessGrant{}, apperrors.ErrInvalidInput
		}
		return ports.InventoryAccessInvitation{}, ports.InventoryAccessGrant{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventInventoryInvitationAccepted,
		Message: "inventory invitation accepted",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"principal_id": input.Principal.ID.String(),
			"relationship": string(grant.Relationship),
			"status":       string(invitation.Status),
		},
	})
	a.DrainAuthorizationOutboxBestEffort(ctx, a.AuthorizationOutboxDrainLimit())
	return invitation, grant, nil
}
