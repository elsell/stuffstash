package agentmodel

import (
	"context"

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
// captured. The caller applies its budgets separately to each user turn.
type PreparedWorkflow struct {
	conversation          *workflowConversationModel
	conversationProfileID string
	revision              domain.WorkflowRevision
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
	if err != nil {
		return nil, apperrors.ErrPrecondition
	}
	return &SelectedWorkflow{revision: revision}, nil
}

type SelectedWorkflow struct {
	revision domain.WorkflowRevision
}

func (selected *SelectedWorkflow) NeedsDefaultLanguage() bool {
	settings := selected.revision.Snapshot().Definition.Settings()
	return settings.ProviderProfileID == ""
}

func (s ConversationWorkflowService) PrepareSelected(ctx context.Context, input PrepareWorkflowInput) (*PreparedWorkflow, error) {
	selected, err := s.Selected(ctx, input.Principal, input.TenantID)
	if err != nil || selected == nil {
		return nil, err
	}
	return selected.Prepare(ctx, input.DefaultProviders, input.Resolver)
}
func (selected *SelectedWorkflow) Prepare(ctx context.Context, defaults ports.RealtimeVoiceProviderSet, resolver ports.WorkflowLanguageProviderResolver) (*PreparedWorkflow, error) {
	settings := selected.revision.Snapshot().Definition.Settings()
	model := defaults.ConversationModel
	profileID, prompt := defaults.LanguageInferenceProfileID, defaults.LanguagePromptTemplate
	if settings.ProviderProfileID != "" && settings.ProviderProfileID != profileID {
		if resolver == nil {
			return nil, apperrors.ErrPrecondition
		}
		binding, err := resolver.ResolveWorkflowLanguageProvider(ctx, ports.WorkflowLanguageProviderResolutionInput{TenantID: tenant.ID(selected.revision.Snapshot().TenantID), ProfileID: settings.ProviderProfileID})
		if err != nil {
			return nil, err
		}
		if binding.ProfileID != settings.ProviderProfileID || binding.Provider == nil {
			return nil, ports.ErrInvalidProviderInput
		}
		model = binding.Provider
		profileID, prompt = binding.ProfileID, binding.PromptTemplate
	}
	if model == nil {
		return nil, ports.ErrInvalidProviderInput
	}
	conversation, err := newWorkflowConversationModel(model, settings, prompt)
	if err != nil {
		return nil, err
	}
	return &PreparedWorkflow{revision: selected.revision, conversation: conversation, conversationProfileID: profileID}, nil
}

func (p *PreparedWorkflow) Revision() domain.WorkflowRevision { return p.revision }
