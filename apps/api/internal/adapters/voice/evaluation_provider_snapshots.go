package voice

import (
	"context"

	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (r ProviderProfileResolver) SnapshotEvaluationProviders(ctx context.Context, tenantID tenant.ID, revision model.WorkflowRevision) ([]model.EvaluationRunProvider, error) {
	identityFactory, supported := r.factory.(ProviderConfigurationIdentityFactory)
	_, versioned := r.vault.(ports.VersionedProviderCredentialVault)
	if !supported || !versioned || r.profiles == nil || revision.Snapshot().TenantID != model.TenantID(tenantID) {
		return nil, ports.ErrInvalidProviderInput
	}
	if _, err := model.NewWorkflowRevision(revision.Snapshot()); err != nil {
		return nil, ports.ErrInvalidProviderInput
	}
	profiles, err := r.profiles.ListProviderProfiles(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	scoped := make([]model.ProviderProfile, 0, len(profiles))
	for _, profile := range profiles {
		if profile.TenantID == model.TenantID(tenantID) {
			scoped = append(scoped, profile)
		}
	}
	definition := revision.Snapshot().Definition.Settings()
	var defaultProfile model.ProviderProfile
	defaultResolved := false
	identities := map[model.ProviderProfileID]string{}
	bindings := make([]model.EvaluationRunProvider, 0, len(definition.Steps))
	for _, step := range definition.Steps {
		if step.Kind == model.WorkflowStepRespond && definition.Response == model.WorkflowResponseGrounded {
			continue
		}
		var profile model.ProviderProfile
		var found bool
		if step.ProviderProfileID != "" {
			profile, found = r.selectConfiguredProviderProfile(scoped, step.ProviderProfileID, true, model.ProviderCapabilityLanguageInference)
		} else {
			if !defaultResolved {
				configuration, explicit, err := r.voiceProviderConfiguration(ctx, tenantID)
				if err != nil {
					return nil, err
				}
				defaultProfile, found = r.selectConfiguredProviderProfile(scoped, configuration.LanguageInferenceProfileID, explicit, model.ProviderCapabilityLanguageInference)
				if !found {
					return nil, ports.ErrInvalidProviderInput
				}
				defaultResolved = true
			}
			profile, found = defaultProfile, true
		}
		if !found {
			return nil, ports.ErrInvalidProviderInput
		}
		configurationID, exists := identities[profile.ID]
		if !exists {
			config, err := r.providerConfig(ctx, tenantID, profile)
			if err != nil {
				return nil, err
			}
			configurationID, err = identityFactory.ProviderConfigurationIdentity(ctx, config)
			if err != nil {
				return nil, err
			}
			if !validProviderConfigurationID(configurationID) {
				return nil, ports.ErrInvalidProviderInput
			}
			identities[profile.ID] = configurationID
		}
		bindings = append(bindings, model.EvaluationRunProvider{Step: step.Kind, ProfileID: profile.ID, ConfigurationID: configurationID})
	}
	return bindings, nil
}
