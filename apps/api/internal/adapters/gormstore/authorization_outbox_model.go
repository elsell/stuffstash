package gormstore

import (
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

type authorizationOutboxEventModel struct {
	ID               string         `gorm:"primaryKey;size:26"`
	Kind             string         `gorm:"not null;size:80;index;check:chk_authorization_outbox_events_kind,kind IN ('grant_tenant_owner','grant_inventory_owner','grant_inventory_viewer','grant_inventory_editor');check:chk_authorization_outbox_events_inventory_required,(kind IN ('grant_inventory_owner','grant_inventory_viewer','grant_inventory_editor') AND inventory_id IS NOT NULL) OR (kind = 'grant_tenant_owner' AND inventory_id IS NULL)"`
	PrincipalID      string         `gorm:"not null;size:128;index"`
	TenantID         string         `gorm:"not null;size:26;index"`
	Tenant           tenantModel    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID      *string        `gorm:"size:26;index"`
	Inventory        inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:InventoryID;references:ID"`
	Attempts         int            `gorm:"not null;default:0"`
	LastError        string         `gorm:"not null;default:''"`
	ClaimID          string         `gorm:"not null;default:'';size:26;index"`
	ClaimedUntil     *time.Time     `gorm:"index"`
	ProcessedAt      *time.Time
	DeadLetteredAt   *time.Time `gorm:"index"`
	DeadLetterReason string     `gorm:"not null;default:''"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (authorizationOutboxEventModel) TableName() string {
	return "authorization_outbox_events"
}

func (m authorizationOutboxEventModel) toPort() ports.AuthorizationOutboxEvent {
	inventoryID := inventory.InventoryID("")
	if m.InventoryID != nil {
		inventoryID = inventory.InventoryID(*m.InventoryID)
	}
	event := ports.AuthorizationOutboxEvent{
		ID:          m.ID,
		Kind:        ports.AuthorizationOutboxEventKind(m.Kind),
		PrincipalID: identity.PrincipalID(m.PrincipalID),
		TenantID:    tenant.ID(m.TenantID),
		InventoryID: inventoryID,
		Attempts:    m.Attempts,
		LastError:   m.LastError,
		ClaimID:     m.ClaimID,
		CreatedAt:   m.CreatedAt,
	}
	if m.ClaimedUntil != nil {
		event.ClaimedUntil = *m.ClaimedUntil
	}
	if m.DeadLetteredAt != nil {
		event.DeadLetteredAt = *m.DeadLetteredAt
	}
	event.DeadLetterReason = m.DeadLetterReason
	return event
}
