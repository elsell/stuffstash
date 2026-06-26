package app

import (
	"context"

	agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type CreateProviderProfileInput = agentmodelapp.CreateProviderProfileInput
type ListProviderProfilesInput = agentmodelapp.ListProviderProfilesInput
type GetProviderProfileInput = agentmodelapp.GetProviderProfileInput
type UpdateProviderProfileInput = agentmodelapp.UpdateProviderProfileInput
type ProviderProfileLifecycleInput = agentmodelapp.ProviderProfileLifecycleInput
type ReplaceProviderProfileCredentialInput = agentmodelapp.ReplaceProviderProfileCredentialInput
type TestProviderProfileInput = agentmodelapp.TestProviderProfileInput

func (a App) CreateProviderProfile(ctx context.Context, input CreateProviderProfileInput) (agentmodel.ProviderProfile, error) {
	return a.providerProfileService.CreateProviderProfile(ctx, input)
}

func (a App) ListProviderProfiles(ctx context.Context, input ListProviderProfilesInput) ([]agentmodel.ProviderProfile, error) {
	return a.providerProfileService.ListProviderProfiles(ctx, input)
}

func (a App) GetProviderProfile(ctx context.Context, input GetProviderProfileInput) (agentmodel.ProviderProfile, error) {
	return a.providerProfileService.GetProviderProfile(ctx, input)
}

func (a App) UpdateProviderProfile(ctx context.Context, input UpdateProviderProfileInput) (agentmodel.ProviderProfile, error) {
	return a.providerProfileService.UpdateProviderProfile(ctx, input)
}

func (a App) EnableProviderProfile(ctx context.Context, input ProviderProfileLifecycleInput) (agentmodel.ProviderProfile, error) {
	return a.providerProfileService.EnableProviderProfile(ctx, input)
}

func (a App) DisableProviderProfile(ctx context.Context, input ProviderProfileLifecycleInput) (agentmodel.ProviderProfile, error) {
	return a.providerProfileService.DisableProviderProfile(ctx, input)
}

func (a App) ArchiveProviderProfile(ctx context.Context, input ProviderProfileLifecycleInput) (agentmodel.ProviderProfile, error) {
	return a.providerProfileService.ArchiveProviderProfile(ctx, input)
}

func (a App) ReplaceProviderProfileCredential(ctx context.Context, input ReplaceProviderProfileCredentialInput) (agentmodel.ProviderProfile, error) {
	return a.providerProfileService.ReplaceProviderProfileCredential(ctx, input)
}

func (a App) TestProviderProfile(ctx context.Context, input TestProviderProfileInput) (ports.ProviderProfileTestResult, error) {
	return a.providerProfileService.TestProviderProfile(ctx, input)
}
