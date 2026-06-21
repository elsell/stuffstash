package ports

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type InventoryAccessRepository interface {
	InventoryAccessGrantByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, principalID identity.PrincipalID, relationship InventoryAccessRelationship) (InventoryAccessGrant, bool, error)
	ListInventoryAccessGrants(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page InventoryAccessGrantPageRequest) ([]InventoryAccessGrant, error)
	InventoryAccessInvitationByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string) (InventoryAccessInvitation, bool, error)
	ListInventoryAccessInvitations(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page InventoryAccessInvitationPageRequest) ([]InventoryAccessInvitation, error)
}

type InventoryAccessUnitOfWork interface {
	SaveInventoryAccessGrantAndEnqueue(ctx context.Context, eventID string, grant InventoryAccessGrant, auditRecord audit.Record) error
	DeleteInventoryAccessGrantAndClaimRevoke(ctx context.Context, eventID string, claimID string, leaseUntil time.Time, grant InventoryAccessGrant, auditRecord audit.Record) (AuthorizationOutboxEvent, bool, error)
	SaveInventoryAccessInvitation(ctx context.Context, invitation InventoryAccessInvitation, auditRecord audit.Record) (InventoryAccessInvitation, error)
	AcceptInventoryAccessInvitationAndEnqueue(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, tokenHash string, acceptor identity.Principal, eventID string, now time.Time, auditRecord audit.Record) (InventoryAccessInvitation, InventoryAccessGrant, error)
	RevokeInventoryAccessInvitation(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, auditRecord audit.Record) (bool, error)
	CancelInventoryAccessInvitation(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, auditRecord audit.Record) (bool, error)
	UpdateInventoryAccessInvitationExpiration(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, expiresAt time.Time, auditRecord audit.Record) (InventoryAccessInvitation, bool, error)
	DeleteInventoryAccessInvitation(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, auditRecord audit.Record) (bool, error)
}

type InventoryAccessGrant struct {
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	PrincipalID  identity.PrincipalID
	Relationship InventoryAccessRelationship
}

func (g InventoryAccessGrant) CursorKey() string {
	return g.PrincipalID.String() + ":" + string(g.Relationship)
}

func (r InventoryAccessRelationship) GrantOutboxKind() (AuthorizationOutboxEventKind, bool) {
	switch r {
	case InventoryAccessViewer:
		return AuthorizationOutboxGrantInventoryViewer, true
	case InventoryAccessEditor:
		return AuthorizationOutboxGrantInventoryEditor, true
	default:
		return "", false
	}
}

func (r InventoryAccessRelationship) RevokeOutboxKind() (AuthorizationOutboxEventKind, bool) {
	switch r {
	case InventoryAccessViewer:
		return AuthorizationOutboxRevokeInventoryViewer, true
	case InventoryAccessEditor:
		return AuthorizationOutboxRevokeInventoryEditor, true
	default:
		return "", false
	}
}

type InventoryAccessGrantPageRequest struct {
	AfterGrantKey string
	Limit         int
}

type InventoryAccessInvitationStatus string

const (
	InventoryAccessInvitationPending   InventoryAccessInvitationStatus = "pending"
	InventoryAccessInvitationAccepted  InventoryAccessInvitationStatus = "accepted"
	InventoryAccessInvitationRevoked   InventoryAccessInvitationStatus = "revoked"
	InventoryAccessInvitationCancelled InventoryAccessInvitationStatus = "cancelled"
)

type InventoryAccessInvitationStatusFilter string

const (
	InventoryAccessInvitationStatusFilterAll       InventoryAccessInvitationStatusFilter = "all"
	InventoryAccessInvitationStatusFilterPending   InventoryAccessInvitationStatusFilter = "pending"
	InventoryAccessInvitationStatusFilterAccepted  InventoryAccessInvitationStatusFilter = "accepted"
	InventoryAccessInvitationStatusFilterRevoked   InventoryAccessInvitationStatusFilter = "revoked"
	InventoryAccessInvitationStatusFilterCancelled InventoryAccessInvitationStatusFilter = "cancelled"
	InventoryAccessInvitationStatusFilterExpired   InventoryAccessInvitationStatusFilter = "expired"
)

type InventoryAccessInvitation struct {
	ID                  string
	TenantID            tenant.ID
	InventoryID         inventory.InventoryID
	Email               identity.Email
	TokenHash           string
	Relationship        InventoryAccessRelationship
	Status              InventoryAccessInvitationStatus
	InviterPrincipalID  identity.PrincipalID
	AcceptedPrincipalID identity.PrincipalID
	CreatedAt           time.Time
	ExpiresAt           time.Time
	AcceptedAt          time.Time
	RevokedAt           time.Time
}

func (i InventoryAccessInvitation) IsExpired(now time.Time) bool {
	return i.Status == InventoryAccessInvitationPending && !i.ExpiresAt.IsZero() && !i.ExpiresAt.After(now)
}

func (i InventoryAccessInvitation) CursorKey() string {
	return i.ID
}

type InventoryAccessInvitationPageRequest struct {
	AfterInvitationID string
	Limit             int
	StatusFilter      InventoryAccessInvitationStatusFilter
	Now               time.Time
}
