package app

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
	agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
	customfieldapp "github.com/stuffstash/stuff-stash/internal/app/customfields"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type App struct {
	observer                     ports.Observer
	auth                         ports.Authenticator
	authorizer                   ports.Authorizer
	users                        ports.UserRepository
	tenants                      ports.TenantRepository
	tenantUnitOfWork             ports.TenantUnitOfWork
	inventories                  ports.InventoryRepository
	inventoryUnitOfWork          ports.InventoryUnitOfWork
	inventoryAccess              ports.InventoryAccessRepository
	inventoryAccessUnitOfWork    ports.InventoryAccessUnitOfWork
	customAssetTypes             ports.CustomAssetTypeRepository
	customAssetTypeUnitOfWork    ports.CustomAssetTypeUnitOfWork
	customFields                 ports.CustomFieldDefinitionRepository
	customFieldUnitOfWork        ports.CustomFieldDefinitionUnitOfWork
	assets                       ports.AssetRepository
	assetTags                    ports.AssetTagRepository
	checkouts                    ports.AssetCheckoutRepository
	assetUnitOfWork              ports.AssetUnitOfWork
	assetTagUnitOfWork           ports.AssetTagUnitOfWork
	assetEditUnitOfWork          ports.AssetEditUnitOfWork
	undoables                    ports.UndoableOperationRepository
	search                       ports.AssetSearchRepository
	attachments                  ports.AttachmentRepository
	attachmentUnitOfWork         ports.AttachmentUnitOfWork
	blobs                        ports.BlobStorage
	directUploads                ports.DirectAttachmentUploader
	imageProcessor               ports.ImageProcessor
	blobDeletionOutbox           ports.BlobDeletionOutbox
	audit                        ports.AuditRepository
	outbox                       ports.AuthorizationOutbox
	providerProfiles             ports.ProviderProfileRepository
	providerProfileUnitOfWork    ports.ProviderProfileUnitOfWork
	voiceProviderConfigs         ports.VoiceProviderConfigurationRepository
	providerCredentialVault      ports.ProviderCredentialVault
	providerProfileTester        ports.ProviderProfileTester
	realtimeSessions             ports.RealtimeSessionRepository
	actionPlans                  ports.ActionPlanRepository
	importSources                ports.ImportSourceReader
	importJobs                   ports.ImportJobRepository
	importSourceVault            ports.ImportJobSourceVault
	importLinks                  ports.ImportLinkRepository
	importAssetUnitOfWork        ports.ImportAssetUnitOfWork
	importAttachmentUnitOfWork   ports.ImportAttachmentUnitOfWork
	importWorker                 ports.ImportWorker
	ids                          ports.IDGenerator
	clock                        ports.Clock
	outboxDrainLimit             int
	outboxClaimLease             time.Duration
	blobDeletionClaimLease       time.Duration
	blobDeletionMaxAttempts      int
	directUploadTTL              time.Duration
	invitationTTL                time.Duration
	invitationPublicBaseURL      string
	invitationAllowInsecureHTTP  bool
	defaultPageLimit             int
	maxPageLimit                 int
	maxAttachmentBytes           int
	importJobTimeout             time.Duration
	primaryThumbnailWarmLimit    int
	primaryThumbnailWarmTimeout  time.Duration
	realtimeVoiceToolCallTimeout time.Duration
	assetService                 assetapp.Service
	customFieldService           customfieldapp.Service
	providerProfileService       agentmodelapp.Service
	speechToText                 ports.SpeechToTextProvider
	languageInference            ports.LanguageInferenceProvider
	voiceResponseGenerator       ports.VoiceResponseGenerator
	textToSpeech                 ports.TextToSpeechProvider
	realtimeVoiceProviders       ports.RealtimeVoiceProviderResolver
	thumbnailWarmState           *primaryThumbnailWarmState
	thumbnailGenerationState     *thumbnailGenerationState
}

type Dependencies struct {
	Observer                         ports.Observer
	Auth                             ports.Authenticator
	Authorizer                       ports.Authorizer
	Users                            ports.UserRepository
	Tenants                          ports.TenantRepository
	TenantUnitOfWork                 ports.TenantUnitOfWork
	Inventories                      ports.InventoryRepository
	InventoryUnitOfWork              ports.InventoryUnitOfWork
	InventoryAccess                  ports.InventoryAccessRepository
	InventoryAccessUnitOfWork        ports.InventoryAccessUnitOfWork
	CustomAssetTypes                 ports.CustomAssetTypeRepository
	CustomAssetTypeUnitOfWork        ports.CustomAssetTypeUnitOfWork
	CustomFields                     ports.CustomFieldDefinitionRepository
	CustomFieldUnitOfWork            ports.CustomFieldDefinitionUnitOfWork
	Assets                           ports.AssetRepository
	AssetTags                        ports.AssetTagRepository
	Checkouts                        ports.AssetCheckoutRepository
	AssetUnitOfWork                  ports.AssetUnitOfWork
	AssetTagUnitOfWork               ports.AssetTagUnitOfWork
	AssetEditUnitOfWork              ports.AssetEditUnitOfWork
	Undoables                        ports.UndoableOperationRepository
	Search                           ports.AssetSearchRepository
	Attachments                      ports.AttachmentRepository
	AttachmentUnitOfWork             ports.AttachmentUnitOfWork
	Blobs                            ports.BlobStorage
	DirectUploads                    ports.DirectAttachmentUploader
	ImageProcessor                   ports.ImageProcessor
	BlobDeletionOutbox               ports.BlobDeletionOutbox
	Audit                            ports.AuditRepository
	Outbox                           ports.AuthorizationOutbox
	ProviderProfiles                 ports.ProviderProfileRepository
	ProviderProfileUnitOfWork        ports.ProviderProfileUnitOfWork
	VoiceProviderConfigs             ports.VoiceProviderConfigurationRepository
	ProviderCredentialVault          ports.ProviderCredentialVault
	ProviderProfileTester            ports.ProviderProfileTester
	RealtimeSessions                 ports.RealtimeSessionRepository
	ActionPlans                      ports.ActionPlanRepository
	ImportSources                    ports.ImportSourceReader
	ImportJobs                       ports.ImportJobRepository
	ImportSourceVault                ports.ImportJobSourceVault
	ImportLinks                      ports.ImportLinkRepository
	ImportAssetUnitOfWork            ports.ImportAssetUnitOfWork
	ImportAttachmentUnitOfWork       ports.ImportAttachmentUnitOfWork
	ImportWorker                     ports.ImportWorker
	IDs                              ports.IDGenerator
	Clock                            ports.Clock
	AuthorizationOutboxDrainLimit    int
	AuthorizationOutboxClaimLease    time.Duration
	BlobDeletionOutboxClaimLease     time.Duration
	BlobDeletionOutboxMaxAttempts    int
	DirectUploadTTL                  time.Duration
	InvitationTTL                    time.Duration
	InvitationPublicBaseURL          string
	InvitationAllowInsecureLocalHTTP bool
	DefaultPageLimit                 int
	MaxPageLimit                     int
	MaxAttachmentBytes               int
	ImportJobTimeout                 time.Duration
	PrimaryThumbnailWarmLimit        int
	PrimaryThumbnailWarmConcurrency  int
	PrimaryThumbnailWarmTimeout      time.Duration
	RealtimeVoiceToolCallTimeout     time.Duration
	SpeechToText                     ports.SpeechToTextProvider
	LanguageInference                ports.LanguageInferenceProvider
	VoiceResponseGenerator           ports.VoiceResponseGenerator
	TextToSpeech                     ports.TextToSpeechProvider
	RealtimeVoiceProviderResolver    ports.RealtimeVoiceProviderResolver
}

func New(deps Dependencies) App {
	observer := deps.Observer
	if observer == nil {
		observer = appNoopObserver{}
	}
	maxPageLimit := normalizeMaxPageLimit(deps.MaxPageLimit)
	ids := deps.IDs
	if ids == nil {
		ids = defaultIDGenerator{}
	}
	clock := deps.Clock
	if clock == nil {
		clock = ports.SystemClock{}
	}
	defaultPageLimit := normalizeDefaultPageLimit(deps.DefaultPageLimit, maxPageLimit)
	primaryThumbnailWarmConcurrency := normalizePrimaryThumbnailWarmConcurrency(deps.PrimaryThumbnailWarmConcurrency)
	realtimeVoiceProviders := deps.RealtimeVoiceProviderResolver
	if realtimeVoiceProviders == nil && deps.SpeechToText != nil && deps.LanguageInference != nil && deps.VoiceResponseGenerator != nil && deps.TextToSpeech != nil {
		realtimeVoiceProviders = staticRealtimeVoiceProviderResolver{providers: ports.RealtimeVoiceProviderSet{
			SpeechToText:      deps.SpeechToText,
			LanguageInference: deps.LanguageInference,
			ResponseGenerator: deps.VoiceResponseGenerator,
			TextToSpeech:      deps.TextToSpeech,
		}}
	}
	app := App{
		observer:                     observer,
		auth:                         deps.Auth,
		authorizer:                   deps.Authorizer,
		users:                        deps.Users,
		tenants:                      deps.Tenants,
		tenantUnitOfWork:             deps.TenantUnitOfWork,
		inventories:                  deps.Inventories,
		inventoryUnitOfWork:          deps.InventoryUnitOfWork,
		inventoryAccess:              deps.InventoryAccess,
		inventoryAccessUnitOfWork:    deps.InventoryAccessUnitOfWork,
		customAssetTypes:             deps.CustomAssetTypes,
		customAssetTypeUnitOfWork:    deps.CustomAssetTypeUnitOfWork,
		customFields:                 deps.CustomFields,
		customFieldUnitOfWork:        deps.CustomFieldUnitOfWork,
		assets:                       deps.Assets,
		assetTags:                    deps.AssetTags,
		checkouts:                    deps.Checkouts,
		assetUnitOfWork:              deps.AssetUnitOfWork,
		assetTagUnitOfWork:           deps.AssetTagUnitOfWork,
		assetEditUnitOfWork:          deps.AssetEditUnitOfWork,
		undoables:                    deps.Undoables,
		search:                       deps.Search,
		attachments:                  deps.Attachments,
		attachmentUnitOfWork:         deps.AttachmentUnitOfWork,
		blobs:                        deps.Blobs,
		directUploads:                deps.DirectUploads,
		imageProcessor:               deps.ImageProcessor,
		blobDeletionOutbox:           deps.BlobDeletionOutbox,
		audit:                        deps.Audit,
		outbox:                       deps.Outbox,
		providerProfiles:             deps.ProviderProfiles,
		providerProfileUnitOfWork:    deps.ProviderProfileUnitOfWork,
		voiceProviderConfigs:         deps.VoiceProviderConfigs,
		providerCredentialVault:      deps.ProviderCredentialVault,
		providerProfileTester:        deps.ProviderProfileTester,
		realtimeSessions:             deps.RealtimeSessions,
		actionPlans:                  deps.ActionPlans,
		importSources:                deps.ImportSources,
		importJobs:                   deps.ImportJobs,
		importSourceVault:            deps.ImportSourceVault,
		importLinks:                  deps.ImportLinks,
		importAssetUnitOfWork:        deps.ImportAssetUnitOfWork,
		importAttachmentUnitOfWork:   deps.ImportAttachmentUnitOfWork,
		importWorker:                 deps.ImportWorker,
		ids:                          ids,
		clock:                        clock,
		outboxDrainLimit:             deps.AuthorizationOutboxDrainLimit,
		outboxClaimLease:             deps.AuthorizationOutboxClaimLease,
		blobDeletionClaimLease:       normalizeOutboxClaimLease(deps.BlobDeletionOutboxClaimLease),
		blobDeletionMaxAttempts:      normalizeBlobDeletionMaxAttempts(deps.BlobDeletionOutboxMaxAttempts),
		directUploadTTL:              normalizeDirectUploadTTL(deps.DirectUploadTTL),
		invitationTTL:                normalizeInvitationTTL(deps.InvitationTTL),
		invitationPublicBaseURL:      normalizeInvitationPublicBaseURL(deps.InvitationPublicBaseURL),
		invitationAllowInsecureHTTP:  deps.InvitationAllowInsecureLocalHTTP,
		defaultPageLimit:             defaultPageLimit,
		maxPageLimit:                 maxPageLimit,
		maxAttachmentBytes:           normalizeMaxAttachmentBytes(deps.MaxAttachmentBytes),
		importJobTimeout:             normalizeImportJobTimeout(deps.ImportJobTimeout),
		primaryThumbnailWarmLimit:    normalizePrimaryThumbnailWarmLimit(deps.PrimaryThumbnailWarmLimit),
		primaryThumbnailWarmTimeout:  normalizePrimaryThumbnailWarmTimeout(deps.PrimaryThumbnailWarmTimeout),
		realtimeVoiceToolCallTimeout: normalizeRealtimeVoiceToolCallTimeout(deps.RealtimeVoiceToolCallTimeout),
		speechToText:                 deps.SpeechToText,
		languageInference:            deps.LanguageInference,
		voiceResponseGenerator:       deps.VoiceResponseGenerator,
		textToSpeech:                 deps.TextToSpeech,
		realtimeVoiceProviders:       realtimeVoiceProviders,
		thumbnailWarmState:           newPrimaryThumbnailWarmState(primaryThumbnailWarmConcurrency),
		thumbnailGenerationState:     newThumbnailGenerationState(),
	}
	app.assetService = assetapp.New(assetapp.Dependencies{
		Observer:            app.observer,
		Authorizer:          app.authorizer,
		Tenants:             app.tenants,
		Inventories:         app.inventories,
		CustomAssetTypes:    app.customAssetTypes,
		CustomFields:        app.customFields,
		Assets:              app.assets,
		Checkouts:           app.checkouts,
		Attachments:         app.attachments,
		AssetTags:           app.assetTags,
		AssetUnitOfWork:     app.assetUnitOfWork,
		AssetTagUnitOfWork:  app.assetTagUnitOfWork,
		AssetEditUnitOfWork: app.assetEditUnitOfWork,
		Undoables:           app.undoables,
		Audit:               app.audit,
		IDs:                 app.ids,
		Clock:               app.clock,
		DefaultPageLimit:    app.defaultPageLimit,
		MaxPageLimit:        app.maxPageLimit,
	})
	app.customFieldService = customfieldapp.New(customfieldapp.Dependencies{
		Observer:                  app.observer,
		Authorizer:                app.authorizer,
		Tenants:                   app.tenants,
		Inventories:               app.inventories,
		CustomAssetTypes:          app.customAssetTypes,
		CustomAssetTypeUnitOfWork: app.customAssetTypeUnitOfWork,
		CustomFields:              app.customFields,
		CustomFieldUnitOfWork:     app.customFieldUnitOfWork,
		Audit:                     app.audit,
		IDs:                       app.ids,
		Clock:                     app.clock,
		DefaultPageLimit:          app.defaultPageLimit,
		MaxPageLimit:              app.maxPageLimit,
	})
	app.providerProfileService = agentmodelapp.New(agentmodelapp.Dependencies{
		Observer:                  app.observer,
		Authorizer:                app.authorizer,
		ProviderProfiles:          app.providerProfiles,
		ProviderProfileUnitOfWork: app.providerProfileUnitOfWork,
		VoiceProviderConfigs:      app.voiceProviderConfigs,
		ProviderCredentialVault:   app.providerCredentialVault,
		ProviderProfileTester:     app.providerProfileTester,
		IDs:                       app.ids,
		Clock:                     app.clock,
	})
	return app
}

func (a App) WithImportWorker(worker ports.ImportWorker) App {
	a.importWorker = worker
	return a
}

type defaultIDGenerator struct{}

func (defaultIDGenerator) NewID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now().UTC()), rand.Reader).String()
}

func (a App) MaxAttachmentJSONBodyBytes() int64 {
	return int64(a.maxAttachmentBytes*2 + 4096)
}

func normalizeMaxAttachmentBytes(maxBytes int) int {
	if maxBytes <= 0 {
		return 25 * 1024 * 1024
	}
	return maxBytes
}

func normalizeImportJobTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return 15 * time.Minute
	}
	return timeout
}

func normalizePrimaryThumbnailWarmLimit(limit int) int {
	if limit <= 0 {
		return defaultPrimarySmallThumbnailWarmLimit
	}
	return limit
}

func normalizePrimaryThumbnailWarmConcurrency(concurrency int) int {
	if concurrency <= 0 {
		return defaultPrimarySmallThumbnailWarmConcurrency
	}
	return concurrency
}

func normalizePrimaryThumbnailWarmTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return defaultPrimarySmallThumbnailWarmTimeout
	}
	return timeout
}

func normalizeRealtimeVoiceToolCallTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return 10 * time.Second
	}
	return timeout
}

func normalizeInvitationTTL(ttl time.Duration) time.Duration {
	if ttl <= 0 {
		return 7 * 24 * time.Hour
	}
	return ttl
}

func normalizeDirectUploadTTL(ttl time.Duration) time.Duration {
	if ttl <= 0 {
		return 15 * time.Minute
	}
	return ttl
}

func normalizeOutboxClaimLease(lease time.Duration) time.Duration {
	if lease <= 0 {
		return 30 * time.Second
	}
	return lease
}

func normalizeBlobDeletionMaxAttempts(attempts int) int {
	if attempts <= 0 {
		return 5
	}
	return attempts
}

func normalizeDefaultPageLimit(defaultLimit int, maxLimit int) int {
	return appsupport.NormalizeDefaultPageLimit(defaultLimit, maxLimit)
}

func normalizeMaxPageLimit(maxLimit int) int {
	return appsupport.NormalizeMaxPageLimit(maxLimit)
}

func (a App) Authenticate(ctx context.Context, authorizationHeader string) (identity.Principal, error) {
	principal, err := a.auth.Authenticate(ctx, authorizationHeader)
	if err != nil {
		a.observer.Record(ctx, ports.Event{
			Name:    ports.EventAuthenticationFailed,
			Message: "authentication failed",
		})
		return identity.Principal{}, err
	}
	if err := a.saveAuthenticatedUser(ctx, principal); err != nil {
		return identity.Principal{}, err
	}

	return principal, nil
}

func (a App) saveAuthenticatedUser(ctx context.Context, principal identity.Principal) error {
	if a.users == nil {
		return nil
	}
	user, ok := identity.NewUser(principal.ID, principal.Email)
	if !ok {
		return nil
	}
	return a.users.SaveUser(ctx, user)
}

func (a App) Health(ctx context.Context) HealthStatus {
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventHealthChecked,
		Message: "health check completed",
	})

	return HealthStatus{
		Service: ServiceNameStuffStash,
		Status:  HealthStatusHealthy,
	}
}
