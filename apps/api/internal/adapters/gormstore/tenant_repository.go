package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
)

func (s Store) SaveTenant(ctx context.Context, item tenant.Tenant) error {
	model := tenantModel{
		ID:             item.ID.String(),
		Name:           item.Name.String(),
		LifecycleState: lifecycleStateOrActive(item.LifecycleState.String()),
	}

	return s.db.WithContext(ctx).Save(&model).Error
}

func (s Store) TenantByID(ctx context.Context, tenantID tenant.ID) (tenant.Tenant, bool, error) {
	var model tenantModel
	err := s.db.WithContext(ctx).Where(&tenantModel{ID: tenantID.String()}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return tenant.Tenant{}, false, nil
	}
	if err != nil {
		return tenant.Tenant{}, false, err
	}
	item, ok := model.toDomain()
	return item, ok, nil
}

func (s Store) TenantExists(ctx context.Context, tenantID tenant.ID) (bool, error) {
	var model tenantModel
	err := s.db.WithContext(ctx).Where(&tenantModel{ID: tenantID.String(), LifecycleState: tenant.LifecycleStateActive.String()}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s Store) UpdateTenant(ctx context.Context, item tenant.Tenant, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing tenantModel
		err := tx.Where(&tenantModel{ID: item.ID.String()}).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		if existing.LifecycleState != tenant.LifecycleStateActive.String() || item.LifecycleState != tenant.LifecycleStateActive {
			return ports.ErrForbidden
		}
		if err := tx.Model(&existing).Update("name", item.Name.String()).Error; err != nil {
			return err
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) UpdateTenantLifecycle(ctx context.Context, item tenant.Tenant, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing tenantModel
		err := tx.Where(&tenantModel{ID: item.ID.String()}).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		if existing.Name != item.Name.String() || existing.LifecycleState == item.LifecycleState.String() {
			return ports.ErrForbidden
		}
		if err := tx.Model(&existing).Update("lifecycle_state", item.LifecycleState.String()).Error; err != nil {
			return err
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) DeleteTenant(ctx context.Context, tenantID tenant.ID, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var childInventories int64
		if err := tx.Model(&inventoryModel{}).Where(&inventoryModel{TenantID: tenantID.String()}).Count(&childInventories).Error; err != nil {
			return err
		}
		if childInventories > 0 {
			return ports.ErrForbidden
		}
		if err := createAuditRecord(tx, auditRecord); err != nil {
			return err
		}
		result := tx.Where(&tenantModel{ID: tenantID.String()}).Delete(&tenantModel{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ports.ErrForbidden
		}
		return nil
	})
}

func (s Store) SaveTenantAndEnqueueOwnerGrant(ctx context.Context, eventID string, item tenant.Tenant, principal identity.Principal, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&tenantModel{
			ID:             item.ID.String(),
			Name:           item.Name.String(),
			LifecycleState: lifecycleStateOrActive(item.LifecycleState.String()),
		}).Error; err != nil {
			return err
		}

		if err := tx.Create(&authorizationOutboxEventModel{
			ID:          eventID,
			Kind:        string(ports.AuthorizationOutboxGrantTenantOwner),
			PrincipalID: principal.ID.String(),
			TenantID:    item.ID.String(),
		}).Error; err != nil {
			return err
		}

		return createAuditRecord(tx, auditRecord)
	})
}

func lifecycleStateOrActive(value string) string {
	if value == "" {
		return "active"
	}
	return value
}
