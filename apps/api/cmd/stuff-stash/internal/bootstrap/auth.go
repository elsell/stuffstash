package bootstrap

import (
	"context"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/adapters/auth"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/adapters/spicedb"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func buildAuthenticator(ctx context.Context, cfg config.Config) (ports.Authenticator, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.AuthMode)) {
	case "local-dev":
		return auth.NewLocalDevAuthenticator(), nil
	case "oidc":
		if strings.TrimSpace(cfg.OIDCIssuer) == "" || len(cfg.OIDCClientIDs) == 0 {
			return nil, errors.New("oidc issuer and client id are required")
		}
		return auth.NewOIDCAuthenticatorFromIssuerForClientIDs(ctx, cfg.OIDCIssuer, cfg.OIDCClientIDs)
	default:
		return nil, errors.New("unsupported authentication mode")
	}
}

func buildAuthorizer(ctx context.Context, cfg config.Config) (ports.Authorizer, func() error, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.AuthzMode)) {
	case "memory":
		return memory.NewAuthorizer(), func() error { return nil }, nil
	case "spicedb":
		gateway, err := spicedb.NewGateway(cfg.SpiceDBEndpoint, cfg.SpiceDBPresharedKey, cfg.SpiceDBTLSEnabled, cfg.SpiceDBCAPath)
		if err != nil {
			return nil, nil, err
		}
		authorizer := spicedb.NewAuthorizer(gateway)
		if cfg.SpiceDBBootstrapSchema {
			if err := bootstrapSpiceDBSchema(ctx, authorizer, cfg.SpiceDBSchemaPath); err != nil {
				_ = gateway.Close()
				return nil, nil, err
			}
		}
		return authorizer, gateway.Close, nil
	default:
		return nil, nil, errors.New("unsupported authorization mode")
	}
}
