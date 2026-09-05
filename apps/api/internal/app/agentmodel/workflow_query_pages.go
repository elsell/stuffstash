package agentmodel

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"strconv"
)

func (s WorkflowQueryService) List(ctx context.Context, input ListWorkflowsInput) (ListWorkflowsResult, error) {
	if err := s.authorize(ctx, input.EvaluationRunAccess); err != nil {
		return ListWorkflowsResult{}, err
	}
	if input.Limit < 0 {
		return ListWorkflowsResult{}, apperrors.ErrValidation
	}
	const collection = "conversation_workflows"
	after, err := appsupport.DecodePageCursor(collection, input.TenantID.String(), input.Cursor)
	if err != nil {
		return ListWorkflowsResult{}, apperrors.ErrValidation
	}
	limit := appsupport.PageLimit(s.deps.DefaultPageLimit, s.deps.MaxPageLimit, input.Limit)
	rows, err := s.deps.Discovery.ListWorkflowHeads(ctx, input.TenantID, ports.WorkflowHeadPageRequest{AfterID: model.WorkflowID(after), Limit: min(limit+1, ports.MaxWorkflowPageLimit)})
	if err != nil {
		return ListWorkflowsResult{}, err
	}
	more := len(rows) > limit
	if more {
		rows = rows[:limit]
	} else if len(rows) == ports.MaxWorkflowPageLimit {
		probe, err := s.deps.Discovery.ListWorkflowHeads(ctx, input.TenantID, ports.WorkflowHeadPageRequest{AfterID: rows[len(rows)-1].ID, Limit: 1})
		if err != nil {
			return ListWorkflowsResult{}, err
		}
		more = len(probe) > 0
	}
	var cursor *string
	if more {
		cursor = appsupport.EncodePageCursor(collection, input.TenantID.String(), string(rows[len(rows)-1].ID))
	}
	if err := s.auditRead(ctx, input.EvaluationRunAccess, audit.ActionConversationWorkflowListed, input.TenantID.String(), map[string]string{"returned_count": strconv.Itoa(len(rows))}); err != nil {
		return ListWorkflowsResult{}, err
	}
	return ListWorkflowsResult{Items: rows, Limit: limit, NextCursor: cursor}, nil
}
func (s WorkflowQueryService) History(ctx context.Context, input ListWorkflowRevisionsInput) (ListWorkflowRevisionsResult, error) {
	if err := s.authorize(ctx, input.EvaluationRunAccess); err != nil {
		return ListWorkflowRevisionsResult{}, err
	}
	if input.Limit < 0 || input.WorkflowID == "" {
		return ListWorkflowRevisionsResult{}, apperrors.ErrValidation
	}
	collection := "conversation_workflow_revisions:" + string(input.WorkflowID)
	after, err := appsupport.DecodePageCursor(collection, input.TenantID.String(), input.Cursor)
	if err != nil {
		return ListWorkflowRevisionsResult{}, apperrors.ErrValidation
	}
	number := 0
	if after != "" {
		number, err = strconv.Atoi(after)
		if err != nil || number < 1 {
			return ListWorkflowRevisionsResult{}, apperrors.ErrValidation
		}
	}
	if _, found, err := s.deps.Repository.WorkflowHead(ctx, input.TenantID, input.WorkflowID); err != nil {
		return ListWorkflowRevisionsResult{}, err
	} else if !found {
		return ListWorkflowRevisionsResult{}, apperrors.ErrNotFound
	}
	limit := appsupport.PageLimit(s.deps.DefaultPageLimit, s.deps.MaxPageLimit, input.Limit)
	rows, err := s.deps.Discovery.ListWorkflowRevisions(ctx, input.TenantID, input.WorkflowID, ports.WorkflowRevisionPageRequest{AfterNumber: number, Limit: min(limit+1, ports.MaxWorkflowPageLimit)})
	if err != nil {
		return ListWorkflowRevisionsResult{}, err
	}
	more := len(rows) > limit
	if more {
		rows = rows[:limit]
	} else if len(rows) == ports.MaxWorkflowPageLimit {
		probe, err := s.deps.Discovery.ListWorkflowRevisions(ctx, input.TenantID, input.WorkflowID, ports.WorkflowRevisionPageRequest{AfterNumber: rows[len(rows)-1].Snapshot().Number, Limit: 1})
		if err != nil {
			return ListWorkflowRevisionsResult{}, err
		}
		more = len(probe) > 0
	}
	var cursor *string
	if more {
		cursor = appsupport.EncodePageCursor(collection, input.TenantID.String(), strconv.Itoa(rows[len(rows)-1].Snapshot().Number))
	}
	if err := s.auditRead(ctx, input.EvaluationRunAccess, audit.ActionConversationWorkflowListed, string(input.WorkflowID), map[string]string{"returned_count": strconv.Itoa(len(rows))}); err != nil {
		return ListWorkflowRevisionsResult{}, err
	}
	return ListWorkflowRevisionsResult{Items: rows, Limit: limit, NextCursor: cursor}, nil
}
