package bootstrap

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/dbmigrations"
	"github.com/stuffstash/stuff-stash/internal/config"
)

func TestRunMigrationCommandRejectsMissingAction(t *testing.T) {
	var output bytes.Buffer

	err := RunMigrationCommand(context.Background(), config.Config{DatabaseDSN: "postgres://example"}, nil, &output)
	if err == nil {
		t.Fatalf("expected missing migration action error")
	}
}

func TestRunMigrationCommandRejectsUnknownAction(t *testing.T) {
	var output bytes.Buffer

	err := RunMigrationCommand(context.Background(), config.Config{DatabaseDSN: "postgres://example"}, []string{"sideways"}, &output)
	if err == nil {
		t.Fatalf("expected unknown migration action error")
	}
}

func TestRunMigrationCommandRejectsMissingDSN(t *testing.T) {
	var output bytes.Buffer

	err := RunMigrationCommand(context.Background(), config.Config{}, []string{"up"}, &output)
	if err == nil {
		t.Fatalf("expected missing database dsn error")
	}
}

func TestValidateMigrationStatusRejectsEmptySchema(t *testing.T) {
	err := validateMigrationStatus(dbmigrations.Status{Latest: 3, Empty: true})
	if err == nil {
		t.Fatalf("expected empty schema error")
	}
}

func TestValidateMigrationStatusRejectsDirtySchema(t *testing.T) {
	err := validateMigrationStatus(dbmigrations.Status{Version: 2, Latest: 3, Dirty: true})
	if err == nil {
		t.Fatalf("expected dirty schema error")
	}
}

func TestValidateMigrationStatusRejectsOutdatedSchema(t *testing.T) {
	err := validateMigrationStatus(dbmigrations.Status{Version: 2, Latest: 3})
	if err == nil {
		t.Fatalf("expected outdated schema error")
	}
}

func TestValidateMigrationStatusAcceptsCurrentSchema(t *testing.T) {
	err := validateMigrationStatus(dbmigrations.Status{Version: 3, Latest: 3})
	if err != nil {
		t.Fatalf("validate current schema: %v", err)
	}
}

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

func TestBuildBlobStorageAcceptsFilesystemMode(t *testing.T) {
	store, err := buildBlobStorage(config.Config{BlobStorageMode: "filesystem", BlobStoragePath: t.TempDir()})
	if err != nil {
		t.Fatalf("build filesystem blob storage: %v", err)
	}
	if store == nil {
		t.Fatalf("expected blob storage")
	}
}

func TestBuildBlobStorageRejectsUnknownMode(t *testing.T) {
	_, err := buildBlobStorage(config.Config{BlobStorageMode: "unknown"})
	if err == nil {
		t.Fatalf("expected unsupported blob storage mode error")
	}
}

func TestBuildBlobStorageRejectsIncompleteS3Config(t *testing.T) {
	_, err := buildBlobStorage(config.Config{BlobStorageMode: "s3"})
	if err == nil {
		t.Fatalf("expected incomplete S3 config error")
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
