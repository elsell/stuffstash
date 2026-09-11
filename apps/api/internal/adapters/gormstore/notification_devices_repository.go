package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"sort"
)

func (s Store) deviceDB(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx).Session(&gorm.Session{Logger: logger.Discard})
}
func (s Store) NotificationDeviceByID(ctx context.Context, scope ports.NotificationScope, id string) (ports.NotificationDevice, bool, error) {
	if !scope.Valid() || id == "" {
		return ports.NotificationDevice{}, false, ports.ErrForbidden
	}
	return readNotificationDevice(s.deviceDB(ctx).Where(notificationDeviceScope(scope)).Where(&notificationDeviceModel{ID: id}))
}
func (s Store) NotificationDeviceByInstallation(ctx context.Context, scope ports.NotificationScope, installation string) (ports.NotificationDevice, bool, error) {
	if !scope.Valid() || installation == "" {
		return ports.NotificationDevice{}, false, ports.ErrForbidden
	}
	return readNotificationDevice(s.deviceDB(ctx).Where(notificationDeviceScope(scope)).Where(&notificationDeviceModel{InstallationID: installation}))
}
func readNotificationDevice(query *gorm.DB) (ports.NotificationDevice, bool, error) {
	var model notificationDeviceModel
	err := query.First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ports.NotificationDevice{}, false, nil
	}
	if err != nil {
		return ports.NotificationDevice{}, false, err
	}
	value, err := model.toDevice()
	return value, err == nil, err
}
func (s Store) ListNotificationDevices(ctx context.Context, scope ports.NotificationScope, after string, limit int) ([]ports.NotificationDevice, error) {
	if !scope.Valid() {
		return nil, ports.ErrForbidden
	}
	if limit < 1 || limit > 1000 {
		return nil, ports.ErrInvalidProviderInput
	}
	query := s.deviceDB(ctx).Where(notificationDeviceScope(scope)).Where(&notificationDeviceModel{Active: true}).Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Limit(limit)
	if after != "" {
		query = query.Where(clause.Gt{Column: "id", Value: after})
	}
	var rows []notificationDeviceModel
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]ports.NotificationDevice, 0, len(rows))
	for _, row := range rows {
		value, err := row.toDevice()
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, nil
}
func (s Store) SaveNotificationDevice(ctx context.Context, value ports.NotificationDevice, expected int64, record audit.Record) error {
	if !value.Valid() || expected < 0 || value.Revision != expected+1 {
		return ports.ErrInvalidProviderInput
	}
	if record.TenantID.String() != value.Scope.TenantID.String() || record.InventoryID.String() != value.Scope.InventoryID.String() || record.PrincipalID.String() != value.Scope.PrincipalID.String() {
		return ports.ErrForbidden
	}
	return s.deviceDB(ctx).Transaction(func(tx *gorm.DB) error {
		var inventory inventoryModel
		if err := tx.Where(&inventoryModel{ID: value.Scope.InventoryID.String(), TenantID: value.Scope.TenantID.String()}).First(&inventory).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ports.ErrForbidden
			}
			return err
		}
		model := deviceModel(value)
		var previous notificationDeviceModel
		if expected > 0 {
			if err := tx.Where(notificationDeviceScope(value.Scope)).Where(&notificationDeviceModel{ID: value.ID, Revision: expected}).First(&previous).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ports.ErrConflict
				}
				return err
			}
			if previous.InstallationID != model.InstallationID || !previous.CreatedAt.Equal(model.CreatedAt) {
				return ports.ErrConflict
			}
		}
		if err := lockNotificationTokens(tx, previous.TokenKey, model.TokenKey); err != nil {
			return err
		}
		if value.Active {
			var count int64
			err := tx.Model(&notificationDeviceModel{}).Where(&notificationDeviceModel{TokenKey: model.TokenKey, Active: true}).Where(clause.Neq{Column: "principal_id", Value: model.PrincipalID}).Count(&count).Error
			if err != nil {
				return err
			}
			if count > 0 {
				return ports.ErrConflict
			}
		}
		var result *gorm.DB
		if expected == 0 {
			result = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model)
		} else {
			result = tx.Model(&notificationDeviceModel{}).Where(notificationDeviceScope(value.Scope)).Where(&notificationDeviceModel{ID: value.ID, Revision: expected}).Updates(map[string]any{"transport": model.Transport, "token": model.Token, "token_key": model.TokenKey, "active": model.Active, "revision": model.Revision, "updated_at": model.UpdatedAt})
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

// Lock digests are an operational ownership guard, never a discovery API.
func lockNotificationTokens(tx *gorm.DB, oldKey, newKey string) error {
	keys := []string{newKey}
	if oldKey != "" && oldKey != newKey {
		keys = append(keys, oldKey)
	}
	sort.Strings(keys)
	for _, key := range keys {
		row := notificationTokenLockModel{Key: key}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&notificationTokenLockModel{Key: key}).First(&row).Error; err != nil {
			return err
		}
	}
	return nil
}
