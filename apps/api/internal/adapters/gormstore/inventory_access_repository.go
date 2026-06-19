package gormstore

import (
	"context"
	"errors"
	"fmt"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) SaveInventoryAccessGrantAndEnqueue(ctx context.Context, eventID string, grant ports.InventoryAccessGrant, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var containingInventory inventoryModel
		err := tx.Where(&inventoryModel{
			ID:       grant.InventoryID.String(),
			TenantID: grant.TenantID.String(),
		}).First(&containingInventory).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}

		grantKey := grant.CursorKey()
		var existingGrant inventoryAccessGrantModel
		err = tx.Where(&inventoryAccessGrantModel{
			TenantID:    grant.TenantID.String(),
			InventoryID: grant.InventoryID.String(),
			GrantKey:    grantKey,
		}).First(&existingGrant).Error
		if err == nil {
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err := tx.Save(&inventoryAccessGrantModel{
			TenantID:     grant.TenantID.String(),
			InventoryID:  grant.InventoryID.String(),
			GrantKey:     grantKey,
			PrincipalID:  grant.PrincipalID.String(),
			Relationship: string(grant.Relationship),
		}).Error; err != nil {
			return err
		}

		inventoryID := grant.InventoryID.String()
		if err := tx.Create(&authorizationOutboxEventModel{
			ID:          eventID,
			Kind:        string(outboxKindForInventoryAccess(grant.Relationship)),
			PrincipalID: grant.PrincipalID.String(),
			TenantID:    grant.TenantID.String(),
			InventoryID: &inventoryID,
		}).Error; err != nil {
			return err
		}

		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) ListInventoryAccessGrants(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.InventoryAccessGrantPageRequest) ([]ports.InventoryAccessGrant, error) {
	var models []inventoryAccessGrantModel
	query := s.db.WithContext(ctx).Where(&inventoryAccessGrantModel{
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
	})
	if page.AfterGrantKey != "" {
		query = query.Where(clause.Gt{Column: clause.Column{Name: "grant_key"}, Value: page.AfterGrantKey})
	}
	if page.Limit > 0 {
		query = query.Limit(page.Limit)
	}
	if err := query.Order(clause.OrderByColumn{Column: clause.Column{Name: "grant_key"}}).Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]ports.InventoryAccessGrant, 0, len(models))
	for _, model := range models {
		item, ok := model.toPort()
		if !ok {
			return nil, fmt.Errorf("invalid inventory access grant row %q", model.GrantKey)
		}
		items = append(items, item)
	}
	return items, nil
}

func outboxKindForInventoryAccess(relationship ports.InventoryAccessRelationship) ports.AuthorizationOutboxEventKind {
	switch relationship {
	case ports.InventoryAccessEditor:
		return ports.AuthorizationOutboxGrantInventoryEditor
	default:
		return ports.AuthorizationOutboxGrantInventoryViewer
	}
}
