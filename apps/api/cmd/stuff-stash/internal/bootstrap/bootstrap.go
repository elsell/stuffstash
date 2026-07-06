package bootstrap

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver"
	"github.com/stuffstash/stuff-stash/internal/adapters/observability"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func Run(ctx context.Context, cfg config.Config, observer ports.Observer) error {
	authenticator, err := buildAuthenticator(ctx, cfg)
	if err != nil {
		return err
	}
	authorizer, closeAuthorizer, err := buildAuthorizer(ctx, cfg)
	if err != nil {
		return err
	}
	defer recordCloseFailure(observer, closeAuthorizer)

	repositories, closeRepositories, err := buildRepositories(ctx, cfg)
	if err != nil {
		return err
	}
	defer recordCloseFailure(observer, closeRepositories)

	application, err := buildApplication(ctx, cfg, observer, authenticator, authorizer, repositories)
	if err != nil {
		return err
	}
	if err := replayLocalDevelopmentAuthorization(ctx, cfg, authorizer, repositories); err != nil {
		return err
	}
	server := httpserver.NewServerWithOptions(cfg.HTTPAddr, application, httpserver.Options{
		CORSAllowedOrigins: cfg.CORSAllowedOrigins,
		MobileAuth: httpserver.MobileAuthOptions{
			Issuer:      cfg.OIDCIssuer,
			ClientID:    cfg.OIDCMobileClientID,
			RedirectURI: cfg.OIDCMobileRedirectURI,
			Scopes:      cfg.OIDCMobileScopes,
		},
		MaxJSONBodyBytes:  cfg.HTTPMaxJSONBodyBytes,
		RateLimitDisabled: !cfg.HTTPRateLimitEnabled,
		RateLimitRequests: cfg.HTTPRateLimitRequests,
		RateLimitWindow:   cfg.HTTPRateLimitWindow,
		RateLimitBurst:    cfg.HTTPRateLimitBurst,
		Observer:          observer,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		ReadTimeout:       cfg.HTTPReadTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
	})
	startOutboxWorkers(ctx, application, observer, cfg)

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			observer.Record(context.Background(), observability.Event{
				Name:    observability.EventHTTPServerShutdownFailed,
				Message: "HTTP server shutdown failed",
				Fields:  map[string]string{"error": err.Error()},
			})
			return err
		}
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			observer.Record(context.Background(), observability.Event{
				Name:    observability.EventHTTPServerStartFailed,
				Message: "HTTP server failed",
				Fields:  map[string]string{"error": err.Error()},
			})
			return err
		}
	}
	return nil
}

func recordCloseFailure(observer ports.Observer, closeFn func() error) {
	if err := closeFn(); err != nil {
		observer.Record(context.Background(), ports.Event{
			Name:    ports.EventApplicationShutdownFailed,
			Message: "application shutdown failed",
			Fields:  map[string]string{"error": err.Error()},
		})
	}
}
