package agentmodel

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type WorkflowQueryDependencies struct {
	Authorizer                     ports.Authorizer
	Repository                     ports.ConversationWorkflowRepository
	Discovery                      ports.WorkflowDiscoveryRepository
	Audit                          ports.AuditRepository
	IDs                            ports.IDGenerator
	Clock                          ports.Clock
	DefaultPageLimit, MaxPageLimit int
}
type WorkflowQueryService struct{ deps WorkflowQueryDependencies }

func NewWorkflowQueryService(deps WorkflowQueryDependencies) WorkflowQueryService {
	if deps.MaxPageLimit <= 0 || deps.MaxPageLimit > ports.MaxWorkflowPageLimit {
		deps.MaxPageLimit = ports.MaxWorkflowPageLimit
	}
	if deps.DefaultPageLimit <= 0 {
		deps.DefaultPageLimit = 50
	}
	deps.DefaultPageLimit = min(deps.DefaultPageLimit, deps.MaxPageLimit)
	return WorkflowQueryService{deps: deps}
}

type GetWorkflowInput struct {
	EvaluationRunAccess
	WorkflowID model.WorkflowID
	RevisionID model.WorkflowRevisionID
}
type ListWorkflowsInput struct {
	EvaluationRunAccess
	Limit  int
	Cursor string
}
type ListWorkflowsResult struct {
	Items      []ports.WorkflowHeadRecord
	Limit      int
	NextCursor *string
}
type ListWorkflowRevisionsInput struct {
	EvaluationRunAccess
	WorkflowID model.WorkflowID
	Limit      int
	Cursor     string
}
type ListWorkflowRevisionsResult struct {
	Items      []model.WorkflowRevision
	Limit      int
	NextCursor *string
}

func (s WorkflowQueryService) authorize(ctx context.Context, access EvaluationRunAccess) error {
	if err := authorizeEvaluationRunAccess(ctx, s.deps.Authorizer, access); err != nil {
		return err
	}
	if s.deps.Repository == nil || s.deps.Discovery == nil || s.deps.Audit == nil || s.deps.IDs == nil || s.deps.Clock == nil {
		return apperrors.ErrPrecondition
	}
	return nil
}
func (s WorkflowQueryService) Get(ctx context.Context, input GetWorkflowInput) (model.WorkflowRevision, error) {
	if err := s.authorize(ctx, input.EvaluationRunAccess); err != nil {
		return model.WorkflowRevision{}, err
	}
	if input.WorkflowID == "" {
		return model.WorkflowRevision{}, apperrors.ErrValidation
	}
	revisionID := input.RevisionID
	if revisionID == "" {
		head, found, err := s.deps.Repository.WorkflowHead(ctx, input.TenantID, input.WorkflowID)
		if err != nil {
			return model.WorkflowRevision{}, err
		}
		if !found {
			return model.WorkflowRevision{}, apperrors.ErrNotFound
		}
		revisionID = head.LatestRevisionID
	}
	revision, found, err := s.deps.Repository.WorkflowRevision(ctx, input.TenantID, input.WorkflowID, revisionID)
	if err != nil {
		return model.WorkflowRevision{}, err
	}
	if !found {
		return model.WorkflowRevision{}, apperrors.ErrNotFound
	}
	if err := s.auditRead(ctx, input.EvaluationRunAccess, audit.ActionConversationWorkflowViewed, string(input.WorkflowID), map[string]string{"revision_id": string(revisionID)}); err != nil {
		return model.WorkflowRevision{}, err
	}
	return revision, nil
}
func (s WorkflowQueryService) Selection(ctx context.Context, access EvaluationRunAccess) (*ports.WorkflowSelectionReference, error) {
	if err := s.authorize(ctx, access); err != nil {
		return nil, err
	}
	selected, found, err := s.deps.Repository.SelectedWorkflowRevision(ctx, access.TenantID)
	if err != nil {
		return nil, err
	}
	if err := s.auditRead(ctx, access, audit.ActionConversationWorkflowViewed, access.TenantID.String(), nil); err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return &selected, nil
}
func (s WorkflowQueryService) auditRead(ctx context.Context, access EvaluationRunAccess, action audit.Action, target string, metadata map[string]string) error {
	record, ok := audit.NewRecord(audit.ID(s.deps.IDs.NewID()), audit.TenantID(access.TenantID), "", audit.PrincipalID(access.Principal.ID), action, access.Source, audit.TargetConversationWorkflow, target, s.deps.Clock.Now(), access.RequestID, metadata)
	if !ok {
		return apperrors.ErrValidation
	}
	return s.deps.Audit.SaveAuditRecord(ctx, record)
}
