package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Attempt lookup has its own scoped index, including settled attempts, so a
// connector can recover a lost API response without scanning other jobs.
type printingAttemptIndex struct {
	ID                                        string `gorm:"primaryKey"`
	TenantID, InventoryID, ConnectorID, JobID string
}

func (printingAttemptIndex) TableName() string { return "print_attempt_index" }
func (s Store) FindPrintAttempt(ctx context.Context, scope printing.Scope, connector printing.ConnectorID, id printing.AttemptID) (printing.Job, error) {
	if scope.TenantID == "" || scope.InventoryID == "" || connector == "" || id == "" {
		return printing.Job{}, ports.ErrPrintJobNotFound
	}
	var index printingAttemptIndex
	err := s.db.WithContext(ctx).Where(&printingAttemptIndex{ID: string(id), TenantID: scope.TenantID, InventoryID: scope.InventoryID, ConnectorID: string(connector)}).First(&index).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return printing.Job{}, ports.ErrPrintJobNotFound
	}
	if err != nil {
		return printing.Job{}, err
	}
	return s.GetPrintJob(ctx, scope, printing.JobID(index.JobID))
}
func (s Store) ClaimPrintJob(ctx context.Context, input ports.PrintClaim) (result printing.Job, found bool, err error) {
	if !input.Owner.Valid() || input.Owner.ConnectorID != input.Authority.ConnectorID || input.Lease <= 0 || input.ReportMaxAge <= 0 || input.Audit == nil {
		return result, false, ports.ErrConflict
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if e := printingConsumerFence(tx, input.Authority, input.Now); e != nil {
			return e
		}
		p, e := printerByScope(tx, input.Authority.Scope, input.Authority.PrinterID)
		if e != nil {
			return e
		}
		if p.Retired {
			return ports.ErrConflict
		}
		if p.ActiveJobID != "" {
			before, e := printJobByID(tx.Clauses(clause.Locking{Strength: "UPDATE"}), input.Authority.Scope, printing.JobID(p.ActiveJobID))
			if e != nil {
				return e
			}
			active, e := before.domain()
			if e != nil {
				return e
			}
			if active.Expire(input.Now) {
				if e = savePrintJobChange(tx, before, active, &p, input.Audit); e != nil {
					return e
				}
			}
			if p.ActiveJobID != "" {
				return nil
			}
		}
		ready, e := printingDispatchReady(tx, input.Authority, input.Now, input.ReportMaxAge)
		if e != nil {
			return e
		}
		if !ready {
			return nil
		}
		var before printingJobModel
		e = scopedPrintJobs(tx, input.Authority.Scope).Clauses(clause.Locking{Strength: "UPDATE"}).Where(&printingJobModel{PrinterID: p.ID, Status: string(printing.JobQueued), MediaFingerprint: p.MediaFingerprint}).Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}}).Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).First(&before).Error
		if errors.Is(e, gorm.ErrRecordNotFound) {
			return nil
		}
		if e != nil {
			return e
		}
		j, e := before.domain()
		if e != nil {
			return e
		}
		if e = j.Claim(input.Owner.AttemptID, input.Owner, input.Now, input.Lease, j.Revision); e != nil {
			return e
		}
		index := printingAttemptIndex{ID: string(input.Owner.AttemptID), TenantID: p.TenantID, InventoryID: p.InventoryID, ConnectorID: string(input.Authority.ConnectorID), JobID: before.ID}
		if e = tx.Create(&index).Error; e != nil {
			return e
		}
		if e = savePrintJobChange(tx, before, j, &p, input.Audit); e != nil {
			return e
		}
		result = j
		found = true
		return nil
	})
	if err != nil {
		return printing.Job{}, false, err
	}
	return
}

var _ ports.PrintJobRepository = Store{}
