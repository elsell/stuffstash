package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/config"
)

func TestBuildAuthenticatorAcceptsLocalDevMode(t *testing.T) {
	authenticator, err := buildAuthenticator(context.Background(), config.Config{AuthMode: "local-dev"})
	if err != nil {
		t.Fatalf("build authenticator: %v", err)
	}
	if authenticator == nil {
		t.Fatalf("expected authenticator")
	}
}

func TestBuildAuthenticatorRejectsUnknownMode(t *testing.T) {
	_, err := buildAuthenticator(context.Background(), config.Config{AuthMode: "unknown"})
	if err == nil {
		t.Fatalf("expected unsupported mode error")
	}
}

func TestBuildAuthenticatorRejectsIncompleteOIDCConfig(t *testing.T) {
	_, err := buildAuthenticator(context.Background(), config.Config{AuthMode: "oidc"})
	if err == nil {
		t.Fatalf("expected incomplete OIDC config error")
	}
}

func TestBuildAuthorizerAcceptsMemoryMode(t *testing.T) {
	authorizer, closeAuthorizer, err := buildAuthorizer(context.Background(), config.Config{AuthzMode: "memory"})
	if err != nil {
		t.Fatalf("build authorizer: %v", err)
	}
	if authorizer == nil {
		t.Fatalf("expected authorizer")
	}
	if err := closeAuthorizer(); err != nil {
		t.Fatalf("close authorizer: %v", err)
	}
}

func TestBuildAuthorizerRejectsUnknownMode(t *testing.T) {
	_, _, err := buildAuthorizer(context.Background(), config.Config{AuthzMode: "unknown"})
	if err == nil {
		t.Fatalf("expected unsupported mode error")
	}
}

func TestBuildRepositoriesAcceptsMemoryMode(t *testing.T) {
	repositories, closeRepositories, err := buildRepositories(context.Background(), config.Config{RepositoryMode: "memory"})
	if err != nil {
		t.Fatalf("build repositories: %v", err)
	}
	if repositories.tenants == nil || repositories.inventories == nil {
		t.Fatalf("expected repositories")
	}
	if err := closeRepositories(); err != nil {
		t.Fatalf("close repositories: %v", err)
	}
}

func TestBuildRepositoriesRejectsUnknownMode(t *testing.T) {
	_, _, err := buildRepositories(context.Background(), config.Config{RepositoryMode: "unknown"})
	if err == nil {
		t.Fatalf("expected unsupported mode error")
	}
}

func TestBuildRepositoriesRejectsPostgresWithoutDSN(t *testing.T) {
	_, _, err := buildRepositories(context.Background(), config.Config{RepositoryMode: "postgres"})
	if err == nil {
		t.Fatalf("expected missing database dsn error")
	}
}

func TestBootstrapSpiceDBSchemaReadsSchemaFile(t *testing.T) {
	schemaPath := filepath.Join(t.TempDir(), "schema.zed")
	if err := os.WriteFile(schemaPath, []byte("definition user {}"), 0o600); err != nil {
		t.Fatalf("write schema: %v", err)
	}
	bootstrapper := &fakeSchemaBootstrapper{}

	if err := bootstrapSpiceDBSchema(context.Background(), bootstrapper, schemaPath); err != nil {
		t.Fatalf("bootstrap schema: %v", err)
	}

	if bootstrapper.schema != "definition user {}" {
		t.Fatalf("expected schema content, got %q", bootstrapper.schema)
	}
}

func TestBootstrapSpiceDBSchemaRetriesTransientFailure(t *testing.T) {
	schemaPath := filepath.Join(t.TempDir(), "schema.zed")
	if err := os.WriteFile(schemaPath, []byte("definition user {}"), 0o600); err != nil {
		t.Fatalf("write schema: %v", err)
	}
	bootstrapper := &fakeSchemaBootstrapper{failuresRemaining: 1}

	if err := bootstrapSpiceDBSchema(context.Background(), bootstrapper, schemaPath); err != nil {
		t.Fatalf("bootstrap schema: %v", err)
	}

	if bootstrapper.calls != 2 {
		t.Fatalf("expected 2 bootstrap attempts, got %d", bootstrapper.calls)
	}
}

type fakeSchemaBootstrapper struct {
	failuresRemaining int
	calls             int
	schema            string
}

func (f *fakeSchemaBootstrapper) BootstrapSchema(_ context.Context, schema string) error {
	f.calls++
	f.schema = schema
	if f.failuresRemaining > 0 {
		f.failuresRemaining--
		return errors.New("not ready")
	}
	return nil
}
