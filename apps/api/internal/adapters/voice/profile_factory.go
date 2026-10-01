package voice

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// ProviderProfileFactory dispatches all profile consumers through one adapter policy.
type ProviderProfileFactory struct {
	Google              GoogleProviderProfileFactory
	CompatibleEndpoints []string
}

func (f ProviderProfileFactory) SpeechToTextProvider(ctx context.Context, c ProviderProfileProviderConfig) (ports.SpeechToTextProvider, error) {
	if c.Profile.ProviderKind != agentmodel.ProviderKindGemini {
		return nil, ports.ErrInvalidProviderInput
	}
	return f.Google.SpeechToTextProvider(ctx, c)
}
func (f ProviderProfileFactory) TextToSpeechProvider(ctx context.Context, c ProviderProfileProviderConfig) (ports.TextToSpeechProvider, error) {
	if c.Profile.ProviderKind != agentmodel.ProviderKindGemini {
		return nil, ports.ErrInvalidProviderInput
	}
	return f.Google.TextToSpeechProvider(ctx, c)
}
func (f ProviderProfileFactory) ConversationModelProvider(ctx context.Context, c ProviderProfileProviderConfig) (ports.ConversationModel, error) {
	if c.Profile.ProviderKind == agentmodel.ProviderKindGemini {
		return f.Google.ConversationModelProvider(ctx, c)
	}
	endpoint, timeout, err := f.compatibleConfiguration(c)
	if err != nil {
		return nil, err
	}
	return compatibleConversation{endpoint: endpoint + "/chat/completions", model: c.Profile.ModelName.String(), credential: string(c.Credential), client: &http.Client{Timeout: timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}
func (f ProviderProfileFactory) ProviderConfigurationIdentity(ctx context.Context, c ProviderProfileProviderConfig) (string, error) {
	if c.Profile.ProviderKind == agentmodel.ProviderKindGemini {
		return f.Google.ProviderConfigurationIdentity(ctx, c)
	}
	if _, _, err := f.compatibleConfiguration(c); err != nil {
		return "", err
	}
	return providerConfigurationID(c, "openai-compatible-chat-v1")
}
func (f ProviderProfileFactory) compatibleConfiguration(c ProviderProfileProviderConfig) (string, time.Duration, error) {
	p := c.Profile
	if p.Capability != agentmodel.ProviderCapabilityLanguageInference || (p.ProviderKind != agentmodel.ProviderKindOpenAICompatible && p.ProviderKind != agentmodel.ProviderKindLocalHTTP) || strings.TrimSpace(p.ModelName.String()) == "" {
		return "", 0, ports.ErrInvalidProviderInput
	}
	if (c.CredentialPurpose != ports.ProviderCredentialPurposeAPIKey && c.CredentialPurpose != ports.ProviderCredentialPurposeOAuthBearer) || strings.TrimSpace(string(c.Credential)) == "" || strings.ContainsAny(string(c.Credential), "\r\n") {
		return "", 0, ports.ErrInvalidProviderInput
	}
	endpoint := strings.TrimRight(p.EndpointURL.String(), "/")
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" || parsed.RawPath != "" || strings.Contains(parsed.Path, "..") || (parsed.Scheme != "https" && !(p.ProviderKind == agentmodel.ProviderKindLocalHTTP && parsed.Scheme == "http")) {
		return "", 0, ports.ErrInvalidProviderInput
	}
	allowed := false
	for _, candidate := range f.CompatibleEndpoints {
		if strings.TrimRight(strings.TrimSpace(candidate), "/") == endpoint {
			allowed = true
			break
		}
	}
	if !allowed {
		return "", 0, ports.ErrInvalidProviderInput
	}
	options, err := providerRuntimeOptions(p)
	if err != nil {
		return "", 0, err
	}
	timeout, err := httpTimeoutOption(options)
	if err != nil {
		return "", 0, err
	}
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	return endpoint, timeout, nil
}
