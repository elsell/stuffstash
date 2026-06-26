package bootstrap

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/adapters/idgen"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func buildApplication(ctx context.Context, cfg config.Config, observer ports.Observer, authenticator ports.Authenticator, authorizer ports.Authorizer, repositories repositories) (app.App, error) {
	if err := validateProviderCredentialSealer(ctx, cfg, repositories.providerCredentials); err != nil {
		return app.App{}, err
	}
	stt, languageInference, tts, err := buildRealtimeVoiceProviders(ctx, cfg)
	if err != nil {
		return app.App{}, err
	}
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
		DirectUploads:                 repositories.directUploads,
		ImageProcessor:                repositories.imageProcessor,
		BlobDeletionOutbox:            repositories.blobDeletionOutbox,
		Audit:                         repositories.audit,
		Outbox:                        repositories.outbox,
		ProviderProfiles:              repositories.providerProfiles,
		ProviderProfileUnitOfWork:     repositories.providerProfileUnitOfWork,
		IDs:                           idgen.NewULIDGenerator(),
		AuthorizationOutboxDrainLimit: cfg.AuthorizationOutboxDrainLimit,
		AuthorizationOutboxClaimLease: cfg.AuthorizationOutboxClaimLease,
		BlobDeletionOutboxClaimLease:  cfg.BlobDeletionOutboxClaimLease,
		BlobDeletionOutboxMaxAttempts: cfg.BlobDeletionOutboxMaxAttempts,
		InvitationTTL:                 cfg.InvitationTTL,
		DefaultPageLimit:              cfg.DefaultPageLimit,
		MaxPageLimit:                  cfg.MaxPageLimit,
		MaxAttachmentBytes:            cfg.MaxAttachmentBytes,
		SpeechToText:                  stt,
		LanguageInference:             languageInference,
		TextToSpeech:                  tts,
	}), nil
}
