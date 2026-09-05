package gormstore

import (
	"context"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm/clause"
)

func (s Store) ListEvaluationCaseRevisions(ctx context.Context, scope tenant.ID, caseID domain.EvaluationCaseID, page ports.EvaluationCaseRevisionPageRequest) ([]domain.EvaluationCaseRevision, error) {
	if page.Limit < 1 || page.Limit > ports.MaxEvaluationCasePageLimit || page.AfterNumber < 0 {
		return nil, ports.ErrInvalidEvaluationCasePage
	}
	var rows []evaluationCaseRevisionModel
	err := s.db.WithContext(ctx).Where(map[string]any{"tenant_id": scope.String(), "case_id": string(caseID)}).Where(clause.Gt{Column: "number", Value: page.AfterNumber}).Order(clause.OrderByColumn{Column: clause.Column{Name: "number"}}).Limit(page.Limit).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]domain.EvaluationCaseRevision, 0, len(rows))
	for _, row := range rows {
		value, err := evaluationCaseRevisionFromModel(row)
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, nil
}
