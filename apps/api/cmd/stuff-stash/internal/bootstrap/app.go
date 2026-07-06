package bootstrap

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/internal/adapters/homebox"
	"github.com/stuffstash/stuff-stash/internal/adapters/idgen"
	"github.com/stuffstash/stuff-stash/internal/adapters/importworker"
	"github.com/stuffstash/stuff-stash/internal/adapters/voice"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func buildApplication(ctx context.Context, cfg config.Config, observer ports.Observer, authenticator ports.Authenticator, authorizer ports.Authorizer, repositories repositories) (app.App, error) {
	if err := validateProviderCredentialSealer(ctx, cfg, repositories.providerCredentials); err != nil {
		return app.App{}, err
	}
	providerCredentialSealer, err := buildProviderCredentialSealer(cfg)
	if err != nil {
		return app.App{}, err
	}
	providerCredentialVault := buildProviderCredentialVault(repositories.providerCredentials, providerCredentialSealer)
	importSourceVault := buildImportJobSourceVault(repositories.importJobSources, providerCredentialSealer)
	stt, languageInference, tts, err := buildRealtimeVoiceProviders(ctx, cfg)
	if err != nil {
		return app.App{}, err
	}
	realtimeVoiceProviderResolver := buildRealtimeVoiceProviderResolver(cfg, repositories, providerCredentialVault, stt, languageInference, tts)
	application := app.New(app.Dependencies{
		Observer:                        observer,
		Auth:                            authenticator,
		Authorizer:                      authorizer,
		Tenants:                         repositories.tenants,
		TenantUnitOfWork:                repositories.tenantUnitOfWork,
		Inventories:                     repositories.inventories,
		InventoryUnitOfWork:             repositories.inventoryUnitOfWork,
		InventoryAccess:                 repositories.inventoryAccess,
		InventoryAccessUnitOfWork:       repositories.inventoryAccessUnitOfWork,
		CustomAssetTypes:                repositories.customAssetTypes,
		CustomAssetTypeUnitOfWork:       repositories.customAssetTypeUnitOfWork,
		CustomFields:                    repositories.customFields,
		CustomFieldUnitOfWork:           repositories.customFieldUnitOfWork,
		Assets:                          repositories.assets,
		AssetUnitOfWork:                 repositories.assetUnitOfWork,
		Undoables:                       repositories.undoables,
		Search:                          repositories.search,
		Attachments:                     repositories.attachments,
		AttachmentUnitOfWork:            repositories.attachmentUnitOfWork,
		Blobs:                           repositories.blobs,
		DirectUploads:                   repositories.directUploads,
		ImageProcessor:                  repositories.imageProcessor,
		BlobDeletionOutbox:              repositories.blobDeletionOutbox,
		Audit:                           repositories.audit,
		Outbox:                          repositories.outbox,
		Users:                           repositories.users,
		ProviderProfiles:                repositories.providerProfiles,
		ProviderProfileUnitOfWork:       repositories.providerProfileUnitOfWork,
		VoiceProviderConfigs:            repositories.voiceProviderConfigs,
		ProviderCredentialVault:         providerCredentialVault,
		ProviderProfileTester:           voice.NewProviderProfileTester(googleProviderProfileFactory(cfg)),
		RealtimeSessions:                repositories.realtimeSessions,
		ActionPlans:                     repositories.actionPlans,
		ImportSources:                   homebox.NewLegacyImporter(nil),
		ImportJobs:                      repositories.importJobs,
		ImportSourceVault:               importSourceVault,
		ImportLinks:                     repositories.importLinks,
		ImportAssetUnitOfWork:           repositories.importAssetUnitOfWork,
		ImportAttachmentUnitOfWork:      repositories.importAttachmentUnitOfWork,
		IDs:                             idgen.NewULIDGenerator(),
		AuthorizationOutboxDrainLimit:   cfg.AuthorizationOutboxDrainLimit,
		AuthorizationOutboxClaimLease:   cfg.AuthorizationOutboxClaimLease,
		BlobDeletionOutboxClaimLease:    cfg.BlobDeletionOutboxClaimLease,
		BlobDeletionOutboxMaxAttempts:   cfg.BlobDeletionOutboxMaxAttempts,
		InvitationTTL:                   cfg.InvitationTTL,
		DefaultPageLimit:                cfg.DefaultPageLimit,
		MaxPageLimit:                    cfg.MaxPageLimit,
		MaxAttachmentBytes:              cfg.MaxAttachmentBytes,
		ImportJobTimeout:                cfg.ImportJobTimeout,
		PrimaryThumbnailWarmLimit:       cfg.PrimaryThumbnailWarmLimit,
		PrimaryThumbnailWarmConcurrency: cfg.PrimaryThumbnailWarmConcurrency,
		PrimaryThumbnailWarmTimeout:     cfg.PrimaryThumbnailWarmTimeout,
		SpeechToText:                    stt,
		LanguageInference:               languageInference,
		TextToSpeech:                    tts,
		RealtimeVoiceProviderResolver:   realtimeVoiceProviderResolver,
	})
	application = application.WithImportWorker(importworker.NewInProcess(application, observer))
	if _, err := application.ResumeRunningImportJobs(ctx, 25); err != nil {
		return app.App{}, err
	}
	return application, nil
}

func buildProviderCredentialVault(repository ports.ProviderCredentialRepository, sealer ports.ProviderCredentialSealer) ports.ProviderCredentialVault {
	if repository == nil || sealer == nil {
		return nil
	}
	return credentials.NewDatabaseProviderCredentialVault(repository, sealer)
}

func buildImportJobSourceVault(repository ports.ImportJobSourceRepository, sealer ports.ProviderCredentialSealer) ports.ImportJobSourceVault {
	if repository == nil || sealer == nil {
		return nil
	}
	importSealer, ok := sealer.(credentials.ImportJobSourceSealer)
	if !ok {
		return nil
	}
	return credentials.NewDatabaseImportJobSourceVault(repository, importSealer)
}
