package gormstore

import (
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func printerByScope(tx *gorm.DB, scope printing.Scope, id printing.PrinterID) (printingPrinterModel, error) {
	if scope.TenantID == "" || scope.InventoryID == "" || id == "" {
		return printingPrinterModel{}, ports.ErrPrintNotFound
	}
	var model printingPrinterModel
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&printingPrinterModel{ID: string(id), TenantID: scope.TenantID, InventoryID: scope.InventoryID}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = ports.ErrPrintNotFound
	}
	return model, err
}

// Acquires the global registry lock order: connector, then binding. Call before
// printer/job locks and keep the transaction open through the protected action.
func printingConsumerFence(tx *gorm.DB, authority printing.ConsumerAuthority, now time.Time) error {
	if authority.Scope.TenantID == "" || authority.Scope.InventoryID == "" || authority.ConnectorID == "" || authority.PrinterID == "" {
		return ports.ErrPrintDenied
	}
	var connector printingConnectorModel
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&printingConnectorModel{ID: string(authority.ConnectorID), TenantID: authority.Scope.TenantID, InventoryID: authority.Scope.InventoryID}).First(&connector).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrPrintDenied
		}
		return err
	}
	var binding printingBindingModel
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&printingBindingModel{ConnectorID: string(authority.ConnectorID), PrinterID: string(authority.PrinterID), TenantID: authority.Scope.TenantID, InventoryID: authority.Scope.InventoryID}).First(&binding).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrPrintDenied
		}
		return err
	}
	if !printing.AcceptsAuthority(connector.domain(), binding.domain(), authority, now) {
		return ports.ErrPrintDenied
	}
	return nil
}
