package agentmodel

import (
	"context"
	"sync"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type PrepareWorkflowInput struct {
	Principal        identity.Principal
	TenantID         tenant.ID
	DefaultProviders ports.RealtimeVoiceProviderSet
	Resolver         ports.WorkflowLanguageProviderResolver
}

// PreparedWorkflow pins immutable configuration and providers before audio is
// captured. Its shared model budget starts with the first model invocation.
type PreparedWorkflow struct {
	revision  domain.WorkflowRevision
	limits    domain.WorkflowLimits
	clock     ports.Clock
	bindings  map[domain.WorkflowStepKind]WorkflowModelBinding
	once      sync.Once
	execution *WorkflowModelExecution
	err       error
}

func (s ConversationWorkflowService) Selected(ctx context.Context, principal identity.Principal, tenantID tenant.ID) (*SelectedWorkflow, error) {
	if principal.ID.String() == "" {
		return nil, apperrors.ErrUnauthenticated
	}
	if s.deps.Authorizer == nil {
		return nil, apperrors.ErrPrecondition
	}
	if err := s.deps.Authorizer.CheckTenant(ctx, principal, ports.TenantPermissionView, tenantID); err != nil {
		return nil, err
	}
	// Older embedders may not provide a workflow repository; no customization is active.
	if s.deps.Repository == nil {
		return nil, nil
	}
	selected, found, err := s.deps.Repository.SelectedWorkflowRevision(ctx, tenantID)
	if err != nil || !found {
		return nil, err
	}
	revision, found, err := s.deps.Repository.WorkflowRevision(ctx, tenantID, selected.WorkflowID, selected.RevisionID)
	if err != nil {
		return nil, err
	}
	snapshot := revision.Snapshot()
	if !found || snapshot.TenantID != domain.TenantID(tenantID.String()) || snapshot.WorkflowID != selected.WorkflowID || snapshot.ID != selected.RevisionID {
		return nil, apperrors.ErrPrecondition
	}
	_, err = domain.NewWorkflowDefinition(snapshot.Definition.Settings(), s.deps.Limits)
	if err != nil || s.deps.Clock == nil {
		return nil, apperrors.ErrPrecondition
	}
	return &SelectedWorkflow{revision: revision, limits: s.deps.Limits, clock: s.deps.Clock}, nil
}

type SelectedWorkflow struct {
	revision domain.WorkflowRevision
	limits   domain.WorkflowLimits
	clock    ports.Clock
}

func (selected *SelectedWorkflow) NeedsDefaultLanguage() bool {
	settings := selected.revision.Snapshot().Definition.Settings()
	for _, step := range settings.Steps {
		if step.Kind == domain.WorkflowStepRespond && settings.Response == domain.WorkflowResponseGrounded {
			continue
		}
		if step.ProviderProfileID == "" {
			return true
		}
	}
	return false
}
func (s ConversationWorkflowService) PrepareSelected(ctx context.Context, input PrepareWorkflowInput) (*PreparedWorkflow, error) {
	selected, err := s.Selected(ctx, input.Principal, input.TenantID)
	if err != nil || selected == nil {
		return nil, err
	}
	return selected.Prepare(ctx, input.DefaultProviders, input.Resolver)
}
func (selected *SelectedWorkflow) Prepare(ctx context.Context, defaults ports.RealtimeVoiceProviderSet, resolver ports.WorkflowLanguageProviderResolver) (*PreparedWorkflow, error) {
	definition := selected.revision.Snapshot().Definition
	tenantID := tenant.ID(selected.revision.Snapshot().TenantID)
	cache := map[string]WorkflowModelBinding{}
	if defaults.LanguageInferenceProfileID != "" {
		cache[defaults.LanguageInferenceProfileID] = WorkflowModelBinding{ProfileID: defaults.LanguageInferenceProfileID, PromptTemplate: defaults.LanguagePromptTemplate, Language: defaults.LanguageInference, Response: defaults.ResponseGenerator}
	}
	bindings := map[domain.WorkflowStepKind]WorkflowModelBinding{}
	for _, step := range definition.Settings().Steps {
		if step.Kind == domain.WorkflowStepRespond && definition.Settings().Response == domain.WorkflowResponseGrounded {
			bindings[step.Kind] = WorkflowModelBinding{ProfileID: step.ProviderProfileID}
			continue
		}
		binding := WorkflowModelBinding{ProfileID: defaults.LanguageInferenceProfileID, PromptTemplate: defaults.LanguagePromptTemplate, Language: defaults.LanguageInference, Response: defaults.ResponseGenerator}
		if step.ProviderProfileID != "" {
			var exists bool
			binding, exists = cache[step.ProviderProfileID]
			if !exists {
				if resolver == nil {
					return nil, apperrors.ErrPrecondition
				}
				resolved, resolveErr := resolver.ResolveWorkflowLanguageProvider(ctx, ports.WorkflowLanguageProviderResolutionInput{TenantID: tenantID, ProfileID: step.ProviderProfileID})
				if resolveErr != nil {
					return nil, resolveErr
				}
				if resolved.ProfileID != step.ProviderProfileID || resolved.Provider == nil {
					return nil, ports.ErrInvalidProviderInput
				}
				binding = WorkflowModelBinding{ProfileID: resolved.ProfileID, PromptTemplate: resolved.PromptTemplate, Language: resolved.Provider, Response: resolved.Provider}
				cache[step.ProviderProfileID] = binding
			}
		}
		bindings[step.Kind] = binding
	}
	// Validate dependencies now, without starting the execution retained by the session.
	if _, err := NewWorkflowModelExecution(definition, selected.limits, selected.clock, bindings); err != nil {
		return nil, err
	}
	return &PreparedWorkflow{revision: selected.revision, limits: selected.limits, clock: selected.clock, bindings: bindings}, nil
}

func (p *PreparedWorkflow) Revision() domain.WorkflowRevision { return p.revision }
func (p *PreparedWorkflow) modelExecution() (*WorkflowModelExecution, error) {
	p.once.Do(func() {
		p.execution, p.err = NewWorkflowModelExecution(p.revision.Snapshot().Definition, p.limits, p.clock, p.bindings)
	})
	return p.execution, p.err
}
func (p *PreparedWorkflow) NextTurn(ctx context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	if err := ctx.Err(); err != nil {
		return ports.LanguageInferenceTurn{}, err
	}
	execution, err := p.modelExecution()
	if err != nil {
		return ports.LanguageInferenceTurn{}, err
	}
	input.PromptTemplate = ""
	return execution.NextTurn(ctx, input)
}
func (p *PreparedWorkflow) GenerateResponse(ctx context.Context, input ports.VoiceResponseGenerationInput) (ports.VoiceResponseGenerationResult, error) {
	if err := ctx.Err(); err != nil {
		return ports.VoiceResponseGenerationResult{}, err
	}
	execution, err := p.modelExecution()
	if err != nil {
		return ports.VoiceResponseGenerationResult{}, err
	}
	return execution.GenerateResponse(ctx, input)
}
