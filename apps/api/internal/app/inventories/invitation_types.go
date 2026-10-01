package inventories

import (
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
