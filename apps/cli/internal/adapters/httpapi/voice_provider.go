package httpapi

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

type voiceProviderEnvelope struct {
	Schema *string                                       `json:"$schema,omitempty"`
	Data   *generated.VoiceProviderConfigurationResponse `json:"data"`
	Meta   generated.Meta                                `json:"meta"`
}

func (c *Client) VoiceProviderConfiguration(ctx context.Context, tenant string) (ports.Result[*ports.VoiceProviderConfiguration], error) {
	return voiceProviderResult(c.sdk.GetTenantsByTenantIdVoiceProviderConfiguration(ctx, tenant, nil))
}
func (c *Client) UpdateVoiceProviderConfiguration(ctx context.Context, tenant string, body []byte) (ports.Result[*ports.VoiceProviderConfiguration], error) {
	return voiceProviderResult(c.sdk.PutTenantsByTenantIdVoiceProviderConfigurationWithBody(ctx, tenant, nil, "application/json", bytes.NewReader(body)))
}
func voiceProviderResult(response *http.Response, err error) (ports.Result[*ports.VoiceProviderConfiguration], error) {
	r, err := read[voiceProviderEnvelope](response, err)
	if err != nil {
		return ports.Result[*ports.VoiceProviderConfiguration]{}, err
	}
	var data *ports.VoiceProviderConfiguration
	if r.Data != nil {
		data = voiceProviderConfiguration(*r.Data)
	}
	return ports.Result[*ports.VoiceProviderConfiguration]{Data: data, Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func voiceProviderConfiguration(v generated.VoiceProviderConfigurationResponse) *ports.VoiceProviderConfiguration {
	result := &ports.VoiceProviderConfiguration{ProfileIDs: ports.VoiceProviderProfiles(v.ProfileIds), Readiness: string(v.Readiness), TenantID: v.TenantId, UpdatedAt: v.UpdatedAt}
	slots := v.Slots.GetOrEmpty()
	if slots != nil {
		result.Slots = make([]ports.VoiceProviderSlot, len(slots))
	}
	for i, slot := range slots {
		mapped := ports.VoiceProviderSlot{Capability: slot.Capability, Issues: slot.Issues.GetOrEmpty(), Label: slot.Label, Readiness: string(slot.Readiness), RecommendedAction: string(slot.RecommendedAction), SelectedProfileID: slot.SelectedProfileId, SelectionSource: string(slot.SelectionSource)}
		if slot.SelectedProfile != nil {
			profile := voiceProviderSummary(*slot.SelectedProfile)
			mapped.SelectedProfile = &profile
		}
		duplicates := slot.DuplicateProfiles.GetOrEmpty()
		if duplicates != nil {
			mapped.DuplicateProfiles = make([]ports.VoiceProviderSummary, len(duplicates))
		}
		for j, p := range duplicates {
			mapped.DuplicateProfiles[j] = voiceProviderSummary(p)
		}
		result.Slots[i] = mapped
	}
	return result
}
func voiceProviderSummary(v generated.ProviderProfileSummaryResponse) ports.VoiceProviderSummary {
	return ports.VoiceProviderSummary{Capability: v.Capability, CredentialPurpose: v.CredentialPurpose, CredentialStatus: v.CredentialStatus, DisplayName: v.DisplayName, ID: v.Id, LastTestedAt: v.LastTestedAt, LifecycleState: v.LifecycleState, ModelName: v.ModelName, ProviderKind: v.ProviderKind}
}
