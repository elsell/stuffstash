package gormstore

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) WorkflowHead(ctx context.Context, tenantID tenant.ID, id agentmodel.WorkflowID) (ports.WorkflowHeadRecord, bool, error) {
	var model conversationWorkflowModel
	err := s.db.WithContext(ctx).Where(map[string]any{"tenant_id": tenantID.String(), "id": string(id)}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ports.WorkflowHeadRecord{}, false, nil
	}
	if err != nil {
		return ports.WorkflowHeadRecord{}, false, err
	}
	return ports.WorkflowHeadRecord{TenantID: tenantID, ID: id, Name: model.Name, LatestRevisionID: agentmodel.WorkflowRevisionID(model.LatestRevisionID), LatestRevision: model.LatestRevision, ActiveRevisionID: agentmodel.WorkflowRevisionID(model.ActiveRevisionID), CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt}, true, nil
}

func (s Store) WorkflowRevision(ctx context.Context, tenantID tenant.ID, workflowID agentmodel.WorkflowID, id agentmodel.WorkflowRevisionID) (agentmodel.WorkflowRevision, bool, error) {
	var model conversationWorkflowRevisionModel
	err := s.db.WithContext(ctx).Where(map[string]any{"tenant_id": tenantID.String(), "workflow_id": string(workflowID), "id": string(id)}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return agentmodel.WorkflowRevision{}, false, nil
	}
	if err != nil {
		return agentmodel.WorkflowRevision{}, false, err
	}
	result, err := workflowRevisionFromModel(model)
	return result, err == nil, err
}

func (s Store) AppendWorkflowRevision(ctx context.Context, revision agentmodel.WorkflowRevision, expected int, record audit.Record) error {
	model, err := workflowRevisionModel(revision)
	if err != nil {
		return err
	}
	if expected < 0 || model.Number != expected+1 {
		return ports.ErrWorkflowConflict
	}
	if string(record.TenantID) != model.TenantID || record.Action != audit.ActionConversationWorkflowRevisionCreated {
		return agentmodel.ErrInvalidWorkflowRevision
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if expected == 0 {
			head := conversationWorkflowModel{Name: revision.Snapshot().Definition.Settings().Name, LatestRevisionID: model.ID, TenantID: model.TenantID, ID: model.WorkflowID, LatestRevision: model.Number, CreatedAt: model.CreatedAt, UpdatedAt: model.CreatedAt}
			result := tx.Omit(clause.Associations).Clauses(clause.OnConflict{DoNothing: true}).Create(&head)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return ports.ErrWorkflowConflict
			}
		} else {
			result := tx.Model(&conversationWorkflowModel{}).Where(map[string]any{"tenant_id": model.TenantID, "id": model.WorkflowID, "latest_revision": expected}).Updates(map[string]any{"latest_revision": model.Number, "updated_at": model.CreatedAt, "name": revision.Snapshot().Definition.Settings().Name, "latest_revision_id": model.ID})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return ports.ErrWorkflowConflict
			}
		}
		result := tx.Omit(clause.Associations).Clauses(clause.OnConflict{DoNothing: true}).Create(&model)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ports.ErrWorkflowConflict
		}
		return createAuditRecord(tx, record)
	})
}

func (s Store) ActivateWorkflowRevision(ctx context.Context, tenantID tenant.ID, workflowID agentmodel.WorkflowID, revisionID agentmodel.WorkflowRevisionID, expectedActive ports.WorkflowSelectionReference, at time.Time, record audit.Record) error {
	if at.IsZero() || string(record.TenantID) != tenantID.String() || record.Action != audit.ActionConversationWorkflowActivated {
		return agentmodel.ErrInvalidWorkflowRevision
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var revision conversationWorkflowRevisionModel
		err := tx.Where(map[string]any{"tenant_id": tenantID.String(), "workflow_id": string(workflowID), "id": string(revisionID)}).First(&revision).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrWorkflowNotFound
		}
		if err != nil {
			return err
		}
		if (expectedActive.WorkflowID == "") != (expectedActive.RevisionID == "") {
			return ports.ErrWorkflowConflict
		}
		if expectedActive == (ports.WorkflowSelectionReference{}) {
			result := tx.Omit(clause.Associations).Clauses(clause.OnConflict{DoNothing: true}).Create(&conversationWorkflowSelectionModel{TenantID: tenantID.String(), WorkflowID: string(workflowID), RevisionID: string(revisionID), ActivatedAt: at})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return ports.ErrWorkflowConflict
			}
		} else {
			result := tx.Model(&conversationWorkflowSelectionModel{}).Where(map[string]any{"tenant_id": tenantID.String(), "workflow_id": string(expectedActive.WorkflowID), "revision_id": string(expectedActive.RevisionID)}).Updates(map[string]any{"workflow_id": string(workflowID), "revision_id": string(revisionID), "activated_at": at})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return ports.ErrWorkflowConflict
			}
		}
		result := tx.Model(&conversationWorkflowModel{}).Where(map[string]any{"tenant_id": tenantID.String(), "id": string(workflowID)}).Updates(map[string]any{"active_revision_id": string(revisionID), "updated_at": at})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ports.ErrWorkflowConflict
		}
		return createAuditRecord(tx, record)
	})
}

func (s Store) SelectedWorkflowRevision(ctx context.Context, tenantID tenant.ID) (ports.WorkflowSelectionReference, bool, error) {
	var model conversationWorkflowSelectionModel
	err := s.db.WithContext(ctx).Where(map[string]any{"tenant_id": tenantID.String()}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ports.WorkflowSelectionReference{}, false, nil
	}
	if err != nil {
		return ports.WorkflowSelectionReference{}, false, err
	}
	return ports.WorkflowSelectionReference{WorkflowID: agentmodel.WorkflowID(model.WorkflowID), RevisionID: agentmodel.WorkflowRevisionID(model.RevisionID)}, true, nil
}
