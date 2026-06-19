package gormstore

import (
	"context"
	"errors"
	"fmt"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) SaveInventory(ctx context.Context, item inventory.Inventory) error {
	model := inventoryModel{
		ID:       item.ID.String(),
		TenantID: item.TenantID.String(),
		Name:     item.Name.String(),
	}

	return s.db.WithContext(ctx).Save(&model).Error
}

func (s Store) InventoryByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (inventory.Inventory, bool, error) {
	var model inventoryModel
	err := s.db.WithContext(ctx).Where(&inventoryModel{
		ID:       inventoryID.String(),
		TenantID: tenantID.String(),
	}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return inventory.Inventory{}, false, nil
	}
	if err != nil {
		return inventory.Inventory{}, false, err
	}
	item, ok := model.toDomain()
	return item, ok, nil
}

func (s Store) SaveInventoryAndEnqueueOwnerGrant(ctx context.Context, eventID string, item inventory.Inventory, tenantID tenant.ID, principal identity.Principal, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&inventoryModel{
			ID:       item.ID.String(),
			TenantID: item.TenantID.String(),
			Name:     item.Name.String(),
		}).Error; err != nil {
			return err
		}

		inventoryID := item.ID.String()
		if err := tx.Create(&authorizationOutboxEventModel{
			ID:          eventID,
			Kind:        string(ports.AuthorizationOutboxGrantInventoryOwner),
			PrincipalID: principal.ID.String(),
			TenantID:    tenantID.String(),
			InventoryID: &inventoryID,
		}).Error; err != nil {
			return err
		}

		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) ListInventoriesByTenant(ctx context.Context, tenantID inventory.TenantID, page ports.InventoryListPageRequest) ([]inventory.Inventory, error) {
	var models []inventoryModel
	query := s.db.WithContext(ctx).Where(&inventoryModel{TenantID: tenantID.String()})
	if page.AfterInventoryID.String() != "" {
		query = query.Where(clause.Gt{Column: clause.Column{Name: "id"}, Value: page.AfterInventoryID.String()})
	}
	if page.Limit > 0 {
		query = query.Limit(page.Limit)
	}
	if err := query.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]inventory.Inventory, 0, len(models))
	for _, model := range models {
		item, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid inventory row %q", model.ID)
		}
		items = append(items, item)
	}

	return items, nil
}
