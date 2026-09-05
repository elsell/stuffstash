package agentmodel

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type WorkflowActivationDependencies struct {
	Authorizer ports.Authorizer
	Workflows  ports.ConversationWorkflowRepository
	Runs       ports.EvaluationRunRepository
	Providers  ports.EvaluationProviderSnapshotResolver
	IDs        ports.IDGenerator
	Clock      ports.Clock
	Observer   ports.Observer
	Limits     model.WorkflowLimits
}
type WorkflowActivationService struct {
	deps WorkflowActivationDependencies
}

func NewWorkflowActivationService(deps WorkflowActivationDependencies) WorkflowActivationService {
	if deps.Observer == nil {
		deps.Observer = noopObserver{}
	}
	return WorkflowActivationService{deps: deps}
}

type ActivateWorkflowInput struct {
	EvaluationRunAccess
	WorkflowID model.WorkflowID
	RevisionID model.WorkflowRevisionID
	RunID      model.EvaluationRunID
	Cases      []EvaluationRunCaseReference
	Expected   ports.WorkflowSelectionReference
}

func (s WorkflowActivationService) Activate(ctx context.Context, input ActivateWorkflowInput) (model.WorkflowRevision, error) {
	if err := authorizeEvaluationRunAccess(ctx, s.deps.Authorizer, input.EvaluationRunAccess); err != nil {
		return model.WorkflowRevision{}, err
	}
	if s.deps.Workflows == nil || s.deps.Runs == nil || s.deps.Providers == nil || s.deps.IDs == nil || s.deps.Clock == nil {
		return model.WorkflowRevision{}, apperrors.ErrPrecondition
	}
	if input.WorkflowID == "" || input.RevisionID == "" || input.RunID == "" || len(input.Cases) == 0 || len(input.Cases) > model.MaxEvaluationRunCases || (input.Expected.WorkflowID == "") != (input.Expected.RevisionID == "") {
		return model.WorkflowRevision{}, apperrors.ErrValidation
	}
	revision, found, err := s.deps.Workflows.WorkflowRevision(ctx, input.TenantID, input.WorkflowID, input.RevisionID)
	if err != nil {
		return model.WorkflowRevision{}, err
	}
	if !found {
		return model.WorkflowRevision{}, apperrors.ErrNotFound
	}
	run, found, err := s.deps.Runs.EvaluationRun(ctx, input.TenantID, input.RunID)
	if err != nil {
		return model.WorkflowRevision{}, err
	}
	if !found {
		return model.WorkflowRevision{}, apperrors.ErrNotFound
	}
	providers, err := s.deps.Providers.SnapshotEvaluationProviders(ctx, input.TenantID, revision)
	if err != nil {
		return model.WorkflowRevision{}, err
	}
	cases := make([]model.EvaluationCasePin, 0, len(input.Cases))
	for _, pin := range input.Cases {
		cases = append(cases, model.EvaluationCasePin{CaseID: pin.CaseID, RevisionID: pin.RevisionID})
	}
	if err := run.ValidateActivation(model.WorkflowActivationCandidate{Workflow: revision, Limits: s.deps.Limits, Cases: cases, Providers: providers}); err != nil {
		return model.WorkflowRevision{}, apperrors.ErrPrecondition
	}
	now := s.deps.Clock.Now()
	record, ok := audit.NewRecord(audit.ID(s.deps.IDs.NewID()), audit.TenantID(input.TenantID), "", audit.PrincipalID(input.Principal.ID), audit.ActionConversationWorkflowActivated, input.Source, audit.TargetConversationWorkflow, string(input.WorkflowID), now, input.RequestID, map[string]string{"revision_id": string(input.RevisionID), "evaluation_run_id": string(input.RunID)})
	if !ok {
		return model.WorkflowRevision{}, apperrors.ErrValidation
	}
	if err := s.deps.Workflows.ActivateWorkflowRevision(ctx, input.TenantID, input.WorkflowID, input.RevisionID, input.Expected, now, record); err != nil {
		if errors.Is(err, ports.ErrWorkflowConflict) {
			return model.WorkflowRevision{}, apperrors.ErrConflict
		}
		if errors.Is(err, ports.ErrWorkflowNotFound) {
			return model.WorkflowRevision{}, apperrors.ErrNotFound
		}
		return model.WorkflowRevision{}, err
	}
	s.deps.Observer.Record(ctx, ports.Event{Name: ports.EventConversationWorkflowActivated, Message: "conversation workflow activated", Fields: map[string]string{"tenant_id": input.TenantID.String(), "workflow_id": string(input.WorkflowID), "revision_id": string(input.RevisionID), "evaluation_run_id": string(input.RunID)}})
	return revision, nil
}
