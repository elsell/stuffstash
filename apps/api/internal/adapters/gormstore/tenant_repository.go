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

func (s Store) SaveTenantAndEnqueueOwnerGrant(ctx context.Context, eventID string, item tenant.Tenant, principal identity.Principal, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&tenantModel{
			ID:   item.ID.String(),
			Name: item.Name.String(),
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
