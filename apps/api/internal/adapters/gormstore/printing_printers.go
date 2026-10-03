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

func (s Store) CreatePrinter(ctx context.Context, p printing.Printer, record audit.Record) (printing.Printer, bool, error) {
	model, err := printingPrinterFromDomain(p)
	if err != nil {
		return printing.Printer{}, false, err
	}
	var result printing.Printer
	created := false
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		insert := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "tenant_id"}, {Name: "inventory_id"}, {Name: "request_key"}}, DoNothing: true}).Create(&model)
		if insert.Error != nil {
			return insert.Error
		}
		if insert.RowsAffected == 0 {
			var existing printingPrinterModel
			if err := tx.Where(&printingPrinterModel{TenantID: p.Scope.TenantID, InventoryID: p.Scope.InventoryID, RequestKey: p.RequestKey}).First(&existing).Error; err != nil {
				return err
			}
			if existing.RequestFingerprint != p.RequestFingerprint {
				return ports.ErrPrintConflict
			}
			var err error
			result, err = existing.domain()
			return err
		}
		if err := createAuditRecord(tx, record); err != nil {
			return err
		}
		result = p
		created = true
		return nil
	})
	return result, created, err
}
func (s Store) GetPrinter(ctx context.Context, scope printing.Scope, id printing.PrinterID) (printing.Printer, error) {
	if scope.TenantID == "" || scope.InventoryID == "" || id == "" {
		return printing.Printer{}, ports.ErrPrintNotFound
	}
	var model printingPrinterModel
	err := s.db.WithContext(ctx).Where(&printingPrinterModel{ID: string(id), TenantID: scope.TenantID, InventoryID: scope.InventoryID}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return printing.Printer{}, ports.ErrPrintNotFound
	}
	if err != nil {
		return printing.Printer{}, err
	}
	return model.domain()
}
func (s Store) ListPrinters(ctx context.Context, scope printing.Scope, limit int, after string) ([]printing.Printer, error) {
	if scope.TenantID == "" || scope.InventoryID == "" {
		return nil, ports.ErrPrintNotFound
	}
	query := s.db.WithContext(ctx).Where(&printingPrinterModel{TenantID: scope.TenantID, InventoryID: scope.InventoryID}).Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}})
	if after != "" {
		query = query.Where(clause.Gt{Column: "id", Value: after})
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	var models []printingPrinterModel
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]printing.Printer, 0, len(models))
	for _, model := range models {
		p, err := model.domain()
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}
func (s Store) UpdatePrinter(ctx context.Context, scope printing.Scope, id printing.PrinterID, revision uint64, change ports.PrinterMutation, makeAudit ports.PrinterAudit) (printing.Printer, error) {
	var result printing.Printer
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		model, err := printerByScope(tx, scope, id)
		if err != nil {
			return err
		}
		if model.Revision != revision {
			return ports.ErrPrintConflict
		}
		p, err := model.domain()
		if err != nil {
			return err
		}
		if err := change(&p); err != nil {
			return err
		}
		if string(p.ID) != model.ID || p.Scope != scope {
			return ports.ErrPrintConflict
		}
		next, err := printingPrinterFromDomain(p)
		if err != nil {
			return err
		}
		record, err := makeAudit(p)
		if err != nil {
			return err
		}
		if err := tx.Save(&next).Error; err != nil {
			return err
		}
		if err := createAuditRecord(tx, record); err != nil {
			return err
		}
		result = p
		return nil
	})
	return result, err
}

var _ ports.PrinterRepository = Store{}
