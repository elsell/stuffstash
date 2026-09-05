package gormstore

import (
	"context"
	"errors"

	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) EvaluationCaseHead(ctx context.Context, tenantID tenant.ID, id domain.EvaluationCaseID) (ports.EvaluationCaseHeadRecord, bool, error) {
	var model evaluationCaseModel
	err := s.db.WithContext(ctx).Where(map[string]any{"tenant_id": tenantID.String(), "id": string(id)}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ports.EvaluationCaseHeadRecord{}, false, nil
	}
	if err != nil {
		return ports.EvaluationCaseHeadRecord{}, false, err
	}
	return evaluationCaseHeadFromModel(model), true, nil
}
func (s Store) EvaluationCaseRevision(ctx context.Context, tenantID tenant.ID, caseID domain.EvaluationCaseID, id domain.EvaluationCaseRevisionID) (domain.EvaluationCaseRevision, bool, error) {
	var model evaluationCaseRevisionModel
	err := s.db.WithContext(ctx).Where(map[string]any{"tenant_id": tenantID.String(), "case_id": string(caseID), "id": string(id)}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.EvaluationCaseRevision{}, false, nil
	}
	if err != nil {
		return domain.EvaluationCaseRevision{}, false, err
	}
	result, err := evaluationCaseRevisionFromModel(model)
	return result, err == nil, err
}
func (s Store) ListEvaluationCases(ctx context.Context, tenantID tenant.ID, page ports.EvaluationCasePageRequest) ([]ports.EvaluationCaseHeadRecord, error) {
	if page.Limit < 1 || page.Limit > ports.MaxEvaluationCasePageLimit {
		return nil, ports.ErrInvalidEvaluationCasePage
	}
	query := s.db.WithContext(ctx).Where(map[string]any{"tenant_id": tenantID.String()})
	if page.AfterID != "" {
		query = query.Where(clause.Gt{Column: "id", Value: string(page.AfterID)})
	}
	var models []evaluationCaseModel
	if err := query.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Limit(page.Limit).Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]ports.EvaluationCaseHeadRecord, 0, len(models))
	for _, model := range models {
		result = append(result, evaluationCaseHeadFromModel(model))
	}
	return result, nil
}
func (s Store) AppendEvaluationCaseRevision(ctx context.Context, revision domain.EvaluationCaseRevision, expected int, record audit.Record) error {
	model, err := evaluationCaseRevisionToModel(revision)
	if err != nil {
		return err
	}
	if expected < 0 || model.Number != expected+1 {
		return ports.ErrEvaluationCaseConflict
	}
	if string(record.TenantID) != model.TenantID || record.InventoryID != "" || record.ID == "" || record.Action != audit.ActionConversationEvaluationCaseRevisionCreated {
		return domain.ErrInvalidEvaluationCaseRevision
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		title := revision.Snapshot().Definition.Settings().Title
		if expected == 0 {
			head := evaluationCaseModel{TenantID: model.TenantID, ID: model.CaseID, Title: title, LatestRevision: model.Number, LatestRevisionID: model.ID, CreatedAt: model.CreatedAt, UpdatedAt: model.CreatedAt}
			result := tx.Omit(clause.Associations).Clauses(clause.OnConflict{DoNothing: true}).Create(&head)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return ports.ErrEvaluationCaseConflict
			}
		} else {
			result := tx.Model(&evaluationCaseModel{}).Where(map[string]any{"tenant_id": model.TenantID, "id": model.CaseID, "latest_revision": expected}).Updates(map[string]any{"title": title, "latest_revision": model.Number, "latest_revision_id": model.ID, "updated_at": model.CreatedAt})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return ports.ErrEvaluationCaseConflict
			}
		}
		result := tx.Omit(clause.Associations).Clauses(clause.OnConflict{DoNothing: true}).Create(&model)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ports.ErrEvaluationCaseConflict
		}
		return createAuditRecord(tx, record)
	})
}
