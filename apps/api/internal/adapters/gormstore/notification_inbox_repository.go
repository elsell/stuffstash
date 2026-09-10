package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func notificationAuditScopeMatches(scope ports.NotificationScope, record audit.Record) bool {
	return scope.Valid() && record.TenantID.String() == scope.TenantID.String() && record.InventoryID.String() == scope.InventoryID.String() && record.PrincipalID.String() == scope.PrincipalID.String()
}
func (s Store) InsertNotification(ctx context.Context, value ports.NotificationRecord, record audit.Record) (ports.NotificationRecord, bool, error) {
	if !value.Valid() || value.ReadAt != nil {
		return ports.NotificationRecord{}, false, ports.ErrInvalidProviderInput
	}
	if !notificationAuditScopeMatches(value.Scope, record) {
		return ports.NotificationRecord{}, false, ports.ErrForbidden
	}
	var result ports.NotificationRecord
	created := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var inv inventoryModel
		err := tx.Where(&inventoryModel{ID: value.Scope.InventoryID.String(), TenantID: value.Scope.TenantID.String()}).First(&inv).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		model := notificationInboxScope(value.Scope)
		model.AssetID = value.Milestone.AssetID
		model.ExpirationDate = value.Milestone.Date.Value()
		model.ExpirationPrecision = string(value.Milestone.Date.Precision())
		model.Kind = string(value.Milestone.Kind)
		identity := model
		model.ID = value.ID
		model.CreatedAt = value.CreatedAt
		inserted := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model)
		if inserted.Error != nil {
			return inserted.Error
		}
		if inserted.RowsAffected == 0 {
			var existing notificationInboxModel
			if err := tx.Where(identity).First(&existing).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ports.ErrConflict
				}
				return err
			}
			result, err = existing.toRecord()
			return err
		}
		if err := createAuditRecord(tx, record); err != nil {
			return err
		}
		result = value.Clone()
		created = true
		return nil
	})
	if err != nil {
		return ports.NotificationRecord{}, false, err
	}
	return result, created, nil
}
func (s Store) NotificationByID(ctx context.Context, scope ports.NotificationScope, id string) (ports.NotificationRecord, bool, error) {
	if id == "" {
		return ports.NotificationRecord{}, false, ports.ErrInvalidProviderInput
	}
	if !scope.Valid() {
		return ports.NotificationRecord{}, false, ports.ErrForbidden
	}
	var model notificationInboxModel
	err := s.db.WithContext(ctx).Where(notificationInboxScope(scope)).Where(&notificationInboxModel{ID: id}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ports.NotificationRecord{}, false, nil
	}
	if err != nil {
		return ports.NotificationRecord{}, false, err
	}
	value, err := model.toRecord()
	return value, err == nil, err
}
func (s Store) ListNotifications(ctx context.Context, scope ports.NotificationScope, beforeID string, limit int) ([]ports.NotificationRecord, error) {
	if !scope.Valid() {
		return nil, ports.ErrForbidden
	}
	if limit < 1 || limit > 1000 {
		return nil, ports.ErrInvalidProviderInput
	}
	query := s.db.WithContext(ctx).Where(notificationInboxScope(scope)).Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}, Desc: true}).Limit(limit)
	if beforeID != "" {
		query = query.Where(clause.Lt{Column: "id", Value: beforeID})
	}
	var rows []notificationInboxModel
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	values := make([]ports.NotificationRecord, 0, len(rows))
	for _, row := range rows {
		value, err := row.toRecord()
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}
func (s Store) MarkNotificationRead(ctx context.Context, scope ports.NotificationScope, id string, at time.Time, record audit.Record) (bool, error) {
	if id == "" {
		return false, ports.ErrInvalidProviderInput
	}
	if !notificationAuditScopeMatches(scope, record) {
		return false, ports.ErrForbidden
	}
	if at.IsZero() {
		return false, ports.ErrInvalidProviderInput
	}
	changed := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Model(&notificationInboxModel{}).Where(notificationInboxScope(scope)).Where(&notificationInboxModel{ID: id})
		updated := query.Where(clause.Eq{Column: "read_at", Value: nil}).Update("read_at", at.UTC())
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected == 0 {
			var existing notificationInboxModel
			err := tx.Where(notificationInboxScope(scope)).Where(&notificationInboxModel{ID: id}).First(&existing).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ports.ErrForbidden
			}
			return err
		}
		if err := createAuditRecord(tx, record); err != nil {
			return err
		}
		changed = true
		return nil
	})
	return changed, err
}
