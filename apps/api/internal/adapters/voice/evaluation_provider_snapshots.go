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
	profileID, explicit := definition.ProviderProfileID, true
	if profileID == "" {
		configuration, configured, err := r.voiceProviderConfiguration(ctx, tenantID)
		if err != nil {
			return nil, err
		}
		profileID, explicit = configuration.LanguageInferenceProfileID, configured
	}
	profile, found := r.selectConfiguredProviderProfile(scoped, profileID, explicit, model.ProviderCapabilityLanguageInference)
	if !found {
		return nil, ports.ErrInvalidProviderInput
	}
	config, err := r.providerConfig(ctx, tenantID, profile)
	if err != nil {
		return nil, err
	}
	configurationID, err := identityFactory.ProviderConfigurationIdentity(ctx, config)
	if err != nil {
		return nil, err
	}
	if !validProviderConfigurationID(configurationID) {
		return nil, ports.ErrInvalidProviderInput
	}
	return []model.EvaluationRunProvider{{ProfileID: profile.ID, ConfigurationID: configurationID}}, nil
}
