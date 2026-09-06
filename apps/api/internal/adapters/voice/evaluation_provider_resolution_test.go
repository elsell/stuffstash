package voice

import (
	"context"
	"errors"
	"testing"

	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
	fixture "github.com/stuffstash/stuff-stash/internal/testsupport/evaluationrun"
)

func evaluationResolutionSetup(t *testing.T, explicit bool) (ProviderProfileResolver, model.EvaluationRun, *evaluationSnapshotVault, *evaluationSnapshotFactory) {
	t.Helper()
	profile := providerResolverProfile(t, "model", model.ProviderCapabilityLanguageInference, model.ProviderProfileEnabled, model.CredentialStatusConfigured)
	profile.TenantID = fixture.TenantID
	vault := &evaluationSnapshotVault{version: "one"}
	factory := &evaluationSnapshotFactory{}
	resolver := NewProviderProfileResolver(providerResolverProfileRepository{profiles: []model.ProviderProfile{profile}}, nil, vault, factory)
	input := fixture.Run(t, "run").Snapshot().Input
	input.Workflow = evaluationSnapshotWorkflow(t, explicit)
	pins, err := resolver.SnapshotEvaluationProviders(context.Background(), fixture.TenantID, input.Workflow)
	if err != nil {
		t.Fatal(err)
	}
	input.Providers = pins
	run, err := model.NewEvaluationRun(input)
	if err != nil {
		t.Fatal(err)
	}
	return resolver, run, vault, factory
}
func TestEvaluationProviderResolutionUsesPinnedInstances(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		resolver, run, vault, factory := evaluationResolutionSetup(t, explicit)
		before := vault.reads
		bundle, err := resolver.ResolveEvaluationRunProviders(context.Background(), fixture.TenantID, run)
		if err != nil || vault.reads != before+1 || len(factory.configs) != 1 {
			t.Fatalf("provider resolution: %v", err)
		}
		binding, err := bundle.WorkflowProviders.ResolveWorkflowLanguageProvider(context.Background(), ports.WorkflowLanguageProviderResolutionInput{TenantID: fixture.TenantID, ProfileID: "model"})
		if err != nil || binding.Provider == nil {
			t.Fatal("pinned provider unavailable")
		}
		if !explicit && (bundle.Providers.ConversationModel == nil || bundle.Providers.LanguageInferenceProfileID != "model") {
			t.Fatal("pinned default unavailable")
		}
		if explicit && bundle.Providers.ConversationModel != nil {
			t.Fatal("unused default constructed")
		}
		for _, input := range []ports.WorkflowLanguageProviderResolutionInput{{TenantID: "outside", ProfileID: "model"}, {TenantID: fixture.TenantID, ProfileID: "other"}} {
			if _, err := bundle.WorkflowProviders.ResolveWorkflowLanguageProvider(context.Background(), input); err == nil {
				t.Fatal("pinned resolver widened scope")
			}
		}
	}
}
func TestEvaluationProviderResolutionRejectsDriftBeforeConstruction(t *testing.T) {
	resolver, run, vault, factory := evaluationResolutionSetup(t, true)
	vault.version = "rotated"
	if _, err := resolver.ResolveEvaluationRunProviders(context.Background(), fixture.TenantID, run); !errors.Is(err, ports.ErrEvaluationConfigurationChanged) {
		t.Fatalf("credential drift: %v", err)
	}
	if len(factory.configs) != 0 {
		t.Fatal("provider constructed before drift check")
	}
	if _, err := resolver.ResolveEvaluationRunProviders(context.Background(), "outside", run); !errors.Is(err, ports.ErrEvaluationConfigurationChanged) {
		t.Fatal("cross-tenant run accepted")
	}
}
func TestEvaluationProviderResolutionDoesNotReselectDefault(t *testing.T) {
	resolver, run, _, _ := evaluationResolutionSetup(t, false)
	resolver.voiceConfigs = providerResolverVoiceConfigurationRepository{found: true, record: ports.VoiceProviderConfigurationRecord{TenantID: fixture.TenantID, LanguageInferenceProfileID: "different-default"}}
	bundle, err := resolver.ResolveEvaluationRunProviders(context.Background(), fixture.TenantID, run)
	if err != nil || bundle.Providers.LanguageInferenceProfileID != "model" {
		t.Fatal("default change redirected pinned run")
	}
}

func TestEvaluationProviderResolutionChecksSelectedModelBeforeConstruction(t *testing.T) {
	resolver, run, _, factory := evaluationResolutionSetup(t, true)
	repository := resolver.profiles.(providerResolverProfileRepository)
	second := repository.profiles[0]
	second.ID = "second"
	repository.profiles = append(repository.profiles, second)
	resolver.profiles = repository
	input := run.Snapshot().Input
	revision := input.Workflow.Snapshot()
	settings := revision.Definition.Settings()
	settings.ProviderProfileID = "second"
	definition, err := model.NewWorkflowDefinition(settings, revision.Limits)
	if err != nil {
		t.Fatal(err)
	}
	revision.Definition = definition
	input.Workflow, err = model.NewWorkflowRevision(revision)
	if err != nil {
		t.Fatal(err)
	}
	input.Providers, err = resolver.SnapshotEvaluationProviders(context.Background(), fixture.TenantID, input.Workflow)
	if err != nil {
		t.Fatal(err)
	}
	run, err = model.NewEvaluationRun(input)
	if err != nil {
		t.Fatal(err)
	}
	repository.profiles[1].ModelName = "changed-model"
	resolver.profiles = repository
	if _, err := resolver.ResolveEvaluationRunProviders(context.Background(), fixture.TenantID, run); !errors.Is(err, ports.ErrEvaluationConfigurationChanged) {
		t.Fatalf("selected profile drift: %v", err)
	}
	if len(factory.configs) != 0 {
		t.Fatal("provider constructed before model drift check")
	}
}
