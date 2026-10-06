package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

// Decode the generated profile DTOs directly so UseNumber also applies to lists.
// The generated nullable slice decoder does not retain arbitrary integer precision.
type providerEnvelope[T any] struct {
	Schema *string        `json:"$schema,omitempty"`
	Data   T              `json:"data"`
	Meta   generated.Meta `json:"meta"`
}

func providerProfile(v generated.ProviderProfileResponse) ports.ProviderProfile {
	return ports.ProviderProfile{Capability: v.Capability, CapabilityMetadata: v.CapabilityMetadata, CreatedAt: v.CreatedAt, CredentialStatus: v.CredentialStatus, DisplayName: v.DisplayName, EndpointURL: v.EndpointUrl, ID: v.Id, LastTestedAt: v.LastTestedAt, LifecycleState: v.LifecycleState, ModelName: v.ModelName, PromptTemplate: v.PromptTemplate, ProviderKind: v.ProviderKind, RuntimeOptions: v.RuntimeOptions, TenantID: v.TenantId, UpdatedAt: v.UpdatedAt}
}
func (c *Client) ProviderProfiles(ctx context.Context, tenant string) (ports.Result[[]ports.ProviderProfile], error) {
	r, err := read[providerEnvelope[[]generated.ProviderProfileResponse]](c.sdk.GetTenantsByTenantIdProviderProfiles(ctx, tenant, nil))
	if err != nil {
		return ports.Result[[]ports.ProviderProfile]{}, err
	}
	var profiles []ports.ProviderProfile
	if r.Data != nil {
		profiles = make([]ports.ProviderProfile, len(r.Data))
	}
	for i, v := range r.Data {
		profiles[i] = providerProfile(v)
	}
	return ports.Result[[]ports.ProviderProfile]{Data: profiles, Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) ProviderProfile(ctx context.Context, tenant, id string) (ports.Result[ports.ProviderProfile], error) {
	r, err := read[providerEnvelope[generated.ProviderProfileResponse]](c.sdk.GetTenantsByTenantIdProviderProfilesByProviderProfileId(ctx, tenant, id, nil))
	if err != nil {
		return ports.Result[ports.ProviderProfile]{}, err
	}
	return ports.Result[ports.ProviderProfile]{Data: providerProfile(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
