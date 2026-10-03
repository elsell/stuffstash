package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func savePrintJobChange(tx *gorm.DB, before printingJobModel, j printing.Job, p *printingPrinterModel, auditFor ports.PrintJobAudit) error {
	previous, err := before.domain()
	if err != nil {
		return err
	}
	if j.ID != previous.ID || j.Scope != previous.Scope || j.PrinterID != previous.PrinterID {
		return ports.ErrConflict
	}
	if j.Revision == previous.Revision {
		return nil
	}
	if previous.Status != printing.JobPrinting && j.Status == printing.JobPrinting && (p.Retired || p.MediaFingerprint != j.MediaFingerprint) {
		return ports.ErrConflict
	}
	if j.HoldsReservation() && p.ActiveJobID != "" && p.ActiveJobID != string(j.ID) {
		return ports.ErrConflict
	}
	if auditFor == nil {
		return ports.ErrConflict
	}
	record, err := auditFor(previous, j)
	if err != nil {
		return err
	}
	if err = validatePrintJobAudit(record, j); err != nil {
		return err
	}
	model, err := printJobModel(j, before.RequestFingerprint)
	if err != nil {
		return err
	}
	if err = tx.Save(&model).Error; err != nil {
		return err
	}
	if j.HoldsReservation() {
		p.ActiveJobID = string(j.ID)
		p.ReservationState = string(j.Status)
	} else if p.ActiveJobID == string(j.ID) {
		p.ActiveJobID = ""
		p.ReservationState = ""
	}
	if err = tx.Model(p).Select("ActiveJobID", "ReservationState").Updates(p).Error; err != nil {
		return err
	}
	return createAuditRecord(tx, record)
}
func validatePrintJobAudit(record audit.Record, j printing.Job) error {
	if record.ID == "" || string(record.TenantID) != j.Scope.TenantID || string(record.InventoryID) != j.Scope.InventoryID || record.TargetID != string(j.ID) {
		return ports.ErrConflict
	}
	return nil
}
func (s Store) UpdatePrintJob(ctx context.Context, input ports.PrintJobUpdate) (result printing.Job, err error) {
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if input.Authority != nil {
			if input.Authority.Scope != input.Scope || input.Authority.PrinterID != input.PrinterID {
				return ports.ErrPrintDenied
			}
			if e := printingConsumerFence(tx, *input.Authority, input.Now); e != nil {
				return e
			}
		}
		p, e := printerByScope(tx, input.Scope, input.PrinterID)
		if e != nil {
			return e
		}
		before, e := printJobByID(tx.Clauses(clause.Locking{Strength: "UPDATE"}), input.Scope, input.JobID)
		if e != nil {
			return e
		}
		if before.PrinterID != p.ID || input.Change == nil || input.Audit == nil {
			return ports.ErrConflict
		}
		j, e := before.domain()
		if e != nil {
			return e
		}
		printer, e := p.domain()
		if e != nil {
			return e
		}
		if e = input.Change(&j, printer); e != nil {
			return e
		}
		if e = savePrintJobChange(tx, before, j, &p, input.Audit); e != nil {
			return e
		}
		result = j
		return nil
	})
	if err != nil {
		return printing.Job{}, err
	}
	return
}
