package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm/clause"
	"time"
)

func (s Store) ListEvaluationRuns(ctx context.Context, tenantID tenant.ID, page ports.EvaluationRunPageRequest) ([]ports.EvaluationRunHead, error) {
	if page.Limit < 1 || page.Limit > ports.MaxEvaluationRunPageLimit {
		return nil, ports.ErrInvalidEvaluationRunPage
	}
	query := s.db.WithContext(ctx).Select("tenant_id", "id", "state", "version", "workflow_id", "revision_id", "total_cases", "completed_cases", "passed_cases", "created_at", "updated_at").Where(map[string]any{"tenant_id": tenantID.String()})
	if page.AfterID != "" {
		query = query.Where(clause.Gt{Column: "id", Value: string(page.AfterID)})
	}
	var rows []evaluationRunModel
	if err := query.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Limit(page.Limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]ports.EvaluationRunHead, 0, len(rows))
	for _, row := range rows {
		result = append(result, evaluationRunHeadFromModel(row))
	}
	return result, nil
}
func (s Store) RunnableEvaluationRuns(ctx context.Context, now time.Time, limit int) ([]ports.EvaluationRunReference, error) {
	if now.IsZero() || limit < 1 || limit > ports.MaxEvaluationRunPageLimit {
		return nil, ports.ErrInvalidEvaluationRunPage
	}
	query := s.db.WithContext(ctx).Select("tenant_id", "id").Where(clause.Or(clause.Eq{Column: "state", Value: string(agentmodel.EvaluationRunQueued)}, clause.And(clause.Eq{Column: "state", Value: string(agentmodel.EvaluationRunRunning)}, clause.Lte{Column: "lease_until", Value: now})))
	var rows []evaluationRunModel
	if err := query.Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}}).Order(clause.OrderByColumn{Column: clause.Column{Name: "tenant_id"}}).Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]ports.EvaluationRunReference, 0, len(rows))
	for _, row := range rows {
		result = append(result, ports.EvaluationRunReference{TenantID: tenant.ID(row.TenantID), ID: agentmodel.EvaluationRunID(row.ID)})
	}
	return result, nil
}
