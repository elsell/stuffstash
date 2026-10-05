package httpapi

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) CreateProvider(ctx context.Context, tenant string, body []byte) (ports.Result[ports.ProviderProfile], error) {
	return providerResult(c.sdk.PostTenantsByTenantIdProviderProfilesWithBody(ctx, tenant, nil, "application/json", bytes.NewReader(body)))
}
func (c *Client) UpdateProvider(ctx context.Context, tenant, id string, body []byte) (ports.Result[ports.ProviderProfile], error) {
	return providerResult(c.sdk.PatchTenantsByTenantIdProviderProfilesByProviderProfileIdWithBody(ctx, tenant, id, nil, "application/json", bytes.NewReader(body)))
}
func (c *Client) ReplaceProviderCredential(ctx context.Context, tenant, id string, body []byte) (ports.Result[ports.ProviderProfile], error) {
	return providerResult(c.sdk.PutTenantsByTenantIdProviderProfilesByProviderProfileIdCredentialWithBody(ctx, tenant, id, nil, "application/json", bytes.NewReader(body)))
}
