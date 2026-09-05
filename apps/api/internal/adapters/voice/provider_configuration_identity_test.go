package voice

import (
	"context"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func identityConfig(t *testing.T) ProviderProfileProviderConfig {
	profile := providerResolverProfile(t, "model", agentmodel.ProviderCapabilityLanguageInference, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured)
	profile.RuntimeOptionsJSON = agentmodel.JSONObject(`{"a":1,"nested":{"x":true,"y":"value"}}`)
	return ProviderProfileProviderConfig{Profile: profile, CredentialPurpose: ports.ProviderCredentialPurposeAPIKey, CredentialVersionID: "credential-one", Credential: []byte("secret")}
}
func TestProviderConfigurationIdentityCanonicalAndSecretIndependent(t *testing.T) {
	config := identityConfig(t)
	first, err := providerConfigurationID(config, "runtime-v1")
	if err != nil || len(first) != 64 {
		t.Fatal("missing fingerprint")
	}
	config.Credential = []byte("not-part-of-fingerprint")
	config.Profile.RuntimeOptionsJSON = agentmodel.JSONObject(`{"nested":{"y":"value","x":true},"a":1}`)
	second, err := providerConfigurationID(config, "runtime-v1")
	if err != nil || first != second {
		t.Fatal("JSON ordering or secret bytes changed identity")
	}
	config.CredentialVersionID = "credential-two"
	changed, err := providerConfigurationID(config, "runtime-v1")
	if err != nil || changed == first {
		t.Fatal("credential rotation did not change identity")
	}
	for _, invalid := range []string{"", " "} {
		if _, err := providerConfigurationID(config, invalid); err == nil {
			t.Fatal("empty runtime identity accepted")
		}
	}
}
func TestProviderConfigurationIdentityCoversEffectiveInputs(t *testing.T) {
	baseline := identityConfig(t)
	original, _ := providerConfigurationID(baseline, "runtime")
	for _, field := range []string{"tenant", "profile", "model", "endpoint", "prompt", "runtime", "capability options", "credential purpose", "adapter"} {
		t.Run(field, func(t *testing.T) {
			config := baseline
			runtime := "runtime"
			switch field {
			case "tenant":
				config.Profile.TenantID = "outside"
			case "profile":
				config.Profile.ID = "other"
			case "model":
				config.Profile.ModelName = "different-model"
			case "endpoint":
				config.Profile.EndpointURL = "https://other.example.test"
			case "prompt":
				config.Profile.PromptTemplate = "Use different guidance"
			case "runtime":
				config.Profile.RuntimeOptionsJSON = agentmodel.JSONObject(`{"a":2}`)
			case "capability options":
				config.Profile.CapabilityJSON = agentmodel.JSONObject(`{"supportsTools":false}`)
			case "credential purpose":
				config.CredentialPurpose = ports.ProviderCredentialPurposeOAuthBearer
			case "adapter":
				runtime = "runtime-v2"
			}
			got, err := providerConfigurationID(config, runtime)
			if err != nil || got == original {
				t.Fatalf("%s not reflected", field)
			}
		})
	}
}
func TestGoogleEvaluationIdentityRequiresADCRevisionAndOperatorBounds(t *testing.T) {
	config := identityConfig(t)
	config.Profile.RuntimeOptionsJSON = agentmodel.JSONObject("{}")
	config.CredentialPurpose = ports.ProviderCredentialPurposeServerADC
	factory := GoogleProviderProfileFactory{ServerADCProjectID: "project", ServerADCLocation: "us-central1"}
	if _, err := factory.ProviderConfigurationIdentity(context.Background(), config); err == nil {
		t.Fatal("unversioned ADC accepted")
	}
	factory.ServerADCCredentialVersion = "revision-one"
	first, err := factory.ProviderConfigurationIdentity(context.Background(), config)
	if err != nil || len(first) != 64 {
		t.Fatal("valid ADC identity rejected")
	}
	factory.ServerADCCredentialVersion = "revision-two"
	second, err := factory.ProviderConfigurationIdentity(context.Background(), config)
	if err != nil || first == second {
		t.Fatal("ADC rotation did not change identity")
	}
	factory.ServerADCProjectID = "other-project"
	third, err := factory.ProviderConfigurationIdentity(context.Background(), config)
	if err != nil || second == third {
		t.Fatal("operator project did not change identity")
	}
	config.Profile.RuntimeOptionsJSON = agentmodel.JSONObject(`{"projectId":"unapproved-project"}`)
	if _, err := factory.ProviderConfigurationIdentity(context.Background(), config); err == nil {
		t.Fatal("ADC override accepted")
	}
	if strings.Contains(first, "revision") {
		t.Fatal("identity is not opaque")
	}
}
