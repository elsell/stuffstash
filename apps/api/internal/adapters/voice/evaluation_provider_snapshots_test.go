package voice

import (
	"context"
	"testing"

	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	fixture "github.com/stuffstash/stuff-stash/internal/testsupport/evaluationrun"
)

type evaluationSnapshotVault struct {
	ports.ProviderCredentialVault
	reads   int
	version string
}

func (v *evaluationSnapshotVault) ActiveVersionedProviderCredential(context.Context, ports.ProviderCredentialScope) (ports.VersionedProviderCredentialMaterial, bool, error) {
	v.reads++
	return ports.VersionedProviderCredentialMaterial{VersionID: v.version, Raw: []byte("secret")}, true, nil
}

type evaluationSnapshotFactory struct{ providerResolverFactory }

func (evaluationSnapshotFactory) ProviderConfigurationIdentity(_ context.Context, config ProviderProfileProviderConfig) (string, error) {
	return providerConfigurationID(config, "controlled-runtime")
}
func evaluationSnapshotWorkflow(t *testing.T, explicit bool) model.WorkflowRevision {
	t.Helper()
	input := fixture.Run(t, "template").Snapshot().Input.Workflow.Snapshot()
	settings := input.Definition.Settings()
	if explicit {
		for i := range settings.Steps {
			settings.Steps[i].ProviderProfileID = "model"
		}
	}
	definition, err := model.NewWorkflowDefinition(settings, input.Limits)
	if err != nil {
		t.Fatal(err)
	}
	input.Definition = definition
	revision, err := model.NewWorkflowRevision(input)
	if err != nil {
		t.Fatal(err)
	}
	return revision
}
func TestEvaluationProviderSnapshotsReuseVersionWithoutModelCalls(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		profile := providerResolverProfile(t, "model", model.ProviderCapabilityLanguageInference, model.ProviderProfileEnabled, model.CredentialStatusConfigured)
		profile.TenantID = fixture.TenantID
		vault := &evaluationSnapshotVault{version: "credential-one"}
		factory := &evaluationSnapshotFactory{}
		resolver := NewProviderProfileResolver(providerResolverProfileRepository{profiles: []model.ProviderProfile{profile}}, nil, vault, factory)
		workflow := evaluationSnapshotWorkflow(t, explicit)
		pins, err := resolver.SnapshotEvaluationProviders(context.Background(), fixture.TenantID, workflow)
		if err != nil || len(pins) != 2 || vault.reads != 1 {
			t.Fatalf("snapshot: count=%d reads=%d error=%v", len(pins), vault.reads, err)
		}
		if pins[0].ProfileID != "model" || pins[0].ConfigurationID != pins[1].ConfigurationID || pins[0].Step != model.WorkflowStepInterpret || pins[1].Step != model.WorkflowStepAssess {
			t.Fatal("bindings did not match workflow")
		}
		if len(factory.configs) != 0 {
			t.Fatal("snapshot constructed providers")
		}
		vault.version = "credential-two"
		changed, err := resolver.SnapshotEvaluationProviders(context.Background(), fixture.TenantID, workflow)
		if err != nil || changed[0].ConfigurationID == pins[0].ConfigurationID {
			t.Fatal("rotation was not visible")
		}
	}
}
func TestEvaluationProviderSnapshotsFailClosed(t *testing.T) {
	for _, scenario := range []string{"workflow tenant", "profile tenant", "disabled", "wrong capability", "missing profile", "unversioned vault", "unversioned credential", "unsupported factory"} {
		t.Run(scenario, func(t *testing.T) {
			profile := providerResolverProfile(t, "model", model.ProviderCapabilityLanguageInference, model.ProviderProfileEnabled, model.CredentialStatusConfigured)
			profile.TenantID = fixture.TenantID
			versioned := &evaluationSnapshotVault{version: "credential"}
			var vault ports.ProviderCredentialVault = versioned
			var factory ProviderProfileResolverFactory = &evaluationSnapshotFactory{}
			target := tenant.ID(fixture.TenantID)
			switch scenario {
			case "workflow tenant":
				target = "outside"
			case "profile tenant":
				profile.TenantID = "outside"
			case "disabled":
				profile.LifecycleState = model.ProviderProfileDisabled
			case "wrong capability":
				profile.Capability = model.ProviderCapabilitySpeechToText
			case "missing profile":
				profile.ID = "other"
			case "unversioned vault":
				vault = newProviderResolverCredentialVault(profile)
			case "unversioned credential":
				versioned.version = ""
			case "unsupported factory":
				factory = &providerResolverFactory{}
			}
			resolver := NewProviderProfileResolver(providerResolverProfileRepository{profiles: []model.ProviderProfile{profile}}, nil, vault, factory)
			if pins, err := resolver.SnapshotEvaluationProviders(context.Background(), target, evaluationSnapshotWorkflow(t, true)); err == nil || len(pins) != 0 {
				t.Fatal("invalid snapshot accepted")
			}
		})
	}
}

func TestEvaluationProviderSnapshotsHonorConfiguredDefaultWithoutFallback(t *testing.T) {
	first := providerResolverProfile(t, "model", model.ProviderCapabilityLanguageInference, model.ProviderProfileEnabled, model.CredentialStatusConfigured)
	first.TenantID = fixture.TenantID
	selected := first
	selected.ID = "selected"
	profiles := providerResolverProfileRepository{profiles: []model.ProviderProfile{first, selected}}
	configuration := providerResolverVoiceConfigurationRepository{found: true, record: ports.VoiceProviderConfigurationRecord{TenantID: fixture.TenantID, LanguageInferenceProfileID: "selected"}}
	resolver := NewProviderProfileResolver(profiles, configuration, &evaluationSnapshotVault{version: "one"}, &evaluationSnapshotFactory{})
	pins, err := resolver.SnapshotEvaluationProviders(context.Background(), fixture.TenantID, evaluationSnapshotWorkflow(t, false))
	if err != nil || pins[0].ProfileID != "selected" || pins[1].ProfileID != "selected" {
		t.Fatal("configured default ignored")
	}
	configuration.record.LanguageInferenceProfileID = "missing"
	resolver = NewProviderProfileResolver(profiles, configuration, &evaluationSnapshotVault{version: "one"}, &evaluationSnapshotFactory{})
	if _, err := resolver.SnapshotEvaluationProviders(context.Background(), fixture.TenantID, evaluationSnapshotWorkflow(t, false)); err == nil {
		t.Fatal("missing default silently substituted")
	}
	if _, err := resolver.SnapshotEvaluationProviders(context.Background(), fixture.TenantID, evaluationSnapshotWorkflow(t, true)); err != nil {
		t.Fatal("unused default required by explicit workflow")
	}
}
