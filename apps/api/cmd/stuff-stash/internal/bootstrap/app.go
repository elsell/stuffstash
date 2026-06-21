package bootstrap

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/idgen"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func buildApplication(cfg config.Config, observer ports.Observer, authenticator ports.Authenticator, authorizer ports.Authorizer, repositories repositories) app.App {
	return app.New(app.Dependencies{
		Observer:                      observer,
		Auth:                          authenticator,
		Authorizer:                    authorizer,
		Tenants:                       repositories.tenants,
		TenantUnitOfWork:              repositories.tenantUnitOfWork,
		Inventories:                   repositories.inventories,
		InventoryUnitOfWork:           repositories.inventoryUnitOfWork,
		InventoryAccess:               repositories.inventoryAccess,
		InventoryAccessUnitOfWork:     repositories.inventoryAccessUnitOfWork,
		CustomAssetTypes:              repositories.customAssetTypes,
		CustomAssetTypeUnitOfWork:     repositories.customAssetTypeUnitOfWork,
		CustomFields:                  repositories.customFields,
		CustomFieldUnitOfWork:         repositories.customFieldUnitOfWork,
		Assets:                        repositories.assets,
		AssetUnitOfWork:               repositories.assetUnitOfWork,
		Undoables:                     repositories.undoables,
		Search:                        repositories.search,
		Attachments:                   repositories.attachments,
		AttachmentUnitOfWork:          repositories.attachmentUnitOfWork,
		Blobs:                         repositories.blobs,
		BlobDeletionOutbox:            repositories.blobDeletionOutbox,
		Audit:                         repositories.audit,
		Outbox:                        repositories.outbox,
		IDs:                           idgen.NewULIDGenerator(),
		AuthorizationOutboxDrainLimit: cfg.AuthorizationOutboxDrainLimit,
		AuthorizationOutboxClaimLease: cfg.AuthorizationOutboxClaimLease,
		BlobDeletionOutboxClaimLease:  cfg.BlobDeletionOutboxClaimLease,
		BlobDeletionOutboxMaxAttempts: cfg.BlobDeletionOutboxMaxAttempts,
		InvitationTTL:                 cfg.InvitationTTL,
		DefaultPageLimit:              cfg.DefaultPageLimit,
		MaxPageLimit:                  cfg.MaxPageLimit,
		MaxAttachmentBytes:            cfg.MaxAttachmentBytes,
	})
}
