package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/blobstore"
	"github.com/stuffstash/stuff-stash/internal/adapters/gormstore"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
)

type repositories struct {
	tenants                    ports.TenantRepository
	tenantUnitOfWork           ports.TenantUnitOfWork
	inventories                ports.InventoryRepository
	inventoryUnitOfWork        ports.InventoryUnitOfWork
	inventoryAccess            ports.InventoryAccessRepository
	inventoryAccessUnitOfWork  ports.InventoryAccessUnitOfWork
	customAssetTypes           ports.CustomAssetTypeRepository
	customAssetTypeUnitOfWork  ports.CustomAssetTypeUnitOfWork
	customFields               ports.CustomFieldDefinitionRepository
	customFieldUnitOfWork      ports.CustomFieldDefinitionUnitOfWork
	assets                     ports.AssetRepository
	checkouts                  ports.AssetCheckoutRepository
	assetTags                  ports.AssetTagRepository
	assetUnitOfWork            ports.AssetUnitOfWork
	assetTagUnitOfWork         ports.AssetTagUnitOfWork
	undoables                  ports.UndoableOperationRepository
	search                     ports.AssetSearchRepository
	attachments                ports.AttachmentRepository
	attachmentUnitOfWork       ports.AttachmentUnitOfWork
	blobs                      ports.BlobStorage
	blobDeletionOutbox         ports.BlobDeletionOutbox
	directUploads              ports.DirectAttachmentUploader
	imageProcessor             ports.ImageProcessor
	audit                      ports.AuditRepository
	outbox                     ports.AuthorizationOutbox
	providerProfiles           ports.ProviderProfileRepository
	providerProfileUnitOfWork  ports.ProviderProfileUnitOfWork
	voiceProviderConfigs       ports.VoiceProviderConfigurationRepository
	providerCredentials        ports.ProviderCredentialRepository
	realtimeSessions           ports.RealtimeSessionRepository
	actionPlans                ports.ActionPlanRepository
	importJobs                 ports.ImportJobRepository
	importJobSources           ports.ImportJobSourceRepository
	importLinks                ports.ImportLinkRepository
	importAssetUnitOfWork      ports.ImportAssetUnitOfWork
	importAttachmentUnitOfWork ports.ImportAttachmentUnitOfWork
	users                      ports.UserRepository
}

func buildRepositories(ctx context.Context, cfg config.Config) (repositories, func() error, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.RepositoryMode)) {
	case "memory":
		store := memory.NewStore()
		return repositories{tenants: store, tenantUnitOfWork: store, inventories: store, inventoryUnitOfWork: store, inventoryAccess: store, inventoryAccessUnitOfWork: store, customAssetTypes: store, customAssetTypeUnitOfWork: store, customFields: store, customFieldUnitOfWork: store, assets: store, checkouts: store, assetTags: store, assetUnitOfWork: store, assetTagUnitOfWork: store, undoables: store, search: store, attachments: store, attachmentUnitOfWork: store, blobs: store, blobDeletionOutbox: store, directUploads: blobstore.NewLocalDirectAttachmentUploader(store), imageProcessor: blobstore.StandardImageProcessor{}, audit: store, outbox: store, providerProfiles: store, providerProfileUnitOfWork: store, voiceProviderConfigs: store, providerCredentials: store, realtimeSessions: store, actionPlans: store, importJobs: store, importJobSources: store, importLinks: store, importAssetUnitOfWork: store, importAttachmentUnitOfWork: store, users: store}, func() error { return nil }, nil
	case "postgres":
		if strings.TrimSpace(cfg.DatabaseDSN) == "" {
			return repositories{}, nil, errors.New("database dsn is required")
		}
		store, closeStore, err := openPostgresStore(ctx, cfg.DatabaseDSN)
		if err != nil {
			return repositories{}, nil, err
		}
		return repositoriesFromGORMStore(cfg, store, closeStore)
	case "sqlite":
		if strings.TrimSpace(cfg.DatabaseDSN) == "" {
			return repositories{}, nil, errors.New("database dsn is required")
		}
		store, closeStore, err := openSQLiteStore(ctx, cfg.DatabaseDSN)
		if err != nil {
			return repositories{}, nil, err
		}
		return repositoriesFromGORMStore(cfg, store, closeStore)
	default:
		return repositories{}, nil, errors.New("unsupported repository mode")
	}
}

func repositoriesFromGORMStore(cfg config.Config, store gormstore.Store, closeStore func() error) (repositories, func() error, error) {
	blobs, directUploads, err := buildBlobStorage(cfg)
	if err != nil {
		_ = closeStore()
		return repositories{}, nil, err
	}
	return repositories{tenants: store, tenantUnitOfWork: store, inventories: store, inventoryAccess: store, inventoryAccessUnitOfWork: store, inventoryUnitOfWork: store, customAssetTypes: store, customAssetTypeUnitOfWork: store, customFields: store, customFieldUnitOfWork: store, assets: store, checkouts: store, assetTags: store, assetUnitOfWork: store, assetTagUnitOfWork: store, undoables: store, search: store, attachments: store, attachmentUnitOfWork: store, blobs: blobs, blobDeletionOutbox: store, directUploads: directUploads, imageProcessor: blobstore.StandardImageProcessor{}, audit: store, outbox: store, providerProfiles: store, providerProfileUnitOfWork: store, voiceProviderConfigs: store, providerCredentials: store, realtimeSessions: store, actionPlans: store, importJobs: store, importJobSources: store, importLinks: store, importAssetUnitOfWork: store, importAttachmentUnitOfWork: store, users: store}, closeStore, nil
}

func buildBlobStorage(cfg config.Config) (ports.BlobStorage, ports.DirectAttachmentUploader, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.BlobStorageMode)) {
	case "", "filesystem":
		store := blobstore.NewFileSystemStoreWithMaxBytes(cfg.BlobStoragePath, int64(cfg.MaxAttachmentBytes))
		return store, blobstore.NewLocalDirectAttachmentUploader(store), nil
	case "s3":
		store, err := blobstore.NewS3Store(blobstore.S3Config{
			Endpoint:       cfg.S3Endpoint,
			PublicEndpoint: cfg.S3PublicEndpoint,
			AccessKey:      cfg.S3AccessKey,
			SecretKey:      cfg.S3SecretKey,
			Bucket:         cfg.S3Bucket,
			Region:         cfg.S3Region,
			Secure:         cfg.S3Secure,
			MaxBytes:       int64(cfg.MaxAttachmentBytes),
		})
		if err != nil {
			return nil, nil, err
		}
		return store, blobstore.NewS3DirectAttachmentUploader(store), nil
	default:
		return nil, nil, errors.New("unsupported blob storage mode")
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

func openSQLiteStore(ctx context.Context, dsn string) (gormstore.Store, func() error, error) {
	if err := ensureSQLiteParentDir(dsn); err != nil {
		return gormstore.Store{}, nil, err
	}
	db, err := gormstore.OpenSQLite(dsn)
	if err != nil {
		return gormstore.Store{}, nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return gormstore.Store{}, nil, err
	}
	closeStore := sqlDB.Close
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = closeStore()
		return gormstore.Store{}, nil, err
	}
	if err := gormstore.Migrate(ctx, db); err != nil {
		_ = closeStore()
		return gormstore.Store{}, nil, err
	}
	return gormstore.NewStore(db), closeStore, nil
}

func ensureSQLiteParentDir(dsn string) error {
	path := sqliteFilePath(dsn)
	if path == "" {
		return nil
	}
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o700)
}

func sqliteFilePath(dsn string) string {
	trimmed := strings.TrimSpace(dsn)
	if trimmed == "" || trimmed == ":memory:" {
		return ""
	}
	if strings.HasPrefix(trimmed, "file:") {
		trimmed = strings.TrimPrefix(trimmed, "file:")
		if queryIndex := strings.Index(trimmed, "?"); queryIndex >= 0 {
			trimmed = trimmed[:queryIndex]
		}
	}
	if trimmed == "" || trimmed == ":memory:" {
		return ""
	}
	return trimmed
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
