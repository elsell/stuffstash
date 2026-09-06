package voice

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (r ProviderProfileResolver) ResolveWorkflowLanguageProvider(ctx context.Context, input ports.WorkflowLanguageProviderResolutionInput) (ports.WorkflowLanguageProviderBinding, error) {
	if r.profiles == nil || r.vault == nil || r.factory == nil || input.ProfileID == "" {
		return ports.WorkflowLanguageProviderBinding{}, ports.ErrInvalidProviderInput
	}
	profile, found, err := r.profiles.ProviderProfileByID(ctx, input.TenantID, agentmodel.ProviderProfileID(input.ProfileID))
	if err != nil {
		return ports.WorkflowLanguageProviderBinding{}, err
	}
	if !found || profile.ID.String() != input.ProfileID || profile.TenantID.String() != input.TenantID.String() || profile.Capability != agentmodel.ProviderCapabilityLanguageInference || !providerProfileRuntimeReady(profile) {
		return ports.WorkflowLanguageProviderBinding{}, ports.ErrInvalidProviderInput
	}
	config, err := r.providerConfig(ctx, input.TenantID, profile)
	if err != nil {
		return ports.WorkflowLanguageProviderBinding{}, err
	}
	provider, err := r.factory.ConversationModelProvider(ctx, config)
	if err != nil {
		return ports.WorkflowLanguageProviderBinding{}, err
	}
	if provider == nil {
		return ports.WorkflowLanguageProviderBinding{}, ports.ErrInvalidProviderInput
	}
	return ports.WorkflowLanguageProviderBinding{ProfileID: input.ProfileID, PromptTemplate: profile.PromptTemplate.String(), Provider: provider}, nil
}
