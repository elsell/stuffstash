package agentmodel

import (
	"context"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type EvaluationRunQueryDependencies struct {
	Authorizer                     ports.Authorizer
	Runs                           ports.EvaluationRunRepository
	Audit                          ports.AuditRepository
	IDs                            ports.IDGenerator
	Clock                          ports.Clock
	DefaultPageLimit, MaxPageLimit int
}
type EvaluationRunQueryService struct {
	deps EvaluationRunQueryDependencies
}

func NewEvaluationRunQueryService(deps EvaluationRunQueryDependencies) EvaluationRunQueryService {
	if deps.MaxPageLimit <= 0 || deps.MaxPageLimit > ports.MaxEvaluationRunPageLimit {
		deps.MaxPageLimit = ports.MaxEvaluationRunPageLimit
	}
	if deps.DefaultPageLimit <= 0 {
		deps.DefaultPageLimit = 50
	}
	deps.DefaultPageLimit = min(deps.DefaultPageLimit, deps.MaxPageLimit)
	return EvaluationRunQueryService{deps: deps}
}

type GetEvaluationRunInput struct {
	EvaluationRunAccess
	RunID model.EvaluationRunID
}
type ListEvaluationRunsInput struct {
	EvaluationRunAccess
	Limit  int
	Cursor string
}
type ListEvaluationRunsResult struct {
	Items      []ports.EvaluationRunHead
	Limit      int
	NextCursor *string
}

func (s EvaluationRunQueryService) authorize(ctx context.Context, input EvaluationRunAccess) error {
	if err := authorizeEvaluationRunAccess(ctx, s.deps.Authorizer, input); err != nil {
		return err
	}
	if s.deps.Runs == nil || s.deps.Audit == nil || s.deps.IDs == nil || s.deps.Clock == nil {
		return apperrors.ErrPrecondition
	}
	return nil
}
func (s EvaluationRunQueryService) Get(ctx context.Context, input GetEvaluationRunInput) (model.EvaluationRun, error) {
	if err := s.authorize(ctx, input.EvaluationRunAccess); err != nil {
		return model.EvaluationRun{}, err
	}
	if input.RunID == "" {
		return model.EvaluationRun{}, apperrors.ErrValidation
	}
	run, found, err := s.deps.Runs.EvaluationRun(ctx, input.TenantID, input.RunID)
	if err != nil {
		return model.EvaluationRun{}, err
	}
	if !found {
		return model.EvaluationRun{}, apperrors.ErrNotFound
	}
	if err := s.auditRead(ctx, input.EvaluationRunAccess, audit.ActionConversationEvaluationRunViewed, string(input.RunID), nil); err != nil {
		return model.EvaluationRun{}, err
	}
	return run, nil
}
func (s EvaluationRunQueryService) List(ctx context.Context, input ListEvaluationRunsInput) (ListEvaluationRunsResult, error) {
	if err := s.authorize(ctx, input.EvaluationRunAccess); err != nil {
		return ListEvaluationRunsResult{}, err
	}
	if input.Limit < 0 {
		return ListEvaluationRunsResult{}, apperrors.ErrValidation
	}
	const collection = "conversation_evaluation_runs"
	after, err := appsupport.DecodePageCursor(collection, input.TenantID.String(), input.Cursor)
	if err != nil {
		return ListEvaluationRunsResult{}, apperrors.ErrValidation
	}
	limit := appsupport.PageLimit(s.deps.DefaultPageLimit, s.deps.MaxPageLimit, input.Limit)
	rows, err := s.deps.Runs.ListEvaluationRuns(ctx, input.TenantID, ports.EvaluationRunPageRequest{AfterID: model.EvaluationRunID(after), Limit: min(limit+1, ports.MaxEvaluationRunPageLimit)})
	if err != nil {
		return ListEvaluationRunsResult{}, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	} else if len(rows) == ports.MaxEvaluationRunPageLimit {
		probe, err := s.deps.Runs.ListEvaluationRuns(ctx, input.TenantID, ports.EvaluationRunPageRequest{AfterID: rows[len(rows)-1].ID, Limit: 1})
		if err != nil {
			return ListEvaluationRunsResult{}, err
		}
		hasMore = len(probe) > 0
	}
	var cursor *string
	if hasMore {
		cursor = appsupport.EncodePageCursor(collection, input.TenantID.String(), string(rows[len(rows)-1].ID))
	}
	if err := s.auditRead(ctx, input.EvaluationRunAccess, audit.ActionConversationEvaluationRunListed, input.TenantID.String(), map[string]string{"returned_count": strconv.Itoa(len(rows))}); err != nil {
		return ListEvaluationRunsResult{}, err
	}
	return ListEvaluationRunsResult{Items: rows, Limit: limit, NextCursor: cursor}, nil
}
func (s EvaluationRunQueryService) auditRead(ctx context.Context, access EvaluationRunAccess, action audit.Action, target string, metadata map[string]string) error {
	record, ok := audit.NewRecord(audit.ID(s.deps.IDs.NewID()), audit.TenantID(access.TenantID), "", audit.PrincipalID(access.Principal.ID), action, access.Source, audit.TargetConversationEvaluationRun, target, s.deps.Clock.Now(), access.RequestID, metadata)
	if !ok {
		return apperrors.ErrValidation
	}
	return s.deps.Audit.SaveAuditRecord(ctx, record)
}
