package gormstore

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type inventoryAccessInvitationModel struct {
	ID                  string         `gorm:"primaryKey;size:64"`
	TenantID            string         `gorm:"not null;size:26;index:idx_inventory_access_invitations_pending,priority:1"`
	Tenant              tenantModel    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID         string         `gorm:"not null;size:26;index:idx_inventory_access_invitations_pending,priority:2"`
	Inventory           inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:InventoryID;references:ID"`
	Email               string         `gorm:"not null;size:320;index:idx_inventory_access_invitations_pending,priority:3"`
	TokenHash           string         `gorm:"not null;size:128"`
	Relationship        string         `gorm:"not null;size:32;index:idx_inventory_access_invitations_pending,priority:4;check:chk_inventory_access_invitations_relationship,relationship IN ('viewer','editor')"`
	Status              string         `gorm:"not null;size:32;index:idx_inventory_access_invitations_pending,priority:5;check:chk_inventory_access_invitations_status,status IN ('pending','accepted','revoked','cancelled')"`
	InviterPrincipalID  string         `gorm:"not null;size:128;index"`
	AcceptedPrincipalID string         `gorm:"size:128;index"`
	ExpiresAt           time.Time      `gorm:"not null;index"`
	AcceptedAt          *time.Time
	RevokedAt           *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (inventoryAccessInvitationModel) TableName() string {
	return "inventory_access_invitations"
}

func inventoryAccessInvitationModelFromPort(invitation ports.InventoryAccessInvitation) inventoryAccessInvitationModel {
	model := inventoryAccessInvitationModel{
		ID:                 invitation.ID,
		TenantID:           invitation.TenantID.String(),
		InventoryID:        invitation.InventoryID.String(),
		Email:              invitation.Email.String(),
		TokenHash:          invitation.TokenHash,
		Relationship:       string(invitation.Relationship),
		Status:             string(invitation.Status),
		InviterPrincipalID: invitation.InviterPrincipalID.String(),
		CreatedAt:          invitation.CreatedAt,
		ExpiresAt:          invitation.ExpiresAt,
	}
	if invitation.AcceptedPrincipalID.String() != "" {
		model.AcceptedPrincipalID = invitation.AcceptedPrincipalID.String()
	}
	if !invitation.AcceptedAt.IsZero() {
		acceptedAt := invitation.AcceptedAt
		model.AcceptedAt = &acceptedAt
	}
	if !invitation.RevokedAt.IsZero() {
		revokedAt := invitation.RevokedAt
		model.RevokedAt = &revokedAt
	}
	return model
}

func (m inventoryAccessInvitationModel) toPort() (ports.InventoryAccessInvitation, bool) {
	email, ok := identity.NewEmail(m.Email)
	if !ok {
		return ports.InventoryAccessInvitation{}, false
	}
	relationship := ports.InventoryAccessRelationship(m.Relationship)
	switch relationship {
	case ports.InventoryAccessViewer, ports.InventoryAccessEditor:
	default:
		return ports.InventoryAccessInvitation{}, false
	}
	status := ports.InventoryAccessInvitationStatus(m.Status)
	switch status {
	case ports.InventoryAccessInvitationPending, ports.InventoryAccessInvitationAccepted, ports.InventoryAccessInvitationRevoked, ports.InventoryAccessInvitationCancelled:
	default:
		return ports.InventoryAccessInvitation{}, false
	}
	invitation := ports.InventoryAccessInvitation{
		ID:                  m.ID,
		TenantID:            tenant.ID(m.TenantID),
		InventoryID:         inventory.InventoryID(m.InventoryID),
		Email:               email,
		TokenHash:           m.TokenHash,
		Relationship:        relationship,
		Status:              status,
		InviterPrincipalID:  identity.PrincipalID(m.InviterPrincipalID),
		AcceptedPrincipalID: identity.PrincipalID(m.AcceptedPrincipalID),
		CreatedAt:           m.CreatedAt,
		ExpiresAt:           m.ExpiresAt,
	}
	if m.AcceptedAt != nil {
		invitation.AcceptedAt = *m.AcceptedAt
	}
	if m.RevokedAt != nil {
		invitation.RevokedAt = *m.RevokedAt
	}
	return invitation, true
}
