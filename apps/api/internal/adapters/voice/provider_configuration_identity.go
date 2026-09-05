package voice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// Factories implement this in addition to provider construction. No model call
// or credential token refresh is allowed while computing an identity.
type ProviderConfigurationIdentityFactory interface {
	ProviderConfigurationIdentity(context.Context, ProviderProfileProviderConfig) (string, error)
}

func providerConfigurationID(config ProviderProfileProviderConfig, runtimeIdentity string) (string, error) {
	if strings.TrimSpace(config.CredentialVersionID) == "" || strings.TrimSpace(runtimeIdentity) == "" {
		return "", ports.ErrInvalidProviderInput
	}
	runtime, err := canonicalProviderOptions(config.Profile.RuntimeOptionsJSON.String())
	if err != nil {
		return "", err
	}
	capabilities, err := canonicalProviderOptions(config.Profile.CapabilityJSON.String())
	if err != nil {
		return "", err
	}
	profile := config.Profile
	encoded, err := json.Marshal(struct {
		Version                                                    string
		Tenant, Profile, Capability, Kind, Endpoint, Model, Prompt string
		Runtime, Capabilities                                      map[string]any
		CredentialPurpose                                          ports.ProviderCredentialPurpose
		CredentialVersion, AdapterRuntime                          string
	}{
		Version: "provider-configuration-v1", Tenant: profile.TenantID.String(), Profile: profile.ID.String(),
		Capability: profile.Capability.String(), Kind: profile.ProviderKind.String(), Endpoint: profile.EndpointURL.String(),
		Model: profile.ModelName.String(), Prompt: profile.PromptTemplate.String(), Runtime: runtime, Capabilities: capabilities,
		CredentialPurpose: config.CredentialPurpose, CredentialVersion: config.CredentialVersionID, AdapterRuntime: runtimeIdentity,
	})
	if err != nil {
		return "", ports.ErrInvalidProviderInput
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}
func canonicalProviderOptions(raw string) (map[string]any, error) {
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	var value map[string]any
	if err := decoder.Decode(&value); err != nil || value == nil {
		return nil, ports.ErrInvalidProviderInput
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, ports.ErrInvalidProviderInput
	}
	return value, nil
}

func (f GoogleProviderProfileFactory) ProviderConfigurationIdentity(_ context.Context, config ProviderProfileProviderConfig) (string, error) {
	if config.Profile.ProviderKind != agentmodel.ProviderKindGemini || config.Profile.Capability != agentmodel.ProviderCapabilityLanguageInference || config.Profile.ModelName == "" {
		return "", ports.ErrInvalidProviderInput
	}
	options, err := providerRuntimeOptions(config.Profile)
	if err != nil {
		return "", err
	}
	if _, err := httpTimeoutOption(options); err != nil {
		return "", err
	}
	runtime := struct{ Contract, Project, Location, QuotaProject, CredentialVersion string }{Contract: "google-gemini-v1"}
	switch config.CredentialPurpose {
	case ports.ProviderCredentialPurposeAPIKey:
	case ports.ProviderCredentialPurposeOAuthBearer, ports.ProviderCredentialPurposeServerADC:
		runtime.Project, err = f.googleProjectOption(config.CredentialPurpose, options)
		if err != nil {
			return "", err
		}
		runtime.Location, err = f.googleLocationOption(config.CredentialPurpose, options)
		if err != nil {
			return "", err
		}
		if runtime.Project == "" || runtime.Location == "" {
			return "", ports.ErrInvalidProviderInput
		}
		if err := f.validateGoogleQuotaProjectOption(config.CredentialPurpose, options); err != nil {
			return "", err
		}
		runtime.QuotaProject = f.googleQuotaProjectOption(config.CredentialPurpose, options)
		if config.CredentialPurpose == ports.ProviderCredentialPurposeServerADC {
			runtime.CredentialVersion = strings.TrimSpace(f.ServerADCCredentialVersion)
			if runtime.CredentialVersion == "" {
				return "", ports.ErrInvalidProviderInput
			}
		}
	default:
		return "", ports.ErrInvalidProviderInput
	}
	encoded, err := json.Marshal(runtime)
	if err != nil {
		return "", ports.ErrInvalidProviderInput
	}
	return providerConfigurationID(config, string(encoded))
}
