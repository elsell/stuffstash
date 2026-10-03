package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Lease renewal is the only mutation that bypasses domain history. Its fixed
// domain command cannot change output evidence, job ownership or reservation.
func (s Store) RenewPrintJob(ctx context.Context, in ports.PrintLeaseRenewal) (result printing.Job, err error) {
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if in.Owner.ConnectorID != in.Authority.ConnectorID {
			return ports.ErrPrintDenied
		}
		if e := printingConsumerFence(tx, in.Authority, in.Now); e != nil {
			return e
		}
		p, e := printerByScope(tx, in.Authority.Scope, in.Authority.PrinterID)
		if e != nil {
			return e
		}
		before, e := printJobByID(tx.Clauses(clause.Locking{Strength: "UPDATE"}), in.Authority.Scope, in.JobID)
		if e != nil {
			return e
		}
		if before.PrinterID != p.ID || p.ActiveJobID != before.ID {
			return ports.ErrConflict
		}
		j, e := before.domain()
		if e != nil {
			return e
		}
		if e = j.Renew(in.Owner, in.Now, in.Lease, in.Revision); e != nil {
			return e
		}
		model, e := printJobModel(j, before.RequestFingerprint, before.ArtifactContent)
		if e != nil {
			return e
		}
		if e = tx.Save(&model).Error; e != nil {
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
