package gormstore

import (
	"context"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm/clause"
)

func (s Store) ListWorkflowHeads(ctx context.Context, scope tenant.ID, page ports.WorkflowHeadPageRequest) ([]ports.WorkflowHeadRecord, error) {
	if page.Limit < 1 || page.Limit > ports.MaxWorkflowPageLimit {
		return nil, ports.ErrInvalidWorkflowPage
	}
	query := s.db.WithContext(ctx).Where(map[string]any{"tenant_id": scope.String()})
	if page.AfterID != "" {
		query = query.Where(clause.Gt{Column: "id", Value: string(page.AfterID)})
	}
	var rows []conversationWorkflowModel
	if err := query.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Limit(page.Limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]ports.WorkflowHeadRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, ports.WorkflowHeadRecord{TenantID: scope, ID: model.WorkflowID(row.ID), Name: row.Name, LatestRevisionID: model.WorkflowRevisionID(row.LatestRevisionID), LatestRevision: row.LatestRevision, ActiveRevisionID: model.WorkflowRevisionID(row.ActiveRevisionID), CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt})
	}
	return result, nil
}
func (s Store) ListWorkflowRevisions(ctx context.Context, scope tenant.ID, workflow model.WorkflowID, page ports.WorkflowRevisionPageRequest) ([]model.WorkflowRevision, error) {
	if page.Limit < 1 || page.Limit > ports.MaxWorkflowPageLimit || page.AfterNumber < 0 {
		return nil, ports.ErrInvalidWorkflowPage
	}
	var rows []conversationWorkflowRevisionModel
	err := s.db.WithContext(ctx).Where(map[string]any{"tenant_id": scope.String(), "workflow_id": string(workflow)}).Where(clause.Gt{Column: "number", Value: page.AfterNumber}).Order(clause.OrderByColumn{Column: clause.Column{Name: "number"}}).Limit(page.Limit).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]model.WorkflowRevision, 0, len(rows))
	for _, row := range rows {
		revision, err := workflowRevisionFromModel(row)
		if err != nil {
			return nil, err
		}
		result = append(result, revision)
	}
	return result, nil
}
