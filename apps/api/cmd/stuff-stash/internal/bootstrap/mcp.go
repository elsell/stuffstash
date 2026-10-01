package bootstrap

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/mcpserver"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"net/http"
)

func buildMCPHandler(cfg config.Config, application app.App, observer ports.Observer) (http.Handler, error) {
	settings, err := config.LoadMCP(cfg.AuthMode, cfg.OIDCIssuer)
	if err != nil {
		return nil, err
	}
	if !settings.Enabled {
		return nil, nil
	}
	return mcpserver.New(application, mcpserver.Options{PublicURL: settings.PublicURL, AuthMode: settings.AuthMode, Issuer: settings.Issuer, AllowedOrigins: cfg.CORSAllowedOrigins, MaxBodyBytes: cfg.HTTPMaxJSONBodyBytes, RequestTimeout: cfg.HTTPWriteTimeout, Observer: observer, Clock: ports.SystemClock{}})
}
