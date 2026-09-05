package gormstore

import (
	"context"
	"errors"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) EvaluationRun(ctx context.Context, tenantID tenant.ID, id model.EvaluationRunID) (model.EvaluationRun, bool, error) {
	var value evaluationRunModel
	err := s.db.WithContext(ctx).Where(map[string]any{"tenant_id": tenantID.String(), "id": string(id)}).First(&value).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.EvaluationRun{}, false, nil
	}
	if err != nil {
		return model.EvaluationRun{}, false, err
	}
	run, err := evaluationRunFromModel(value)
	return run, err == nil, err
}
func (s Store) SaveEvaluationRun(ctx context.Context, run model.EvaluationRun, expected int, record audit.Record) error {
	value, err := evaluationRunToModel(run)
	if err != nil {
		return err
	}
	if expected < 0 || value.Version != expected+1 {
		return ports.ErrEvaluationRunConflict
	}
	action := audit.ActionConversationEvaluationRunProgressed
	if expected == 0 {
		action = audit.ActionConversationEvaluationRunCreated
	} else if run.Snapshot().State == model.EvaluationRunCancelled {
		action = audit.ActionConversationEvaluationRunCancelled
	}
	if record.TenantID.String() != value.TenantID || record.InventoryID != "" || record.ID == "" || record.Action != action || record.TargetType != audit.TargetConversationEvaluationRun || record.TargetID != value.ID {
		return model.ErrInvalidEvaluationRun
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if expected == 0 {
			if value.State != string(model.EvaluationRunQueued) {
				return model.ErrInvalidEvaluationRun
			}
			result := tx.Omit(clause.Associations).Clauses(clause.OnConflict{DoNothing: true}).Create(&value)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return ports.ErrEvaluationRunConflict
			}
		} else {
			var stored evaluationRunModel
			err := tx.Where(map[string]any{"tenant_id": value.TenantID, "id": value.ID, "version": expected}).First(&stored).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ports.ErrEvaluationRunConflict
			}
			if err != nil {
				return err
			}
			previous, err := evaluationRunFromModel(stored)
			if err != nil {
				return err
			}
			if !run.IsSuccessorOf(previous) {
				return ports.ErrEvaluationRunConflict
			}
			result := tx.Model(&evaluationRunModel{}).Where(map[string]any{"tenant_id": value.TenantID, "id": value.ID, "version": expected}).Updates(map[string]any{"state": value.State, "version": value.Version, "lease_until": value.LeaseUntil, "completed_cases": value.CompletedCases, "passed_cases": value.PassedCases, "progress_json": value.ProgressJSON, "updated_at": value.UpdatedAt})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return ports.ErrEvaluationRunConflict
			}
		}
		return createAuditRecord(tx, record)
	})
}
