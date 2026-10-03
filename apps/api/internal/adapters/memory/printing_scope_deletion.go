package memory

import (
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"maps"
)

func (s *Store) deleteInventoryPrintingLocked(scope printing.Scope, deletion audit.Record, effects *ports.InventoryDeletionEffects) error {
	stage := &Store{printingConnectors: maps.Clone(s.printingConnectors), printingBindings: maps.Clone(s.printingBindings), printingPrinters: maps.Clone(s.printingPrinters), printingJobs: maps.Clone(s.printingJobs), printSettings: maps.Clone(s.printSettings), auditRecords: maps.Clone(s.auditRecords)}
	add := func(record audit.Record) error {
		if _, exists := stage.auditRecords[record.ID]; exists {
			return ports.ErrConflict
		}
		stage.auditRecords[record.ID] = record
		return nil
	}
	for id, before := range stage.printingConnectors {
		if before.Scope != scope {
			continue
		}
		if effects == nil || effects.Audit == nil || deletion.OccurredAt.IsZero() {
			return ports.ErrConflict
		}
		connector := clonePrintConnector(before)
		if !connector.RevokeDeletedScope(deletion.OccurredAt) {
			continue
		}
		stage.printingConnectors[id] = connector
		for key, binding := range stage.printingBindings {
			if binding.Scope == scope && binding.ConnectorID == id {
				binding.Revoked = true
				binding.Generation = connector.Generation
				stage.printingBindings[key] = binding
			}
		}
		record, err := effects.Audit(audit.ActionPrintConnectorUpdated, audit.TargetInventory, scope.InventoryID)
		if err != nil {
			return err
		}
		record.Metadata = map[string]string{"connector_id": string(id), "reason": "inventory_deleted"}
		if err = add(record); err != nil {
			return err
		}
	}
	for id, before := range stage.printingPrinters {
		if before.Scope != scope {
			continue
		}
		if effects == nil || effects.Audit == nil || deletion.OccurredAt.IsZero() {
			return ports.ErrConflict
		}
		printer := clonePrintingPrinter(before)
		for jobID, old := range stage.printingJobs {
			if old.Scope != scope || old.PrinterID != id {
				continue
			}
			job := clonePrintJob(old)
			if !job.DeleteScope(deletion.OccurredAt) {
				continue
			}
			action := audit.ActionPrintJobCanceled
			if job.Status == printing.JobUncertain {
				action = audit.ActionPrintJobUncertain
			}
			record, err := effects.Audit(action, audit.TargetPrintJob, string(jobID))
			if err != nil {
				return err
			}
			if err = add(record); err != nil {
				return err
			}
			stage.printingJobs[jobID] = job
			if printer.ActiveJobID == string(jobID) {
				if job.HoldsReservation() {
					printer.ReservationState = string(job.Status)
				} else {
					printer.ActiveJobID = ""
					printer.ReservationState = ""
				}
			}
		}
		settings := stage.printSettings[scope]
		if settings.ClearRetiredPrinter(id, deletion.OccurredAt) {
			record, err := effects.Audit(audit.ActionPrintSettingsUpdated, audit.TargetInventory, scope.InventoryID)
			if err != nil {
				return err
			}
			if err = add(record); err != nil {
				return err
			}
			stage.printSettings[scope] = settings
		}
		if !printer.Retired {
			printer.Retired = true
			printer.Revision++
			printer.UpdatedAt = deletion.OccurredAt
			record, err := effects.Audit(audit.ActionPrinterUpdated, audit.TargetInventory, scope.InventoryID)
			if err != nil {
				return err
			}
			record.Metadata = map[string]string{"printer_id": string(id), "reason": "inventory_deleted"}
			if err = add(record); err != nil {
				return err
			}
		}
		stage.printingPrinters[id] = printer
	}
	if err := add(deletion); err != nil {
		return err
	}
	s.printingConnectors = stage.printingConnectors
	s.printingBindings = stage.printingBindings
	s.printingPrinters = stage.printingPrinters
	s.printingJobs = stage.printingJobs
	s.printSettings = stage.printSettings
	s.auditRecords = stage.auditRecords
	return nil
}

func (s *Store) printingInventoryExistsLocked(scope printing.Scope) bool {
	item, found := s.inventories[inventory.InventoryID(scope.InventoryID)]
	return found && item.TenantID.String() == scope.TenantID && item.IsActive()
}
