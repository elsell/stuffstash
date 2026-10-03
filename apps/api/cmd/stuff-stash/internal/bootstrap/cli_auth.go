package bootstrap

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/identity/dto"
	"github.com/stuffstash/stuff-stash/internal/config"
)

func buildCLIAuthMetadata(ctx context.Context, cfg config.Config) (*dto.CLIAuthMetadata, error) {
	if strings.ToLower(strings.TrimSpace(cfg.AuthMode)) != "oidc" || cfg.OIDCCLIClientID == "" {
		return nil, nil
	}
	provider, err := oidc.NewProvider(ctx, cfg.OIDCIssuer)
	if err != nil {
		return nil, err
	}
	var discovery struct {
		DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint"`
	}
	if err = provider.Claims(&discovery); err != nil {
		return nil, err
	}
	scopes := make([]string, 0, len(cfg.OIDCCLIScopes))
	hasOpenID := false
	seen := map[string]bool{}
	for _, scope := range cfg.OIDCCLIScopes {
		scope = strings.TrimSpace(scope)
		if scope == "" || seen[scope] {
			continue
		}
		if strings.ContainsAny(scope, " \t\r\n") {
			return nil, errors.New("CLI OIDC scopes must be comma-separated scope names")
		}
		if scope == "openid" {
			hasOpenID = true
		}
		seen[scope] = true
		scopes = append(scopes, scope)
	}
	if !hasOpenID {
		return nil, errors.New("CLI OIDC scopes must include openid")
	}
	methods := []string{"authorization_code"}
	if cfg.OIDCCLIDeviceAuthEnabled && discovery.DeviceAuthorizationEndpoint != "" {
		endpoint, parseErr := url.Parse(discovery.DeviceAuthorizationEndpoint)
		issuer, issuerErr := url.Parse(cfg.OIDCIssuer)
		if parseErr != nil || issuerErr != nil || endpoint.Host == "" || endpoint.User != nil || endpoint.Fragment != "" || (endpoint.Scheme != "https" && !(issuer.Scheme == "http" && endpoint.Scheme == "http" && issuer.Host == endpoint.Host)) {
			return nil, errors.New("invalid CLI OIDC device authorization endpoint")
		}
		methods = append(methods, "device_code")
	}
	return &dto.CLIAuthMetadata{Issuer: cfg.OIDCIssuer, ClientID: cfg.OIDCCLIClientID, Scopes: scopes, LoginMethods: methods, LoopbackRedirect: dto.CLILoopbackRedirect{Host: "127.0.0.1", PathPrefix: "/callback/", EphemeralPort: true}}, nil
}
