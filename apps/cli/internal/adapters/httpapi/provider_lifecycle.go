package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

func providerResult(response *http.Response, err error) (ports.Result[ports.ProviderProfile], error) {
	r, err := read[providerEnvelope[generated.ProviderProfileResponse]](response, err)
	if err != nil {
		return ports.Result[ports.ProviderProfile]{}, err
	}
	return ports.Result[ports.ProviderProfile]{Data: providerProfile(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) EnableProvider(ctx context.Context, tenant, id string) (ports.Result[ports.ProviderProfile], error) {
	return providerResult(c.sdk.PostTenantsByTenantIdProviderProfilesByProviderProfileIdEnable(ctx, tenant, id, nil))
}
func (c *Client) DisableProvider(ctx context.Context, tenant, id string) (ports.Result[ports.ProviderProfile], error) {
	return providerResult(c.sdk.PostTenantsByTenantIdProviderProfilesByProviderProfileIdDisable(ctx, tenant, id, nil))
}
func (c *Client) ArchiveProvider(ctx context.Context, tenant, id string) (ports.Result[ports.ProviderProfile], error) {
	return providerResult(c.sdk.PostTenantsByTenantIdProviderProfilesByProviderProfileIdArchive(ctx, tenant, id, nil))
}
func (c *Client) TestProvider(ctx context.Context, tenant, id string) (ports.Result[ports.ProviderTest], error) {
	r, err := read[generated.SuccessEnvelopeTestProviderProfileResponse](c.sdk.PostTenantsByTenantIdProviderProfilesByProviderProfileIdTest(ctx, tenant, id, nil))
	if err != nil {
		return ports.Result[ports.ProviderTest]{}, err
	}
	v := r.Data
	return ports.Result[ports.ProviderTest]{Schema: r.Schema, Meta: metadata(r.Meta), Data: ports.ProviderTest{ProfileID: v.ProviderProfileId, Status: v.Status, Message: v.Message, ProviderKind: v.ProviderKind, Capability: v.Capability, TestedAt: v.TestedAt}}, nil
}
