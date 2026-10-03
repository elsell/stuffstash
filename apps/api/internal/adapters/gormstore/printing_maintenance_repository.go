package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Maintenance is an operational stream. Each candidate is rechecked under the
// same printer lock as claim/start; content and reservation changes are atomic.
func (s Store) MaintainPrintJobs(ctx context.Context, in ports.PrintJobMaintenance) (ports.PrintMaintenancePage, error) {
	page := ports.PrintMaintenancePage{}
	if in.Limit < 1 || in.Limit > 1000 || in.Audit == nil {
		return page, ports.ErrConflict
	}
	var candidates []printingJobModel
	err := s.db.WithContext(ctx).Select("id", "tenant_id", "inventory_id", "printer_id").Where(clause.Gt{Column: "id", Value: in.After}).Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Limit(in.Limit + 1).Find(&candidates).Error
	if err != nil {
		return page, err
	}
	page.HasMore = len(candidates) > in.Limit
	if page.HasMore {
		candidates = candidates[:in.Limit]
	}
	for _, candidate := range candidates {
		err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			scope := printing.Scope{TenantID: candidate.TenantID, InventoryID: candidate.InventoryID}
			printer, err := printerByScope(tx, scope, printing.PrinterID(candidate.PrinterID))
			if err != nil {
				return err
			}
			row, err := printJobByID(tx, scope, printing.JobID(candidate.ID))
			if err != nil {
				return err
			}
			j, err := row.domain()
			if err != nil {
				return err
			}
			before := j
			j.Maintain(in.Now)
			if j.Revision != before.Revision {
				if err = savePrintJobChange(tx, row, j, &printer, in.Audit); err != nil {
					return err
				}

			}
			if j.Terminal() && !j.UpdatedAt.After(in.TerminalBefore) {
				if err = tx.Where(&printingAttemptIndex{JobID: row.ID}).Delete(&printingAttemptIndex{}).Error; err != nil {
					return err
				}
				return tx.Delete(&row).Error
			}
			if j.Terminal() && !j.Artifact.ExpiresAt.After(in.Now) && len(row.ArtifactContent) > 0 {
				return tx.Model(&row).Update("artifact_content", []byte{}).Error
			}
			return nil
		})
		if err != nil {
			return page, err
		}
		page.After = candidate.ID
	}
	return page, nil
}
