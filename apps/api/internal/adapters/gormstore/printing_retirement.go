package gormstore

import (
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// The caller holds the printer lock, shared by enqueue, claim, start and settings.
func retirePrinterWork(tx *gorm.DB, printer *printingPrinterModel, effects *ports.PrinterRetirement) error {
	if effects == nil || effects.JobAudit == nil || effects.SettingsAudit == nil || effects.Now.IsZero() {
		return ports.ErrPrintConflict
	}
	var jobs []printingJobModel
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&printingJobModel{TenantID: printer.TenantID, InventoryID: printer.InventoryID, PrinterID: printer.ID}).Where(clause.IN{Column: "status", Values: []any{string(printing.JobQueued), string(printing.JobClaimed)}}).Find(&jobs).Error; err != nil {
		return err
	}
	for _, model := range jobs {
		job, err := model.domain()
		if err != nil {
			return err
		}
		if err = job.Cancel(effects.Now, job.Revision); err != nil {
			return err
		}
		if err = savePrintJobChange(tx, model, job, printer, effects.JobAudit); err != nil {
			return err
		}
	}
	var settings printSettingsModel
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&printSettingsModel{TenantID: printer.TenantID, InventoryID: printer.InventoryID}).First(&settings).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	next := settings.domain()
	if !next.ClearRetiredPrinter(printing.PrinterID(printer.ID), effects.Now) {
		return nil
	}
	record, err := effects.SettingsAudit(next)
	if err != nil {
		return err
	}
	model := settingsModel(next)
	if err = tx.Save(&model).Error; err != nil {
		return err
	}
	return createAuditRecord(tx, record)
}
