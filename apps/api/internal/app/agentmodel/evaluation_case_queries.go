package agentmodel

import (
	"context"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const evaluationCaseCollection = "conversation_evaluation_cases"

type GetEvaluationCaseInput struct {
	EvaluationCaseAccess
	CaseID     domain.EvaluationCaseID
	RevisionID domain.EvaluationCaseRevisionID
}
type ListEvaluationCasesInput struct {
	EvaluationCaseAccess
	Limit  int
	Cursor string
}
type ListEvaluationCasesResult struct {
	Items      []ports.EvaluationCaseHeadRecord
	Limit      int
	NextCursor *string
}

func (s EvaluationCaseService) Get(ctx context.Context, input GetEvaluationCaseInput) (domain.EvaluationCaseRevision, error) {
	if err := s.authorize(ctx, input.EvaluationCaseAccess); err != nil {
		return domain.EvaluationCaseRevision{}, err
	}
	if s.deps.Audit == nil {
		return domain.EvaluationCaseRevision{}, apperrors.ErrPrecondition
	}
	revisionID := input.RevisionID
	if revisionID == "" {
		head, found, err := s.deps.Repository.EvaluationCaseHead(ctx, input.TenantID, input.CaseID)
		if err != nil {
			return domain.EvaluationCaseRevision{}, err
		}
		if !found {
			return domain.EvaluationCaseRevision{}, apperrors.ErrNotFound
		}
		revisionID = head.LatestRevisionID
	}
	revision, found, err := s.deps.Repository.EvaluationCaseRevision(ctx, input.TenantID, input.CaseID, revisionID)
	if err != nil {
		return domain.EvaluationCaseRevision{}, err
	}
	if !found {
		return domain.EvaluationCaseRevision{}, apperrors.ErrNotFound
	}
	record, err := s.auditRecord(input.EvaluationCaseAccess, audit.ActionConversationEvaluationCaseViewed, string(input.CaseID), map[string]string{"revision_id": string(revisionID)})
	if err != nil {
		return domain.EvaluationCaseRevision{}, err
	}
	if err := s.deps.Audit.SaveAuditRecord(ctx, record); err != nil {
		return domain.EvaluationCaseRevision{}, err
	}
	return revision, nil
}
func (s EvaluationCaseService) List(ctx context.Context, input ListEvaluationCasesInput) (ListEvaluationCasesResult, error) {
	if err := s.authorize(ctx, input.EvaluationCaseAccess); err != nil {
		return ListEvaluationCasesResult{}, err
	}
	if s.deps.Audit == nil {
		return ListEvaluationCasesResult{}, apperrors.ErrPrecondition
	}
	if input.Limit < 0 {
		return ListEvaluationCasesResult{}, apperrors.ErrValidation
	}
	limit := appsupport.PageLimit(s.deps.DefaultPageLimit, s.deps.MaxPageLimit, input.Limit)
	after, err := appsupport.DecodePageCursor(evaluationCaseCollection, input.TenantID.String(), input.Cursor)
	if err != nil {
		return ListEvaluationCasesResult{}, apperrors.ErrValidation
	}
	rows, err := s.deps.Repository.ListEvaluationCases(ctx, input.TenantID, ports.EvaluationCasePageRequest{AfterID: domain.EvaluationCaseID(after), Limit: min(limit+1, ports.MaxEvaluationCasePageLimit)})
	if err != nil {
		return ListEvaluationCasesResult{}, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	} else if len(rows) == ports.MaxEvaluationCasePageLimit {
		probe, err := s.deps.Repository.ListEvaluationCases(ctx, input.TenantID, ports.EvaluationCasePageRequest{AfterID: rows[len(rows)-1].ID, Limit: 1})
		if err != nil {
			return ListEvaluationCasesResult{}, err
		}
		hasMore = len(probe) > 0
	}
	var cursor *string
	if hasMore {
		cursor = appsupport.EncodePageCursor(evaluationCaseCollection, input.TenantID.String(), string(rows[len(rows)-1].ID))
	}
	record, err := s.auditRecord(input.EvaluationCaseAccess, audit.ActionConversationEvaluationCaseListed, input.TenantID.String(), map[string]string{"returned_count": strconv.Itoa(len(rows))})
	if err != nil {
		return ListEvaluationCasesResult{}, err
	}
	if err := s.deps.Audit.SaveAuditRecord(ctx, record); err != nil {
		return ListEvaluationCasesResult{}, err
	}
	return ListEvaluationCasesResult{Items: rows, Limit: limit, NextCursor: cursor}, nil
}
