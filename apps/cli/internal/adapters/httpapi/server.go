package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) ServerInfo(ctx context.Context) (ports.Result[ports.ServerInfo], error) {
	r, err := read[generated.SuccessEnvelopeInstanceResponse](c.sdk.GetInstance(ctx))
	if err != nil {
		return ports.Result[ports.ServerInfo]{}, err
	}
	return ports.Result[ports.ServerInfo]{Data: ports.ServerInfo{InstanceID: r.Data.InstanceId, ProtocolVersion: r.Data.ProtocolVersion}, Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) ServerAuthConfig(ctx context.Context) (ports.Result[ports.ServerAuthConfig], error) {
	r, err := read[generated.SuccessEnvelopeCLIAuthMetadata](c.sdk.GetCliAuthConfig(ctx))
	if err != nil {
		return ports.Result[ports.ServerAuthConfig]{}, err
	}
	v := r.Data
	return ports.Result[ports.ServerAuthConfig]{Data: ports.ServerAuthConfig{Issuer: v.Issuer, ClientID: v.ClientId, Scopes: v.Scopes.GetOrEmpty(), LoginMethods: v.LoginMethods.GetOrEmpty(), LoopbackRedirect: ports.LoopbackRedirect{Host: v.LoopbackRedirect.Host, PathPrefix: v.LoopbackRedirect.PathPrefix, EphemeralPort: v.LoopbackRedirect.EphemeralPort}}, Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
