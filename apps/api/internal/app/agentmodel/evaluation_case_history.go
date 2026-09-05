package agentmodel

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"strconv"
)

type ListEvaluationCaseRevisionsInput struct {
	EvaluationCaseAccess
	CaseID domain.EvaluationCaseID
	Limit  int
	Cursor string
}
type ListEvaluationCaseRevisionsResult struct {
	Items      []domain.EvaluationCaseRevision
	Limit      int
	NextCursor *string
}

func (s EvaluationCaseService) History(ctx context.Context, input ListEvaluationCaseRevisionsInput) (ListEvaluationCaseRevisionsResult, error) {
	if err := s.authorize(ctx, input.EvaluationCaseAccess); err != nil {
		return ListEvaluationCaseRevisionsResult{}, err
	}
	if s.deps.Audit == nil {
		return ListEvaluationCaseRevisionsResult{}, apperrors.ErrPrecondition
	}
	if input.Limit < 0 || input.CaseID == "" {
		return ListEvaluationCaseRevisionsResult{}, apperrors.ErrValidation
	}
	collection := "conversation_evaluation_case_revisions:" + string(input.CaseID)
	after, err := appsupport.DecodePageCursor(collection, input.TenantID.String(), input.Cursor)
	if err != nil {
		return ListEvaluationCaseRevisionsResult{}, apperrors.ErrValidation
	}
	number := 0
	if after != "" {
		number, err = strconv.Atoi(after)
		if err != nil || number < 1 {
			return ListEvaluationCaseRevisionsResult{}, apperrors.ErrValidation
		}
	}
	if _, found, err := s.deps.Repository.EvaluationCaseHead(ctx, input.TenantID, input.CaseID); err != nil {
		return ListEvaluationCaseRevisionsResult{}, err
	} else if !found {
		return ListEvaluationCaseRevisionsResult{}, apperrors.ErrNotFound
	}
	limit := appsupport.PageLimit(s.deps.DefaultPageLimit, s.deps.MaxPageLimit, input.Limit)
	rows, err := s.deps.Repository.ListEvaluationCaseRevisions(ctx, input.TenantID, input.CaseID, ports.EvaluationCaseRevisionPageRequest{AfterNumber: number, Limit: min(limit+1, ports.MaxEvaluationCasePageLimit)})
	if err != nil {
		return ListEvaluationCaseRevisionsResult{}, err
	}
	more := len(rows) > limit
	if more {
		rows = rows[:limit]
	} else if len(rows) == ports.MaxEvaluationCasePageLimit {
		probe, err := s.deps.Repository.ListEvaluationCaseRevisions(ctx, input.TenantID, input.CaseID, ports.EvaluationCaseRevisionPageRequest{AfterNumber: rows[len(rows)-1].Snapshot().Number, Limit: 1})
		if err != nil {
			return ListEvaluationCaseRevisionsResult{}, err
		}
		more = len(probe) > 0
	}
	var cursor *string
	if more {
		cursor = appsupport.EncodePageCursor(collection, input.TenantID.String(), strconv.Itoa(rows[len(rows)-1].Snapshot().Number))
	}
	record, err := s.auditRecord(input.EvaluationCaseAccess, audit.ActionConversationEvaluationCaseListed, string(input.CaseID), map[string]string{"returned_count": strconv.Itoa(len(rows))})
	if err != nil {
		return ListEvaluationCaseRevisionsResult{}, err
	}
	if err := s.deps.Audit.SaveAuditRecord(ctx, record); err != nil {
		return ListEvaluationCaseRevisionsResult{}, err
	}
	return ListEvaluationCaseRevisionsResult{Items: rows, Limit: limit, NextCursor: cursor}, nil
}
