package agentmodel

import (
	"context"
	"errors"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type EvaluationRunCommandDependencies struct {
	Authorizer  ports.Authorizer
	Runs        ports.EvaluationRunRepository
	Workflows   ports.ConversationWorkflowRepository
	Cases       ports.EvaluationCaseRepository
	Providers   ports.EvaluationProviderSnapshotResolver
	IDs         ports.IDGenerator
	Clock       ports.Clock
	Observer    ports.Observer
	Limits      domain.WorkflowLimits
	MaxAttempts int
}
type EvaluationRunCommandService struct {
	deps EvaluationRunCommandDependencies
}

func NewEvaluationRunCommandService(deps EvaluationRunCommandDependencies) EvaluationRunCommandService {
	if deps.Observer == nil {
		deps.Observer = noopObserver{}
	}
	return EvaluationRunCommandService{deps: deps}
}

type EvaluationRunAccess struct {
	Principal identity.Principal
	TenantID  tenant.ID
	Source    audit.Source
	RequestID string
}
type EvaluationRunCaseReference struct {
	CaseID     domain.EvaluationCaseID
	RevisionID domain.EvaluationCaseRevisionID
}
type QueueEvaluationRunInput struct {
	EvaluationRunAccess
	WorkflowID domain.WorkflowID
	RevisionID domain.WorkflowRevisionID
	Cases      []EvaluationRunCaseReference
}
type CancelEvaluationRunInput struct {
	EvaluationRunAccess
	RunID           domain.EvaluationRunID
	ExpectedVersion int
}

func (s EvaluationRunCommandService) authorize(ctx context.Context, input EvaluationRunAccess) error {
	if err := authorizeEvaluationRunAccess(ctx, s.deps.Authorizer, input); err != nil {
		return err
	}
	if s.deps.Runs == nil || s.deps.IDs == nil || s.deps.Clock == nil {
		return apperrors.ErrPrecondition
	}
	return nil
}

func (s EvaluationRunCommandService) Queue(ctx context.Context, input QueueEvaluationRunInput) (domain.EvaluationRun, error) {
	if err := s.authorize(ctx, input.EvaluationRunAccess); err != nil {
		return domain.EvaluationRun{}, err
	}
	if s.deps.Workflows == nil || s.deps.Cases == nil || s.deps.Providers == nil {
		return domain.EvaluationRun{}, apperrors.ErrPrecondition
	}
	if input.WorkflowID == "" || input.RevisionID == "" || len(input.Cases) == 0 || len(input.Cases) > domain.MaxEvaluationRunCases {
		return domain.EvaluationRun{}, apperrors.ErrValidation
	}
	seen := make(map[domain.EvaluationCaseID]bool, len(input.Cases))
	for _, ref := range input.Cases {
		if ref.CaseID == "" || ref.RevisionID == "" || seen[ref.CaseID] {
			return domain.EvaluationRun{}, apperrors.ErrValidation
		}
		seen[ref.CaseID] = true
	}
	workflow, found, err := s.deps.Workflows.WorkflowRevision(ctx, input.TenantID, input.WorkflowID, input.RevisionID)
	if err != nil {
		return domain.EvaluationRun{}, err
	}
	if !found {
		return domain.EvaluationRun{}, apperrors.ErrNotFound
	}
	cases := make([]domain.EvaluationCaseRevision, 0, len(input.Cases))
	for _, ref := range input.Cases {
		revision, found, err := s.deps.Cases.EvaluationCaseRevision(ctx, input.TenantID, ref.CaseID, ref.RevisionID)
		if err != nil {
			return domain.EvaluationRun{}, err
		}
		if !found {
			return domain.EvaluationRun{}, apperrors.ErrNotFound
		}
		cases = append(cases, revision)
	}
	providers, err := s.deps.Providers.SnapshotEvaluationProviders(ctx, input.TenantID, workflow)
	if err != nil {
		return domain.EvaluationRun{}, err
	}
	run, err := domain.NewEvaluationRun(domain.EvaluationRunInput{ID: domain.EvaluationRunID(s.deps.IDs.NewID()), TenantID: domain.TenantID(input.TenantID), AuthorID: domain.WorkflowAuthorID(input.Principal.ID), CreatedAt: s.deps.Clock.Now(), Workflow: workflow, Cases: cases, Providers: providers, Limits: s.deps.Limits, MaxAttempts: s.deps.MaxAttempts})
	if err != nil {
		return domain.EvaluationRun{}, apperrors.ErrValidation
	}
	return s.save(ctx, input.EvaluationRunAccess, run, 0, audit.ActionConversationEvaluationRunCreated, ports.EventConversationEvaluationRunCreated)
}

func (s EvaluationRunCommandService) Cancel(ctx context.Context, input CancelEvaluationRunInput) (domain.EvaluationRun, error) {
	if err := s.authorize(ctx, input.EvaluationRunAccess); err != nil {
		return domain.EvaluationRun{}, err
	}
	if input.RunID == "" || input.ExpectedVersion < 1 {
		return domain.EvaluationRun{}, apperrors.ErrValidation
	}
	run, found, err := s.deps.Runs.EvaluationRun(ctx, input.TenantID, input.RunID)
	if err != nil {
		return domain.EvaluationRun{}, err
	}
	if !found {
		return domain.EvaluationRun{}, apperrors.ErrNotFound
	}
	current := run.Snapshot()
	if current.Version != input.ExpectedVersion {
		return domain.EvaluationRun{}, apperrors.ErrConflict
	}
	if current.State == domain.EvaluationRunCancelled {
		return run, nil
	}
	cancelled, err := run.Cancel(s.deps.Clock.Now())
	if err != nil {
		return domain.EvaluationRun{}, apperrors.ErrConflict
	}
	return s.save(ctx, input.EvaluationRunAccess, cancelled, input.ExpectedVersion, audit.ActionConversationEvaluationRunCancelled, ports.EventConversationEvaluationRunCancelled)
}

func (s EvaluationRunCommandService) save(ctx context.Context, access EvaluationRunAccess, run domain.EvaluationRun, expected int, action audit.Action, event ports.EventName) (domain.EvaluationRun, error) {
	snapshot := run.Snapshot()
	metadata := map[string]string{"state": string(snapshot.State), "completed_cases": strconv.Itoa(len(snapshot.Results)), "total_cases": strconv.Itoa(len(snapshot.Input.Cases))}
	record, ok := audit.NewRecord(audit.ID(s.deps.IDs.NewID()), audit.TenantID(access.TenantID), "", audit.PrincipalID(access.Principal.ID), action, access.Source, audit.TargetConversationEvaluationRun, string(snapshot.Input.ID), snapshot.UpdatedAt, access.RequestID, metadata)
	if !ok {
		return domain.EvaluationRun{}, apperrors.ErrValidation
	}
	if err := s.deps.Runs.SaveEvaluationRun(ctx, run, expected, record); err != nil {
		if errors.Is(err, ports.ErrEvaluationRunConflict) {
			err = apperrors.ErrConflict
		}
		return domain.EvaluationRun{}, err
	}
	s.deps.Observer.Record(ctx, ports.Event{Name: event, Message: "evaluation run state changed", Fields: map[string]string{"tenant_id": access.TenantID.String(), "run_id": string(snapshot.Input.ID), "state": string(snapshot.State)}})
	return run, nil
}
