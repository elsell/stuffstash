package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
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

type CreateInventoryAccessInvitationInput struct {
	Principal    identity.Principal
	Source       audit.Source
	RequestID    string
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	Email        string
	Relationship string
}

type CreateInventoryAccessInvitationResult struct {
	Invitation      ports.InventoryAccessInvitation
	AcceptanceToken string
}

type AcceptInventoryAccessInvitationInput struct {
	Principal    identity.Principal
	Source       audit.Source
	RequestID    string
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	InvitationID string
	Token        string
}

type RevokeInventoryAccessInvitationInput struct {
	Principal    identity.Principal
	Source       audit.Source
	RequestID    string
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	InvitationID string
}

type GetInventoryAccessInvitationInput struct {
	Principal    identity.Principal
	Source       audit.Source
	RequestID    string
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	InvitationID string
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

	if err := a.inventories.SaveInventoryAccessGrantAndEnqueue(ctx, a.ids.NewID(), grant, auditRecord); err != nil {
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
	event, removed, err := a.inventories.DeleteInventoryAccessGrantAndClaimRevoke(ctx, eventID, claimID, time.Now().Add(a.authorizationOutboxClaimLease()), grant, auditRecord)
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

	items, err := a.inventories.ListInventoryAccessGrants(ctx, input.TenantID, input.InventoryID, ports.InventoryAccessGrantPageRequest{
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
	grant, found, err := a.inventories.InventoryAccessGrantByID(ctx, input.TenantID, input.InventoryID, targetPrincipalID, relationship)
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

func (a App) CreateInventoryAccessInvitation(ctx context.Context, input CreateInventoryAccessInvitationInput) (CreateInventoryAccessInvitationResult, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return CreateInventoryAccessInvitationResult{}, err
	}
	email, ok := identity.NewEmail(input.Email)
	if !ok {
		return CreateInventoryAccessInvitationResult{}, ErrInvalidInput
	}
	relationship, ok := inventoryAccessRelationship(input.Relationship)
	if !ok {
		return CreateInventoryAccessInvitationResult{}, ErrInvalidInput
	}
	acceptanceToken, err := newInventoryInvitationToken()
	if err != nil {
		return CreateInventoryAccessInvitationResult{}, err
	}

	invitation := ports.InventoryAccessInvitation{
		ID:                 a.ids.NewID(),
		TenantID:           input.TenantID,
		InventoryID:        input.InventoryID,
		Email:              email,
		TokenHash:          hashInventoryInvitationToken(acceptanceToken),
		Relationship:       relationship,
		Status:             ports.InventoryAccessInvitationPending,
		InviterPrincipalID: input.Principal.ID,
		ExpiresAt:          time.Now().Add(a.invitationTTL),
	}
	auditRecord, err := a.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionInventoryInvitationCreated,
		TargetType:  audit.TargetInventoryInvitation,
		TargetID:    invitation.ID,
		Metadata: map[string]string{
			"relationship": string(relationship),
		},
	})
	if err != nil {
		return CreateInventoryAccessInvitationResult{}, err
	}

	saved, err := a.inventories.SaveInventoryAccessInvitation(ctx, invitation, auditRecord)
	if err != nil {
		if errors.Is(err, ports.ErrForbidden) || errors.Is(err, ports.ErrConflict) {
			return CreateInventoryAccessInvitationResult{}, ErrInvalidInput
		}
		return CreateInventoryAccessInvitationResult{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventInventoryInvitationCreated,
		Message: "inventory invitation created",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"principal_id": input.Principal.ID.String(),
			"relationship": string(relationship),
			"status":       string(saved.Status),
		},
	})
	return CreateInventoryAccessInvitationResult{
		Invitation:      saved,
		AcceptanceToken: acceptanceToken,
	}, nil
}

func (a App) AcceptInventoryAccessInvitation(ctx context.Context, input AcceptInventoryAccessInvitationInput) (ports.InventoryAccessInvitation, ports.InventoryAccessGrant, error) {
	if input.Principal.Email.String() == "" {
		return ports.InventoryAccessInvitation{}, ports.InventoryAccessGrant{}, ErrUnauthorized
	}
	if input.Token == "" {
		return ports.InventoryAccessInvitation{}, ports.InventoryAccessGrant{}, ErrUnauthorized
	}
	item, found, err := a.inventories.InventoryByID(ctx, input.TenantID, input.InventoryID)
	if err != nil {
		return ports.InventoryAccessInvitation{}, ports.InventoryAccessGrant{}, err
	}
	if !found || !item.IsActive() {
		return ports.InventoryAccessInvitation{}, ports.InventoryAccessGrant{}, ErrUnauthorized
	}

	auditRecord, err := a.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
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

	invitation, grant, err := a.inventories.AcceptInventoryAccessInvitationAndEnqueue(ctx, input.TenantID, input.InventoryID, input.InvitationID, hashInventoryInvitationToken(input.Token), input.Principal, a.ids.NewID(), auditRecord)
	if err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return ports.InventoryAccessInvitation{}, ports.InventoryAccessGrant{}, ErrUnauthorized
		}
		if errors.Is(err, ports.ErrConflict) {
			return ports.InventoryAccessInvitation{}, ports.InventoryAccessGrant{}, ErrInvalidInput
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
	a.drainAuthorizationOutboxBestEffort(ctx, a.authorizationOutboxDrainLimit())
	return invitation, grant, nil
}

func (a App) RevokeInventoryAccessInvitation(ctx context.Context, input RevokeInventoryAccessInvitationInput) (bool, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return false, err
	}
	auditRecord, err := a.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
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
	revoked, err := a.inventories.RevokeInventoryAccessInvitation(ctx, input.TenantID, input.InventoryID, input.InvitationID, auditRecord)
	if err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return false, ErrInvalidInput
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

func (a App) GetInventoryAccessInvitation(ctx context.Context, input GetInventoryAccessInvitationInput) (ports.InventoryAccessInvitation, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return ports.InventoryAccessInvitation{}, err
	}
	invitation, found, err := a.inventories.InventoryAccessInvitationByID(ctx, input.TenantID, input.InventoryID, input.InvitationID)
	if err != nil {
		return ports.InventoryAccessInvitation{}, err
	}
	if !found {
		return ports.InventoryAccessInvitation{}, ErrNotFound
	}
	if err := a.saveReadAuditRecord(ctx, auditRecordInput{
		PrincipalID: input.Principal.ID,
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

func (a App) CancelInventoryAccessInvitation(ctx context.Context, input RevokeInventoryAccessInvitationInput) (bool, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return false, err
	}
	auditRecord, err := a.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
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
	cancelled, err := a.inventories.CancelInventoryAccessInvitation(ctx, input.TenantID, input.InventoryID, input.InvitationID, auditRecord)
	if err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return false, ErrInvalidInput
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

func (a App) DeleteInventoryAccessInvitation(ctx context.Context, input RevokeInventoryAccessInvitationInput) (bool, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return false, err
	}
	auditRecord, err := a.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
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
	deleted, err := a.inventories.DeleteInventoryAccessInvitation(ctx, input.TenantID, input.InventoryID, input.InvitationID, auditRecord)
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

func inventoryAccessRelationship(value string) (ports.InventoryAccessRelationship, bool) {
	relationship := ports.InventoryAccessRelationship(value)
	switch relationship {
	case ports.InventoryAccessViewer, ports.InventoryAccessEditor:
		return relationship, true
	default:
		return "", false
	}
}

func encodeInventoryAccessGrantCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, key string) *string {
	return encodePageCursor("inventory_access_grants", tenantID.String()+":"+inventoryID.String(), key)
}

func decodeInventoryAccessGrantCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, cursor string) (string, error) {
	return decodePageCursor("inventory_access_grants", tenantID.String()+":"+inventoryID.String(), cursor)
}

func newInventoryInvitationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func hashInventoryInvitationToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
