package gormstore

import (
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Parent scope is locked first. Connector -> printer -> job order matches
// consumer commands, and generations remain pending for durable reconciliation.
func deleteInventoryPrinting(tx *gorm.DB, scope printing.Scope, deletion audit.Record, effects *ports.InventoryDeletionEffects) error {
	var connectors []printingConnectorModel
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&printingConnectorModel{TenantID: scope.TenantID, InventoryID: scope.InventoryID}).Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Find(&connectors).Error; err != nil {
		return err
	}
	var printers []printingPrinterModel
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&printingPrinterModel{TenantID: scope.TenantID, InventoryID: scope.InventoryID}).Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Find(&printers).Error; err != nil {
		return err
	}
	if len(connectors)+len(printers) == 0 {
		return nil
	}
	if effects == nil || effects.Audit == nil || deletion.OccurredAt.IsZero() {
		return ports.ErrConflict
	}
	for _, model := range connectors {
		connector := model.domain()
		if !connector.RevokeDeletedScope(deletion.OccurredAt) {
			continue
		}
		var bindings []printingBindingModel
		if err := tx.Where(&printingBindingModel{TenantID: scope.TenantID, InventoryID: scope.InventoryID, ConnectorID: model.ID}).Find(&bindings).Error; err != nil {
			return err
		}
		for _, binding := range bindings {
			binding.Revoked = true
			binding.Generation = connector.Generation
			if err := tx.Save(&binding).Error; err != nil {
				return err
			}
		}
		next := printingConnectorFromDomain(connector)
		if err := tx.Save(&next).Error; err != nil {
			return err
		}
		record, err := effects.Audit(audit.ActionPrintConnectorUpdated, audit.TargetInventory, scope.InventoryID)
		if err != nil {
			return err
		}
		record.Metadata = map[string]string{"connector_id": model.ID, "reason": "inventory_deleted"}
		if err = createAuditRecord(tx, record); err != nil {
			return err
		}
	}
	for _, model := range printers {
		jobAudit := func(_, job printing.Job) (audit.Record, error) {
			action := audit.ActionPrintJobCanceled
			if job.Status == printing.JobUncertain {
				action = audit.ActionPrintJobUncertain
			}
			return effects.Audit(action, audit.TargetPrintJob, string(job.ID))
		}
		if err := retirePrinterWork(tx, &model, &ports.PrinterRetirement{Now: deletion.OccurredAt, JobAudit: jobAudit, SettingsAudit: func(printing.InventoryPrintSettings) (audit.Record, error) {
			return effects.Audit(audit.ActionPrintSettingsUpdated, audit.TargetInventory, scope.InventoryID)
		}}); err != nil {
			return err
		}
		var started []printingJobModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&printingJobModel{TenantID: scope.TenantID, InventoryID: scope.InventoryID, PrinterID: model.ID, Status: string(printing.JobPrinting)}).Find(&started).Error; err != nil {
			return err
		}
		for _, before := range started {
			job, err := before.domain()
			if err != nil {
				return err
			}
			job.DeleteScope(deletion.OccurredAt)
			if err = savePrintJobChange(tx, before, job, &model, jobAudit); err != nil {
				return err
			}
		}
		if !model.Retired {
			model.Retired = true
			model.Revision++
			model.UpdatedAt = deletion.OccurredAt
			if err := tx.Save(&model).Error; err != nil {
				return err
			}
			record, err := effects.Audit(audit.ActionPrinterUpdated, audit.TargetInventory, scope.InventoryID)
			if err != nil {
				return err
			}
			record.Metadata = map[string]string{"printer_id": model.ID, "reason": "inventory_deleted"}
			if err = createAuditRecord(tx, record); err != nil {
				return err
			}
		}
	}
	return nil
}

// Removing historical parent FKs requires new registrations to serialize with
// scope deletion. Existing printer mutations are fenced by their printer lock.
func lockPrintingInventory(tx *gorm.DB, scope printing.Scope) error {
	var model inventoryModel
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&inventoryModel{ID: scope.InventoryID, TenantID: scope.TenantID}).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrPrintNotFound
		}
		return err
	}
	if model.LifecycleState != "active" {
		return ports.ErrPrintDenied
	}
	return nil
}
