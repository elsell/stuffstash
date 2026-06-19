package gormstore

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) Store {
	return Store{db: db}
}

func Migrate(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).AutoMigrate(&tenantModel{}, &inventoryModel{}, &authorizationOutboxEventModel{})
}

func (s Store) SaveTenant(ctx context.Context, item tenant.Tenant) error {
	model := tenantModel{
		ID:   item.ID.String(),
		Name: item.Name.String(),
	}

	return s.db.WithContext(ctx).Save(&model).Error
}

func (s Store) TenantExists(ctx context.Context, tenantID tenant.ID) (bool, error) {
	var model tenantModel
	err := s.db.WithContext(ctx).Where(&tenantModel{ID: tenantID.String()}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s Store) SaveInventory(ctx context.Context, item inventory.Inventory) error {
	model := inventoryModel{
		ID:       item.ID.String(),
		TenantID: item.TenantID.String(),
		Name:     item.Name.String(),
	}

	return s.db.WithContext(ctx).Save(&model).Error
}

func (s Store) SaveTenantAndEnqueueOwnerGrant(ctx context.Context, eventID string, item tenant.Tenant, principal identity.Principal) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&tenantModel{
			ID:   item.ID.String(),
			Name: item.Name.String(),
		}).Error; err != nil {
			return err
		}

		return tx.Create(&authorizationOutboxEventModel{
			ID:          eventID,
			Kind:        string(ports.AuthorizationOutboxGrantTenantOwner),
			PrincipalID: principal.ID.String(),
			TenantID:    item.ID.String(),
		}).Error
	})
}

func (s Store) SaveInventoryAndEnqueueOwnerGrant(ctx context.Context, eventID string, item inventory.Inventory, tenantID tenant.ID, principal identity.Principal) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&inventoryModel{
			ID:       item.ID.String(),
			TenantID: item.TenantID.String(),
			Name:     item.Name.String(),
		}).Error; err != nil {
			return err
		}

		inventoryID := item.ID.String()
		return tx.Create(&authorizationOutboxEventModel{
			ID:          eventID,
			Kind:        string(ports.AuthorizationOutboxGrantInventoryOwner),
			PrincipalID: principal.ID.String(),
			TenantID:    tenantID.String(),
			InventoryID: &inventoryID,
		}).Error
	})
}

func (s Store) ClaimPendingAuthorizationOutboxEvents(ctx context.Context, claimID string, limit int, leaseUntil time.Time) ([]ports.AuthorizationOutboxEvent, error) {
	if limit <= 0 {
		limit = 25
	}

	events := []ports.AuthorizationOutboxEvent{}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var models []authorizationOutboxEventModel
		now := time.Now()
		if err := tx.
			Clauses(skipLockedForUpdate()).
			Where(claimableAuthorizationOutboxEvent(now)).
			Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}}).
			Limit(limit).
			Find(&models).Error; err != nil {
			return err
		}

		claimed := make([]authorizationOutboxEventModel, 0, len(models))
		for _, model := range models {
			model.ClaimID = claimID
			model.ClaimedUntil = &leaseUntil
			claimed = append(claimed, model)
		}

		for _, model := range claimed {
			if err := tx.
				Model(&authorizationOutboxEventModel{}).
				Where(&authorizationOutboxEventModel{ID: model.ID}).
				Updates(map[string]any{
					"claim_id":      model.ClaimID,
					"claimed_until": model.ClaimedUntil,
				}).Error; err != nil {
				return err
			}
			events = append(events, model.toPort())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return events, nil
}

func skipLockedForUpdate() clause.Locking {
	return clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}
}

func claimableAuthorizationOutboxEvent(now time.Time) clause.Expression {
	return clause.And(
		clause.Eq{Column: clause.Column{Name: "processed_at"}, Value: nil},
		clause.Or(
			clause.Eq{Column: clause.Column{Name: "claim_id"}, Value: ""},
			clause.Lte{Column: clause.Column{Name: "claimed_until"}, Value: now},
		),
	)
}

func (s Store) MarkAuthorizationOutboxEventProcessed(ctx context.Context, eventID string, claimID string) error {
	now := time.Now()
	result := s.db.WithContext(ctx).
		Model(&authorizationOutboxEventModel{}).
		Where(&authorizationOutboxEventModel{ID: eventID, ClaimID: claimID}).
		Updates(map[string]any{
			"processed_at":  now,
			"last_error":    "",
			"claim_id":      "",
			"claimed_until": nil,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ports.ErrAuthorizationOutboxClaimLost
	}
	return nil
}

func (s Store) MarkAuthorizationOutboxEventFailed(ctx context.Context, eventID string, claimID string, reason string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var model authorizationOutboxEventModel
		if err := tx.Where(&authorizationOutboxEventModel{ID: eventID, ClaimID: claimID}).First(&model).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ports.ErrAuthorizationOutboxClaimLost
			}
			return err
		}

		model.Attempts++
		model.LastError = reason
		model.ClaimID = ""
		model.ClaimedUntil = nil
		return tx.Save(&model).Error
	})
}

func (s Store) ListInventoriesByTenant(ctx context.Context, tenantID inventory.TenantID) ([]inventory.Inventory, error) {
	var models []inventoryModel
	if err := s.db.WithContext(ctx).Where(&inventoryModel{TenantID: tenantID.String()}).Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]inventory.Inventory, 0, len(models))
	for _, model := range models {
		item, ok := model.toDomain()
		if !ok {
			continue
		}
		items = append(items, item)
	}

	sort.Slice(items, func(left int, right int) bool {
		return items[left].ID.String() < items[right].ID.String()
	})

	return items, nil
}

type tenantModel struct {
	ID        string `gorm:"primaryKey;size:26"`
	Name      string `gorm:"not null;size:120"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (tenantModel) TableName() string {
	return "tenants"
}

type inventoryModel struct {
	ID        string      `gorm:"primaryKey;size:26"`
	TenantID  string      `gorm:"not null;size:26;index"`
	Tenant    tenantModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	Name      string      `gorm:"not null;size:120"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type authorizationOutboxEventModel struct {
	ID           string         `gorm:"primaryKey;size:26"`
	Kind         string         `gorm:"not null;size:80;index;check:chk_authorization_outbox_events_kind,kind IN ('grant_tenant_owner','grant_inventory_owner');check:chk_authorization_outbox_events_inventory_required,(kind = 'grant_inventory_owner' AND inventory_id IS NOT NULL) OR (kind = 'grant_tenant_owner' AND inventory_id IS NULL)"`
	PrincipalID  string         `gorm:"not null;size:128;index"`
	TenantID     string         `gorm:"not null;size:26;index"`
	Tenant       tenantModel    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID  *string        `gorm:"size:26;index"`
	Inventory    inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:InventoryID;references:ID"`
	Attempts     int            `gorm:"not null;default:0"`
	LastError    string         `gorm:"not null;default:''"`
	ClaimID      string         `gorm:"not null;default:'';size:26;index"`
	ClaimedUntil *time.Time     `gorm:"index"`
	ProcessedAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (authorizationOutboxEventModel) TableName() string {
	return "authorization_outbox_events"
}

func (inventoryModel) TableName() string {
	return "inventories"
}

func (m inventoryModel) toDomain() (inventory.Inventory, bool) {
	id, ok := inventory.NewID(m.ID)
	if !ok {
		return inventory.Inventory{}, false
	}
	name, ok := inventory.NewName(m.Name)
	if !ok {
		return inventory.Inventory{}, false
	}

	return inventory.Inventory{
		ID:       id,
		TenantID: inventory.TenantID(m.TenantID),
		Name:     name,
	}, true
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
	return event
}
