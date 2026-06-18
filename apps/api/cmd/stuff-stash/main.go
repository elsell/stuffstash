package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/auth"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver"
	"github.com/stuffstash/stuff-stash/internal/adapters/idgen"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/adapters/observability"
	"github.com/stuffstash/stuff-stash/internal/adapters/spicedb"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	observer := observability.NewFanOut(
		observability.NewSlogObserver(logger),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	authenticator, err := buildAuthenticator(ctx, cfg)
	if err != nil {
		recordStartupFailure(observer, err)
		os.Exit(1)
	}
	authorizer, closeAuthorizer, err := buildAuthorizer(ctx, cfg)
	if err != nil {
		recordStartupFailure(observer, err)
		os.Exit(1)
	}
	defer func() {
		if err := closeAuthorizer(); err != nil {
			observer.Record(context.Background(), ports.Event{
				Name:    ports.EventApplicationShutdownFailed,
				Message: "application shutdown failed",
				Fields:  map[string]string{"error": err.Error()},
			})
		}
	}()

	store := memory.NewStore()
	application := app.New(app.Dependencies{
		Observer:    observer,
		Auth:        authenticator,
		Authorizer:  authorizer,
		Tenants:     store,
		Inventories: store,
		IDs:         idgen.NewULIDGenerator(),
	})
	server := httpserver.NewServer(cfg.HTTPAddr, application)

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
			os.Exit(1)
		}
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			observer.Record(context.Background(), observability.Event{
				Name:    observability.EventHTTPServerStartFailed,
				Message: "HTTP server failed",
				Fields:  map[string]string{"error": err.Error()},
			})
			os.Exit(1)
		}
	}
}

func buildAuthenticator(ctx context.Context, cfg config.Config) (ports.Authenticator, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.AuthMode)) {
	case "local-dev":
		return auth.NewLocalDevAuthenticator(), nil
	case "oidc":
		if strings.TrimSpace(cfg.OIDCIssuer) == "" || strings.TrimSpace(cfg.OIDCClientID) == "" {
			return nil, errors.New("oidc issuer and client id are required")
		}
		return auth.NewOIDCAuthenticatorFromIssuer(ctx, cfg.OIDCIssuer, cfg.OIDCClientID)
	default:
		return nil, errors.New("unsupported authentication mode")
	}
}

func buildAuthorizer(ctx context.Context, cfg config.Config) (ports.Authorizer, func() error, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.AuthzMode)) {
	case "memory":
		return memory.NewAuthorizer(), func() error { return nil }, nil
	case "spicedb":
		gateway, err := spicedb.NewGateway(cfg.SpiceDBEndpoint, cfg.SpiceDBPresharedKey, cfg.SpiceDBTLSEnabled)
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

type schemaBootstrapper interface {
	BootstrapSchema(ctx context.Context, schema string) error
}

func bootstrapSpiceDBSchema(ctx context.Context, authorizer schemaBootstrapper, schemaPath string) error {
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.NewTimer(30 * time.Second)
	defer deadline.Stop()

	var lastErr error
	for {
		if err := authorizer.BootstrapSchema(ctx, string(schema)); err == nil {
			return nil
		} else {
			lastErr = err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return fmt.Errorf("bootstrap spicedb schema: %w", lastErr)
		case <-ticker.C:
		}
	}
}

func recordStartupFailure(observer ports.Observer, err error) {
	observer.Record(context.Background(), ports.Event{
		Name:    ports.EventApplicationStartupFailed,
		Message: "application startup failed",
		Fields:  map[string]string{"error": err.Error()},
	})
}
