package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func scopedPrintJobs(db *gorm.DB, scope printing.Scope) *gorm.DB {
	return db.Where(&printingJobModel{TenantID: scope.TenantID, InventoryID: scope.InventoryID})
}
func printJobByID(db *gorm.DB, scope printing.Scope, id printing.JobID) (printingJobModel, error) {
	if scope.TenantID == "" || scope.InventoryID == "" || id == "" {
		return printingJobModel{}, ports.ErrPrintJobNotFound
	}
	var m printingJobModel
	err := scopedPrintJobs(db, scope).Where(&printingJobModel{ID: string(id)}).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = ports.ErrPrintJobNotFound
	}
	return m, err
}
func (s Store) GetPrintJob(ctx context.Context, scope printing.Scope, id printing.JobID) (printing.Job, error) {
	m, err := printJobByID(s.db.WithContext(ctx), scope, id)
	if err != nil {
		return printing.Job{}, err
	}
	return m.domain()
}
func (s Store) ListPrintJobs(ctx context.Context, scope printing.Scope, printer printing.PrinterID, limit int, after string) ([]printing.Job, error) {
	if scope.TenantID == "" || scope.InventoryID == "" || limit < 1 || limit > 100 {
		return nil, ports.ErrConflict
	}
	q := scopedPrintJobs(s.db.WithContext(ctx), scope).Where("id > ?", after)
	if printer != "" {
		q = q.Where(&printingJobModel{PrinterID: string(printer)})
	}
	var rows []printingJobModel
	if err := q.Order("id ASC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]printing.Job, 0, len(rows))
	for _, m := range rows {
		j, err := m.domain()
		if err != nil {
			return nil, err
		}
		result = append(result, j)
	}
	return result, nil
}
func printRequest(db *gorm.DB, j printing.Job) (printingJobModel, error) {
	var m printingJobModel
	err := scopedPrintJobs(db, j.Scope).Where(&printingJobModel{RequestedBy: j.RequestedBy, IdempotencyKey: j.IdempotencyKey}).First(&m).Error
	return m, err
}
func (s Store) CreatePrintJob(ctx context.Context, input ports.PrintJobCreate) (result printing.Job, created bool, err error) {
	j := input.Job
	if j.Scope.TenantID == "" || j.Scope.InventoryID == "" || j.RequestedBy == "" || j.IdempotencyKey == "" || input.RequestFingerprint == "" {
		return result, false, ports.ErrConflict
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		replay := func(m printingJobModel) error {
			if m.RequestFingerprint != input.RequestFingerprint {
				return ports.ErrConflict
			}
			var e error
			result, e = m.domain()
			return e
		}
		if m, e := printRequest(tx, j); e == nil {
			return replay(m)
		} else if !errors.Is(e, gorm.ErrRecordNotFound) {
			return e
		}
		p, e := printerByScope(tx, j.Scope, j.PrinterID)
		if e != nil {
			return e
		}
		// A concurrent creator may have committed while we waited for this printer.
		if m, e := printRequest(tx, j); e == nil {
			return replay(m)
		} else if !errors.Is(e, gorm.ErrRecordNotFound) {
			return e
		}
		if p.Retired || p.Revision != input.PrinterRevision || p.MediaFingerprint != j.MediaFingerprint || j.ID == "" || j.Status != printing.JobQueued || j.Revision != 1 || len(j.Attempts) != 0 {
			return ports.ErrConflict
		}
		if input.Audit.ID == "" || string(input.Audit.TenantID) != j.Scope.TenantID || string(input.Audit.InventoryID) != j.Scope.InventoryID || input.Audit.TargetID != string(j.ID) {
			return ports.ErrConflict
		}
		model, e := printJobModel(j, input.RequestFingerprint)
		if e != nil {
			return e
		}
		insert := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model)
		if insert.Error != nil {
			return insert.Error
		}
		if insert.RowsAffected == 0 {
			m, e := printRequest(tx, j)
			if e != nil {
				return ports.ErrConflict
			}
			return replay(m)
		}
		if e = createAuditRecord(tx, input.Audit); e != nil {
			return e
		}
		result = j
		created = true
		return nil
	})
	if err != nil {
		return printing.Job{}, false, err
	}
	return
}
