package agentmodel

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type ConversationWorkflowDependencies struct {
	Authorizer ports.Authorizer
	Repository ports.ConversationWorkflowRepository
	Profiles   ports.ProviderProfileRepository
	IDs        ports.IDGenerator
	Clock      ports.Clock
	Observer   ports.Observer
	Limits     domain.WorkflowLimits
}

type ConversationWorkflowService struct {
	deps ConversationWorkflowDependencies
}

func NewConversationWorkflowService(deps ConversationWorkflowDependencies) ConversationWorkflowService {
	if deps.Observer == nil {
		deps.Observer = noopObserver{}
	}
	return ConversationWorkflowService{deps: deps}
}

type SaveConversationWorkflowInput struct {
	Principal        identity.Principal
	TenantID         tenant.ID
	WorkflowID       domain.WorkflowID
	ExpectedRevision int
	Definition       domain.WorkflowDefinitionInput
	Source           audit.Source
	RequestID        string
}

func (s ConversationWorkflowService) SaveRevision(ctx context.Context, input SaveConversationWorkflowInput) (domain.WorkflowRevision, error) {
	if input.Principal.ID.String() == "" {
		return domain.WorkflowRevision{}, apperrors.ErrUnauthenticated
	}
	if s.deps.Authorizer == nil {
		return domain.WorkflowRevision{}, apperrors.ErrPrecondition
	}
	if err := s.deps.Authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		return domain.WorkflowRevision{}, err
	}
	if s.deps.Repository == nil || s.deps.Profiles == nil || s.deps.IDs == nil || s.deps.Clock == nil {
		return domain.WorkflowRevision{}, apperrors.ErrPrecondition
	}
	if strings.TrimSpace(input.TenantID.String()) == "" || input.ExpectedRevision < 0 || (input.WorkflowID == "" && input.ExpectedRevision != 0) {
		return domain.WorkflowRevision{}, apperrors.ErrValidation
	}
	definition, err := domain.NewWorkflowDefinition(input.Definition, s.deps.Limits)
	if err != nil {
		return domain.WorkflowRevision{}, apperrors.ErrValidation
	}
	if err := s.validateWorkflowProviders(ctx, input.TenantID, definition); err != nil {
		return domain.WorkflowRevision{}, err
	}
	workflowID := input.WorkflowID
	if workflowID == "" {
		workflowID = domain.WorkflowID(s.deps.IDs.NewID())
	} else {
		head, found, err := s.deps.Repository.WorkflowHead(ctx, input.TenantID, workflowID)
		if err != nil {
			return domain.WorkflowRevision{}, err
		}
		if !found {
			return domain.WorkflowRevision{}, apperrors.ErrNotFound
		}
		if head.LatestRevision != input.ExpectedRevision {
			return domain.WorkflowRevision{}, apperrors.ErrConflict
		}
	}
	revision, err := domain.NewWorkflowRevision(domain.WorkflowRevisionInput{ID: domain.WorkflowRevisionID(s.deps.IDs.NewID()), WorkflowID: workflowID, TenantID: domain.TenantID(input.TenantID.String()), AuthorID: domain.WorkflowAuthorID(input.Principal.ID.String()), Number: input.ExpectedRevision + 1, Definition: definition, Limits: s.deps.Limits, CreatedAt: s.deps.Clock.Now()})
	if err != nil {
		return domain.WorkflowRevision{}, apperrors.ErrValidation
	}
	snapshot := revision.Snapshot()
	record, ok := audit.NewRecord(audit.ID(s.deps.IDs.NewID()), audit.TenantID(input.TenantID.String()), "", audit.PrincipalID(input.Principal.ID.String()), audit.ActionConversationWorkflowRevisionCreated, input.Source, audit.TargetType("conversation_workflow"), string(workflowID), snapshot.CreatedAt, input.RequestID, map[string]string{"workflow_id": string(workflowID), "revision_id": string(snapshot.ID), "revision_number": strconv.Itoa(snapshot.Number)})
	if !ok {
		return domain.WorkflowRevision{}, apperrors.ErrValidation
	}
	if err := s.deps.Repository.AppendWorkflowRevision(ctx, revision, input.ExpectedRevision, record); err != nil {
		if errors.Is(err, ports.ErrWorkflowConflict) {
			return domain.WorkflowRevision{}, apperrors.ErrConflict
		}
		return domain.WorkflowRevision{}, err
	}
	s.deps.Observer.Record(ctx, ports.Event{Name: ports.EventConversationWorkflowRevisionCreated, Message: "conversation workflow revision created", Fields: map[string]string{"tenant_id": input.TenantID.String(), "workflow_id": string(workflowID), "revision_id": string(snapshot.ID)}})
	return revision, nil
}

func (s ConversationWorkflowService) validateWorkflowProviders(ctx context.Context, tenantID tenant.ID, definition domain.WorkflowDefinition) error {
	checked := map[domain.ProviderProfileID]struct{}{}
	for _, step := range definition.Settings().Steps {
		if step.ProviderProfileID == "" {
			continue
		}
		id := domain.ProviderProfileID(step.ProviderProfileID)
		if _, ok := checked[id]; ok {
			continue
		}
		profile, found, err := s.deps.Profiles.ProviderProfileByID(ctx, tenantID, id)
		if err != nil {
			return err
		}
		if !found || profile.TenantID.String() != tenantID.String() || profile.Capability != domain.ProviderCapabilityLanguageInference || profile.LifecycleState == domain.ProviderProfileArchived {
			return apperrors.ErrValidation
		}
		checked[id] = struct{}{}
	}
	return nil
}
