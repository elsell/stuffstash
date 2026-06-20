package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/auth"
	"github.com/stuffstash/stuff-stash/internal/adapters/blobstore"
	"github.com/stuffstash/stuff-stash/internal/adapters/dbmigrations"
	"github.com/stuffstash/stuff-stash/internal/adapters/gormstore"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver"
	"github.com/stuffstash/stuff-stash/internal/adapters/idgen"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/adapters/observability"
	"github.com/stuffstash/stuff-stash/internal/adapters/spicedb"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	observer := observability.NewFanOut(
		observability.NewSlogObserver(logger),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		if err := runMigrationCommand(ctx, cfg, os.Args[2:], os.Stdout); err != nil {
			recordStartupFailure(observer, err)
			os.Exit(1)
		}
		return
	}

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

	repositories, closeRepositories, err := buildRepositories(ctx, cfg)
	if err != nil {
		recordStartupFailure(observer, err)
		os.Exit(1)
	}
	defer func() {
		if err := closeRepositories(); err != nil {
			observer.Record(context.Background(), ports.Event{
				Name:    ports.EventApplicationShutdownFailed,
				Message: "application shutdown failed",
				Fields:  map[string]string{"error": err.Error()},
			})
		}
	}()

	application := app.New(app.Dependencies{
		Observer:                      observer,
		Auth:                          authenticator,
		Authorizer:                    authorizer,
		Tenants:                       repositories.tenants,
		Inventories:                   repositories.inventories,
		CustomAssetTypes:              repositories.customAssetTypes,
		CustomFields:                  repositories.customFields,
		Assets:                        repositories.assets,
		Search:                        repositories.search,
		Attachments:                   repositories.attachments,
		Blobs:                         repositories.blobs,
		Audit:                         repositories.audit,
		Outbox:                        repositories.outbox,
		IDs:                           idgen.NewULIDGenerator(),
		AuthorizationOutboxDrainLimit: cfg.AuthorizationOutboxDrainLimit,
		AuthorizationOutboxClaimLease: cfg.AuthorizationOutboxClaimLease,
		InvitationTTL:                 cfg.InvitationTTL,
		DefaultPageLimit:              cfg.DefaultPageLimit,
		MaxPageLimit:                  cfg.MaxPageLimit,
		MaxAttachmentBytes:            cfg.MaxAttachmentBytes,
	})
	server := httpserver.NewServerWithOptions(cfg.HTTPAddr, application, httpserver.Options{
		CORSAllowedOrigins: cfg.CORSAllowedOrigins,
	})
	go drainAuthorizationOutbox(ctx, application, observer, cfg.AuthorizationOutboxDrainLimit, cfg.AuthorizationOutboxDrainInterval)

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

func drainAuthorizationOutbox(ctx context.Context, application app.App, observer ports.Observer, limit int, interval time.Duration) {
	drain := func() {
		if err := application.DrainAuthorizationOutbox(ctx, limit); err != nil {
			observer.Record(ctx, ports.Event{
				Name:    ports.EventAuthorizationOutboxFailed,
				Message: "authorization outbox background drain failed",
				Fields:  map[string]string{"error": err.Error()},
			})
		}
	}

	drain()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			drain()
		}
	}
}

type repositories struct {
	tenants          ports.TenantRepository
	inventories      ports.InventoryRepository
	customAssetTypes ports.CustomAssetTypeRepository
	customFields     ports.CustomFieldDefinitionRepository
	assets           ports.AssetRepository
	search           ports.AssetSearchRepository
	attachments      ports.AttachmentRepository
	blobs            ports.BlobStorage
	audit            ports.AuditRepository
	outbox           ports.AuthorizationOutbox
}

func buildRepositories(ctx context.Context, cfg config.Config) (repositories, func() error, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.RepositoryMode)) {
	case "memory":
		store := memory.NewStore()
		return repositories{tenants: store, inventories: store, customAssetTypes: store, customFields: store, assets: store, search: store, attachments: store, blobs: store, audit: store, outbox: store}, func() error { return nil }, nil
	case "postgres":
		if strings.TrimSpace(cfg.DatabaseDSN) == "" {
			return repositories{}, nil, errors.New("database dsn is required")
		}
		store, closeStore, err := openPostgresStore(ctx, cfg.DatabaseDSN)
		if err != nil {
			return repositories{}, nil, err
		}
		blobs, err := buildBlobStorage(cfg)
		if err != nil {
			_ = closeStore()
			return repositories{}, nil, err
		}
		return repositories{tenants: store, inventories: store, customAssetTypes: store, customFields: store, assets: store, search: store, attachments: store, blobs: blobs, audit: store, outbox: store}, closeStore, nil
	default:
		return repositories{}, nil, errors.New("unsupported repository mode")
	}
}

func buildBlobStorage(cfg config.Config) (ports.BlobStorage, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.BlobStorageMode)) {
	case "", "filesystem":
		return blobstore.NewFileSystemStore(cfg.BlobStoragePath), nil
	case "s3":
		return blobstore.NewS3Store(blobstore.S3Config{
			Endpoint:  cfg.S3Endpoint,
			AccessKey: cfg.S3AccessKey,
			SecretKey: cfg.S3SecretKey,
			Bucket:    cfg.S3Bucket,
			Region:    cfg.S3Region,
			Secure:    cfg.S3Secure,
			MaxBytes:  int64(cfg.MaxAttachmentBytes),
		})
	default:
		return nil, errors.New("unsupported blob storage mode")
	}
}

func openPostgresStore(ctx context.Context, dsn string) (gormstore.Store, func() error, error) {
	db, closeStore, err := openPostgresDB(ctx, dsn)
	if err != nil {
		return gormstore.Store{}, nil, err
	}
	if err := verifyPostgresSchemaCurrent(db); err != nil {
		_ = closeStore()
		return gormstore.Store{}, nil, err
	}
	return gormstore.NewStore(db), closeStore, nil
}

func openPostgresDB(ctx context.Context, dsn string) (*gorm.DB, func() error, error) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.NewTimer(30 * time.Second)
	defer deadline.Stop()

	var lastErr error
	for {
		db, closeStore, err := tryOpenPostgresDB(ctx, dsn)
		if err == nil {
			return db, closeStore, nil
		}
		lastErr = err

		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-deadline.C:
			return nil, nil, fmt.Errorf("open postgres store: %w", lastErr)
		case <-ticker.C:
		}
	}
}

func tryOpenPostgresDB(ctx context.Context, dsn string) (*gorm.DB, func() error, error) {
	db, err := gormstore.OpenPostgres(dsn)
	if err != nil {
		return nil, nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	closeStore := sqlDB.Close

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = closeStore()
		return nil, nil, err
	}

	return db, closeStore, nil
}

func runMigrationCommand(ctx context.Context, cfg config.Config, args []string, output io.Writer) error {
	if len(args) != 1 {
		return errors.New("migration command must be one of: up, status")
	}
	switch args[0] {
	case "up", "status":
	default:
		return errors.New("migration command must be one of: up, status")
	}
	if strings.TrimSpace(cfg.DatabaseDSN) == "" {
		return errors.New("database dsn is required")
	}

	db, closeDB, err := openPostgresDB(ctx, cfg.DatabaseDSN)
	if err != nil {
		return err
	}
	defer func() {
		_ = closeDB()
	}()

	runner := dbmigrations.NewRunner(db)
	switch args[0] {
	case "up":
		if err := runner.Up(); err != nil {
			return err
		}
		_, _ = fmt.Fprintln(output, "migrations applied")
	case "status":
		status, err := runner.Status()
		if err != nil {
			return err
		}
		if status.Empty {
			_, _ = fmt.Fprintln(output, "migration version: none")
			return nil
		}
		_, _ = fmt.Fprintf(output, "migration version: %d latest: %d dirty: %t\n", status.Version, status.Latest, status.Dirty)
	}
	return nil
}

func verifyPostgresSchemaCurrent(db *gorm.DB) error {
	status, err := dbmigrations.NewRunner(db).Status()
	if err != nil {
		return err
	}
	return validateMigrationStatus(status)
}

func validateMigrationStatus(status dbmigrations.Status) error {
	if status.Empty {
		return errors.New("database migrations have not been applied")
	}
	if status.Dirty {
		return fmt.Errorf("database migrations are dirty at version %d", status.Version)
	}
	if status.Version != status.Latest {
		return fmt.Errorf("database migration version %d does not match latest %d", status.Version, status.Latest)
	}
	return nil
}

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
