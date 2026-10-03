package gormstore

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) GetPrintSettings(ctx context.Context, scope printing.Scope) (printing.InventoryPrintSettings, error) {
	if scope.TenantID == "" || scope.InventoryID == "" {
		return printing.InventoryPrintSettings{}, ports.ErrPrintNotFound
	}
	var value printSettingsModel
	err := s.db.WithContext(ctx).Where(&printSettingsModel{TenantID: scope.TenantID, InventoryID: scope.InventoryID}).First(&value).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return printing.DefaultInventoryPrintSettings(scope), nil
	}
	return value.domain(), err
}
func (s Store) SavePrintSettings(ctx context.Context, next printing.InventoryPrintSettings, expected uint64, destination *printing.SettingsDestination, record audit.Record) (printing.InventoryPrintSettings, error) {
	if !next.ValidReplacement(expected) {
		return printing.InventoryPrintSettings{}, ports.ErrPrintConflict
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockPrintingInventory(tx, next.Scope); err != nil {
			return err
		}
		// Lock order is destination printer, then settings. This fences retirement or
		// media changes without moving settings state into the printer aggregate.
		if next.DefaultPrinterID != "" {
			if destination == nil || destination.ID != next.DefaultPrinterID {
				return ports.ErrPrintConflict
			}
			printer, err := printerByScope(tx, next.Scope, next.DefaultPrinterID)
			if err != nil {
				return err
			}
			if printer.Retired || printer.Revision != destination.Revision || printer.MediaFingerprint != destination.MediaFingerprint {
				return ports.ErrPrintConflict
			}
		}
		var current printSettingsModel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&printSettingsModel{TenantID: next.Scope.TenantID, InventoryID: next.Scope.InventoryID}).First(&current).Error
		missing := errors.Is(err, gorm.ErrRecordNotFound)
		if err != nil && !missing {
			return err
		}
		if current.Revision != expected {
			return ports.ErrPrintConflict
		}
		model := settingsModel(next)
		if missing {
			result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return ports.ErrPrintConflict
			}
		} else {
			result := tx.Model(&printSettingsModel{}).Where(&printSettingsModel{TenantID: next.Scope.TenantID, InventoryID: next.Scope.InventoryID, Revision: expected}).Select("*").Updates(&model)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return ports.ErrPrintConflict
			}
		}
		return createAuditRecord(tx, record)
	})
	if err != nil {
		return printing.InventoryPrintSettings{}, err
	}
	return next, nil
}

var _ ports.PrintSettingsRepository = Store{}
