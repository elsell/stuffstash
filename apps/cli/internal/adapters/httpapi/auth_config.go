package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) AuthConfig(ctx context.Context) (ports.AuthConfig, error) {
	r, err := read[generated.SuccessEnvelopeCLIAuthMetadata](c.sdk.GetCliAuthConfig(ctx))
	if err != nil {
		return ports.AuthConfig{}, err
	}
	v := r.Data
	return ports.AuthConfig{Issuer: v.Issuer, ClientID: v.ClientId, Scopes: v.Scopes.GetOrEmpty(), LoginMethods: v.LoginMethods.GetOrEmpty(), LoopbackHost: v.LoopbackRedirect.Host, LoopbackPathPrefix: v.LoopbackRedirect.PathPrefix, EphemeralPort: v.LoopbackRedirect.EphemeralPort}, nil
}
