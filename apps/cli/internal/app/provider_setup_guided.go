package app

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (r Runner) providerChoice(ctx context.Context, title string, choices []ports.Choice) (string, error) {
	choices = append([]ports.Choice{{ID: "cancel", Label: "Cancel", Detail: "Do not change provider settings"}}, choices...)
	v, err := r.Picker.Pick(ctx, title, choices)
	if err != nil {
		return "", err
	}
	if v == "cancel" {
		return "", context.Canceled
	}
	for _, c := range choices {
		if v == c.ID {
			return v, nil
		}
	}
	return "", ports.Failure("input", "The selected option is not available. Start provider setup again.")
}
func (r Runner) guideProviderSetup(ctx context.Context, action string) ([]byte, error) {
	fields := map[string]any{}
	if action == "credential" {
		purpose, err := r.providerChoice(ctx, "Credential purpose", []ports.Choice{{ID: "api_key", Label: "API key"}, {ID: "oauth_bearer", Label: "OAuth bearer token"}, {ID: "server_adc", Label: "Server application default credentials", Detail: "Use the server's Google credentials; no secret input"}})
		if err != nil {
			return nil, err
		}
		fields["purpose"] = purpose
		if purpose != "server_adc" {
			if r.SecretInput == nil {
				return nil, ports.Failure("configuration", "Masked credential input is not available. Supply --input FILE|- instead.")
			}
			secret, err := r.SecretInput.ReadSecret(ctx, "Provider credential", 65536)
			if err != nil {
				return nil, err
			}
			fields["credential"] = secret
		}
		return json.Marshal(fields)
	}
	if action == "create" {
		kind, err := r.providerChoice(ctx, "Provider", []ports.Choice{{ID: "gemini", Label: "Gemini"}, {ID: "openai_compatible", Label: "OpenAI compatible"}, {ID: "local_http", Label: "Local HTTP"}})
		if err != nil {
			return nil, err
		}
		fields["providerKind"] = kind
		capability, err := r.providerChoice(ctx, "Capability", []ports.Choice{{ID: "speech_to_text", Label: "Speech input"}, {ID: "language_inference", Label: "Language inference"}, {ID: "text_to_speech", Label: "Spoken output"}})
		if err != nil {
			return nil, err
		}
		fields["capability"] = capability
		name, err := r.TextInput.ReadText(ctx, "Profile display name", 120)
		if err != nil {
			return nil, err
		}
		fields["displayName"] = name
	}
	for {
		choices := []ports.Choice{{ID: "save", Label: "Review and save"}, {ID: "displayName", Label: "Display name"}, {ID: "endpointUrl", Label: "Endpoint URL", Detail: "Use the provider's documented endpoint; no default is supplied"}, {ID: "modelName", Label: "Model or deployment name"}, {ID: "promptTemplate", Label: "Prompt template"}, {ID: "runtimeOptions", Label: "Runtime options", Detail: "Non-secret JSON object"}, {ID: "capabilityMetadata", Label: "Capability metadata", Detail: "Non-secret JSON object"}}
		if action == "create" {
			choices = append(choices, ports.Choice{ID: "enable", Label: "Initial enabled state"})
		}
		field, err := r.providerChoice(ctx, "Provider settings", choices)
		if err != nil {
			return nil, err
		}
		if field == "save" {
			return json.Marshal(fields)
		}
		if field == "enable" {
			v, err := r.providerChoice(ctx, "Initial state", []ports.Choice{{ID: "false", Label: "Disabled"}, {ID: "true", Label: "Enabled"}})
			if err != nil {
				return nil, err
			}
			fields[field] = v == "true"
			continue
		}
		value, err := r.TextInput.ReadText(ctx, field+" (empty clears text)", 65536)
		if err != nil {
			return nil, err
		}
		if field == "runtimeOptions" || field == "capabilityMetadata" {
			var object map[string]json.RawMessage
			if json.Unmarshal([]byte(value), &object) != nil || object == nil {
				return nil, ports.Failure("input", "Enter a JSON object for this field, or use --input FILE for complete settings.")
			}
			fields[field] = object
		} else {
			fields[field] = value
		}
	}
}
