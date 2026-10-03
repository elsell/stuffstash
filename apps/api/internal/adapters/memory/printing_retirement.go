package memory

import (
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *Store) retirePrinterWorkLocked(printer *printing.Printer, effects *ports.PrinterRetirement, printerAudit audit.Record) error {
	if effects == nil || effects.JobAudit == nil || effects.SettingsAudit == nil || effects.Now.IsZero() {
		return ports.ErrPrintConflict
	}
	changes := map[printing.JobID]printing.Job{}
	records := map[audit.ID]audit.Record{printerAudit.ID: printerAudit}
	addRecord := func(record audit.Record) error {
		if _, exists := s.auditRecords[record.ID]; exists {
			return ports.ErrPrintConflict
		}
		if _, exists := records[record.ID]; exists {
			return ports.ErrPrintConflict
		}
		records[record.ID] = record
		return nil
	}
	for id, before := range s.printingJobs {
		if before.Scope != printer.Scope || before.PrinterID != printer.ID || (before.Status != printing.JobQueued && before.Status != printing.JobClaimed) {
			continue
		}
		job := clonePrintJob(before)
		if err := job.Cancel(effects.Now, job.Revision); err != nil {
			return err
		}
		record, err := effects.JobAudit(before, job)
		if err != nil {
			return err
		}
		if err = addRecord(record); err != nil {
			return err
		}
		changes[id] = job
	}
	settings := s.printSettings[printer.Scope]
	changedSettings := settings.ClearRetiredPrinter(printer.ID, effects.Now)
	if changedSettings {
		record, err := effects.SettingsAudit(settings)
		if err != nil {
			return err
		}
		if err = addRecord(record); err != nil {
			return err
		}
	}
	// No writes occur until every callback and history key is accepted.
	for id, job := range changes {
		s.printingJobs[id] = job
		if printer.ActiveJobID == string(id) {
			printer.ActiveJobID = ""
			printer.ReservationState = ""
		}
	}
	if changedSettings {
		s.printSettings[printer.Scope] = settings
	}
	for id, record := range records {
		s.auditRecords[id] = record
	}
	return nil
}
