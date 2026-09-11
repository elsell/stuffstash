package gormstore

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func notificationPreferencesScope(scope ports.NotificationScope) notificationPreferencesModel {
	return notificationPreferencesModel{TenantID: scope.TenantID.String(), InventoryID: scope.InventoryID.String(), PrincipalID: scope.PrincipalID.String()}
}
func (s Store) NotificationPreferences(ctx context.Context, scope ports.NotificationScope) (ports.NotificationPreferencesRecord, bool, error) {
	if !scope.Valid() {
		return ports.NotificationPreferencesRecord{}, false, ports.ErrForbidden
	}
	var model notificationPreferencesModel
	err := s.db.WithContext(ctx).Where(notificationPreferencesScope(scope)).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ports.NotificationPreferencesRecord{}, false, nil
	}
	if err != nil {
		return ports.NotificationPreferencesRecord{}, false, err
	}
	value, err := model.toRecord()
	return value, err == nil, err
}
func (s Store) SaveNotificationPreferences(ctx context.Context, value ports.NotificationPreferencesRecord, expected int64, record audit.Record) error {
	if !value.Scope.Valid() || value.ID == "" || value.Settings.Validate() != nil || value.Revision != expected+1 {
		return ports.ErrInvalidProviderInput
	}
	if record.TenantID.String() != value.Scope.TenantID.String() || record.InventoryID.String() != value.Scope.InventoryID.String() || record.PrincipalID.String() != value.Scope.PrincipalID.String() {
		return ports.ErrForbidden
	}
	encoded, err := json.Marshal(value.Settings)
	if err != nil {
		return err
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var inventory inventoryModel
		err := tx.Where(&inventoryModel{ID: value.Scope.InventoryID.String(), TenantID: value.Scope.TenantID.String()}).First(&inventory).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		model := notificationPreferencesScope(value.Scope)
		model.ID = value.ID
		model.Revision = value.Revision
		model.Settings = string(encoded)
		model.CreatedAt = value.CreatedAt
		model.UpdatedAt = value.UpdatedAt
		var result *gorm.DB
		if expected == 0 {
			result = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model)
		} else {
			result = tx.Model(&notificationPreferencesModel{}).Where(notificationPreferencesScope(value.Scope)).Where(&notificationPreferencesModel{ID: value.ID, Revision: expected}).Updates(map[string]any{"revision": value.Revision, "settings": string(encoded), "updated_at": value.UpdatedAt})
		}
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ports.ErrConflict
		}
		return createAuditRecord(tx, record)
	})
}
func (s Store) ListNotificationRecipients(ctx context.Context, afterID string, limit int) ([]ports.NotificationPreferencesRecord, error) {
	if limit < 1 || limit > 1000 {
		return nil, ports.ErrInvalidProviderInput
	}
	query := s.db.WithContext(ctx).Model(&notificationPreferencesModel{}).Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Limit(limit)
	if afterID != "" {
		query = query.Where(clause.Gt{Column: "id", Value: afterID})
	}
	var rows []notificationPreferencesModel
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]ports.NotificationPreferencesRecord, 0, len(rows))
	for _, row := range rows {
		value, err := row.toRecord()
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, nil
}
