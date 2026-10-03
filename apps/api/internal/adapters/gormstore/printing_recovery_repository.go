package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm/clause"
)

// Recovery enumerates only unresolved scoped reservations. Filtering the current
// attempt after decoding avoids exposing an older connector's settled index entry.
func (s Store) ListPrintConsumerAttempts(ctx context.Context, scope printing.Scope, connector printing.ConnectorID, printer printing.PrinterID, limit int, after string) ([]printing.Job, error) {
	if scope.TenantID == "" || scope.InventoryID == "" || connector == "" || limit < 1 || limit > 101 {
		return nil, ports.ErrConflict
	}
	result := []printing.Job{}
	const batchSize = 100
	for {
		q := scopedPrintJobs(s.db.WithContext(ctx), scope).Select("id", "snapshot").Where(clause.Gt{Column: "id", Value: after}).Where(clause.IN{Column: "status", Values: []any{string(printing.JobClaimed), string(printing.JobPrinting), string(printing.JobUncertain)}})
		if printer != "" {
			q = q.Where(&printingJobModel{PrinterID: string(printer)})
		}
		var rows []printingJobModel
		if err := q.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Limit(batchSize).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			j, err := row.domain()
			if err != nil {
				return nil, err
			}
			if len(j.Attempts) > 0 && j.Attempts[len(j.Attempts)-1].Authority.ConnectorID == connector {
				result = append(result, j)
				if len(result) == limit {
					return result, nil
				}
			}
			after = row.ID
		}
		if len(rows) < batchSize {
			return result, nil
		}
	}
}
