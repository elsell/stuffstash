package voice

import (
	"context"
	"errors"

	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (r ProviderProfileResolver) ResolveEvaluationRunProviders(ctx context.Context, tenantID tenant.ID, run model.EvaluationRun) (ports.EvaluationExecutionProviders, error) {
	if err := ctx.Err(); err != nil {
		return ports.EvaluationExecutionProviders{}, err
	}
	snapshot := run.Snapshot()
	if snapshot.Input.TenantID != model.TenantID(tenantID) {
		return ports.EvaluationExecutionProviders{}, ports.ErrEvaluationConfigurationChanged
	}
	if _, err := model.RestoreEvaluationRun(snapshot); err != nil {
		return ports.EvaluationExecutionProviders{}, ports.ErrEvaluationConfigurationChanged
	}
	identityFactory, supported := r.factory.(ProviderConfigurationIdentityFactory)
	_, versioned := r.vault.(ports.VersionedProviderCredentialVault)
	if !supported || !versioned || r.profiles == nil {
		return ports.EvaluationExecutionProviders{}, ports.ErrEvaluationConfigurationChanged
	}
	configurations := map[model.ProviderProfileID]ProviderProfileProviderConfig{}
	order := []model.ProviderProfileID{}
	for _, pin := range snapshot.Input.Providers {
		if _, exists := configurations[pin.ProfileID]; exists {
			continue
		}
		profile, found, err := r.profiles.ProviderProfileByID(ctx, tenantID, pin.ProfileID)
		if err != nil {
			return ports.EvaluationExecutionProviders{}, err
		}
		if !found || profile.ID != pin.ProfileID || profile.TenantID != model.TenantID(tenantID) || profile.Capability != model.ProviderCapabilityLanguageInference || !providerProfileRuntimeReady(profile) {
			return ports.EvaluationExecutionProviders{}, ports.ErrEvaluationConfigurationChanged
		}
		config, err := r.providerConfig(ctx, tenantID, profile)
		if err != nil {
			return ports.EvaluationExecutionProviders{}, evaluationResolutionError(err)
		}
		identity, err := identityFactory.ProviderConfigurationIdentity(ctx, config)
		if err != nil {
			return ports.EvaluationExecutionProviders{}, evaluationResolutionError(err)
		}
		if identity != pin.ConfigurationID {
			return ports.EvaluationExecutionProviders{}, ports.ErrEvaluationConfigurationChanged
		}
		configurations[pin.ProfileID] = config
		order = append(order, pin.ProfileID)
	}
	// All pins have passed before any factory can construct a provider.
	bindings := make(map[string]ports.WorkflowLanguageProviderBinding, len(configurations))
	for _, id := range order {
		if err := ctx.Err(); err != nil {
			return ports.EvaluationExecutionProviders{}, err
		}
		config := configurations[id]
		provider, err := r.factory.RealtimeLanguageProvider(ctx, config)
		if err != nil {
			return ports.EvaluationExecutionProviders{}, evaluationResolutionError(err)
		}
		if provider == nil {
			return ports.EvaluationExecutionProviders{}, ports.ErrEvaluationConfigurationChanged
		}
		bindings[id.String()] = ports.WorkflowLanguageProviderBinding{ProfileID: id.String(), PromptTemplate: config.Profile.PromptTemplate.String(), Provider: provider}
	}
	result := ports.EvaluationExecutionProviders{WorkflowProviders: pinnedEvaluationProviderResolver{tenantID: tenantID, bindings: bindings}}
	byStep := map[model.WorkflowStepKind]model.ProviderProfileID{}
	for _, pin := range snapshot.Input.Providers {
		byStep[pin.Step] = pin.ProfileID
	}
	for _, step := range snapshot.Input.Workflow.Snapshot().Definition.Settings().Steps {
		if step.ProviderProfileID != "" {
			continue
		}
		id, used := byStep[step.Kind]
		if !used {
			continue
		}
		binding := bindings[id.String()]
		result.Providers.LanguageInferenceProfileID = binding.ProfileID
		result.Providers.LanguagePromptTemplate = binding.PromptTemplate
		result.Providers.LanguageInference = binding.Provider
		result.Providers.ResponseGenerator = binding.Provider
		break
	}
	return result, nil
}
func evaluationResolutionError(err error) error {
	if errors.Is(err, ports.ErrInvalidProviderInput) {
		return ports.ErrEvaluationConfigurationChanged
	}
	return err
}

type pinnedEvaluationProviderResolver struct {
	tenantID tenant.ID
	bindings map[string]ports.WorkflowLanguageProviderBinding
}

func (r pinnedEvaluationProviderResolver) ResolveWorkflowLanguageProvider(ctx context.Context, input ports.WorkflowLanguageProviderResolutionInput) (ports.WorkflowLanguageProviderBinding, error) {
	if err := ctx.Err(); err != nil {
		return ports.WorkflowLanguageProviderBinding{}, err
	}
	if input.TenantID != r.tenantID {
		return ports.WorkflowLanguageProviderBinding{}, ports.ErrEvaluationConfigurationChanged
	}
	binding, found := r.bindings[input.ProfileID]
	if !found {
		return ports.WorkflowLanguageProviderBinding{}, ports.ErrEvaluationConfigurationChanged
	}
	return binding, nil
}
