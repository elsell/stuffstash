package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

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
	Invitation ports.InventoryAccessInvitation
	InviteURL  string
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

type PreviewInventoryAccessInvitationInput struct {
	Principal    identity.Principal
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	InvitationID string
	Token        string
}

type InventoryAccessInvitationPreview struct {
	InventoryID   inventory.InventoryID
	InventoryName string
	Relationship  ports.InventoryAccessRelationship
	Status        ports.InventoryAccessInvitationStatus
	ExpiresAt     time.Time
	IsExpired     bool
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

type ListInventoryAccessInvitationsInput struct {
	Principal    identity.Principal
	Source       audit.Source
	RequestID    string
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	Limit        int
	Cursor       string
	StatusFilter string
}

type ListInventoryAccessInvitationsResult struct {
	Items      []ports.InventoryAccessInvitation
	Limit      int
	NextCursor *string
	HasMore    bool
	Now        time.Time
}

type UpdateInventoryAccessInvitationExpirationInput struct {
	Principal    identity.Principal
	Source       audit.Source
	RequestID    string
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	InvitationID string
	ExpiresAt    time.Time
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
		ExpiresAt:          a.clock.Now().Add(a.invitationTTL),
	}
	inviteURL, err := buildInventoryInvitationURL(a.invitationPublicBaseURL, invitation, acceptanceToken, a.invitationAllowInsecureHTTP)
	if err != nil {
		return CreateInventoryAccessInvitationResult{}, err
	}
	auditRecord, err := a.newAuditRecord(auditRecordInput{
		Principal:   input.Principal,
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

	saved, err := a.inventoryAccessUnitOfWork.SaveInventoryAccessInvitation(ctx, invitation, auditRecord)
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
		Invitation: saved,
		InviteURL:  inviteURL,
	}, nil
}

func normalizeInvitationPublicBaseURL(value string) string {
	return strings.TrimSpace(value)
}

func buildInventoryInvitationURL(baseURL string, invitation ports.InventoryAccessInvitation, acceptanceToken string, allowInsecureLocalHTTP bool) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("invalid invitation public base URL")
	}
	if parsed.Scheme != "https" && !(parsed.Scheme == "http" && allowInsecureLocalHTTP && isLocalInvitationHost(parsed.Hostname())) {
		return "", errors.New("invitation public base URL must use HTTPS")
	}
	parsed.Path = "/invitations/accept"
	parsed.RawPath = ""
	query := parsed.Query()
	query.Set("tenant", invitation.TenantID.String())
	query.Set("inventory", invitation.InventoryID.String())
	query.Set("invitation", invitation.ID)
	parsed.RawQuery = query.Encode()
	parsed.Fragment = "token=" + acceptanceToken
	return parsed.String(), nil
}

func isLocalInvitationHost(host string) bool {
	if host == "localhost" {
		return true
	}
	address, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	if address.IsLoopback() {
		return true
	}
	if !address.Is4() {
		return false
	}
	octets := address.As4()
	return octets[0] == 10 ||
		(octets[0] == 172 && octets[1] >= 16 && octets[1] <= 31) ||
		(octets[0] == 192 && octets[1] == 168)
}

func (a App) PreviewInventoryAccessInvitation(ctx context.Context, input PreviewInventoryAccessInvitationInput) (InventoryAccessInvitationPreview, error) {
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
	actualHash := hashInventoryInvitationToken(token)
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

	invitation, grant, err := a.inventoryAccessUnitOfWork.AcceptInventoryAccessInvitationAndEnqueue(ctx, input.TenantID, input.InventoryID, input.InvitationID, hashInventoryInvitationToken(input.Token), input.Principal, a.ids.NewID(), a.clock.Now().UTC(), auditRecord)
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
	invitation, found, err := a.inventoryAccess.InventoryAccessInvitationByID(ctx, input.TenantID, input.InventoryID, input.InvitationID)
	if err != nil {
		return ports.InventoryAccessInvitation{}, err
	}
	if !found {
		return ports.InventoryAccessInvitation{}, ErrNotFound
	}
	if err := a.saveReadAuditRecord(ctx, auditRecordInput{
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

func (a App) ListInventoryAccessInvitations(ctx context.Context, input ListInventoryAccessInvitationsInput) (ListInventoryAccessInvitationsResult, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return ListInventoryAccessInvitationsResult{}, err
	}

	limit := pageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	afterInvitationID, err := decodeInventoryAccessInvitationCursor(input.TenantID, input.InventoryID, input.Cursor)
	if err != nil {
		return ListInventoryAccessInvitationsResult{}, ErrInvalidInput
	}
	statusFilter, ok := inventoryAccessInvitationStatusFilter(input.StatusFilter)
	if !ok {
		return ListInventoryAccessInvitationsResult{}, ErrInvalidInput
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
	if err := a.saveReadAuditRecord(ctx, auditRecordInput{
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

func (a App) UpdateInventoryAccessInvitationExpiration(ctx context.Context, input UpdateInventoryAccessInvitationExpirationInput) (ports.InventoryAccessInvitation, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return ports.InventoryAccessInvitation{}, err
	}
	if input.ExpiresAt.IsZero() {
		return ports.InventoryAccessInvitation{}, ErrInvalidInput
	}
	auditRecord, err := a.newAuditRecord(auditRecordInput{
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
			return ports.InventoryAccessInvitation{}, ErrInvalidInput
		}
		return ports.InventoryAccessInvitation{}, err
	}
	if !updated {
		return ports.InventoryAccessInvitation{}, ErrNotFound
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

func (a App) CancelInventoryAccessInvitation(ctx context.Context, input RevokeInventoryAccessInvitationInput) (bool, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return false, err
	}
	auditRecord, err := a.newAuditRecord(auditRecordInput{
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

func encodeInventoryAccessInvitationCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, key string) *string {
	return encodePageCursor("inventory_access_invitations", tenantID.String()+":"+inventoryID.String(), key)
}

func decodeInventoryAccessInvitationCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, cursor string) (string, error) {
	return decodePageCursor("inventory_access_invitations", tenantID.String()+":"+inventoryID.String(), cursor)
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
