package app

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
	customfieldapp "github.com/stuffstash/stuff-stash/internal/app/customfields"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type App struct {
	observer                  ports.Observer
	auth                      ports.Authenticator
	authorizer                ports.Authorizer
	tenants                   ports.TenantRepository
	tenantUnitOfWork          ports.TenantUnitOfWork
	inventories               ports.InventoryRepository
	inventoryUnitOfWork       ports.InventoryUnitOfWork
	inventoryAccess           ports.InventoryAccessRepository
	inventoryAccessUnitOfWork ports.InventoryAccessUnitOfWork
	customAssetTypes          ports.CustomAssetTypeRepository
	customAssetTypeUnitOfWork ports.CustomAssetTypeUnitOfWork
	customFields              ports.CustomFieldDefinitionRepository
	customFieldUnitOfWork     ports.CustomFieldDefinitionUnitOfWork
	assets                    ports.AssetRepository
	assetUnitOfWork           ports.AssetUnitOfWork
	undoables                 ports.UndoableOperationRepository
	search                    ports.AssetSearchRepository
	attachments               ports.AttachmentRepository
	attachmentUnitOfWork      ports.AttachmentUnitOfWork
	blobs                     ports.BlobStorage
	directUploads             ports.DirectAttachmentUploader
	imageProcessor            ports.ImageProcessor
	blobDeletionOutbox        ports.BlobDeletionOutbox
	audit                     ports.AuditRepository
	outbox                    ports.AuthorizationOutbox
	ids                       ports.IDGenerator
	clock                     ports.Clock
	outboxDrainLimit          int
	outboxClaimLease          time.Duration
	blobDeletionClaimLease    time.Duration
	blobDeletionMaxAttempts   int
	directUploadTTL           time.Duration
	invitationTTL             time.Duration
	defaultPageLimit          int
	maxPageLimit              int
	maxAttachmentBytes        int
	assetService              assetapp.Service
	customFieldService        customfieldapp.Service
	speechToText              ports.SpeechToTextProvider
	languageInference         ports.LanguageInferenceProvider
	textToSpeech              ports.TextToSpeechProvider
}

type Dependencies struct {
	Observer                      ports.Observer
	Auth                          ports.Authenticator
	Authorizer                    ports.Authorizer
	Tenants                       ports.TenantRepository
	TenantUnitOfWork              ports.TenantUnitOfWork
	Inventories                   ports.InventoryRepository
	InventoryUnitOfWork           ports.InventoryUnitOfWork
	InventoryAccess               ports.InventoryAccessRepository
	InventoryAccessUnitOfWork     ports.InventoryAccessUnitOfWork
	CustomAssetTypes              ports.CustomAssetTypeRepository
	CustomAssetTypeUnitOfWork     ports.CustomAssetTypeUnitOfWork
	CustomFields                  ports.CustomFieldDefinitionRepository
	CustomFieldUnitOfWork         ports.CustomFieldDefinitionUnitOfWork
	Assets                        ports.AssetRepository
	AssetUnitOfWork               ports.AssetUnitOfWork
	Undoables                     ports.UndoableOperationRepository
	Search                        ports.AssetSearchRepository
	Attachments                   ports.AttachmentRepository
	AttachmentUnitOfWork          ports.AttachmentUnitOfWork
	Blobs                         ports.BlobStorage
	DirectUploads                 ports.DirectAttachmentUploader
	ImageProcessor                ports.ImageProcessor
	BlobDeletionOutbox            ports.BlobDeletionOutbox
	Audit                         ports.AuditRepository
	Outbox                        ports.AuthorizationOutbox
	IDs                           ports.IDGenerator
	Clock                         ports.Clock
	AuthorizationOutboxDrainLimit int
	AuthorizationOutboxClaimLease time.Duration
	BlobDeletionOutboxClaimLease  time.Duration
	BlobDeletionOutboxMaxAttempts int
	DirectUploadTTL               time.Duration
	InvitationTTL                 time.Duration
	DefaultPageLimit              int
	MaxPageLimit                  int
	MaxAttachmentBytes            int
	SpeechToText                  ports.SpeechToTextProvider
	LanguageInference             ports.LanguageInferenceProvider
	TextToSpeech                  ports.TextToSpeechProvider
}

func New(deps Dependencies) App {
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
	app := App{
		observer:                  deps.Observer,
		auth:                      deps.Auth,
		authorizer:                deps.Authorizer,
		tenants:                   deps.Tenants,
		tenantUnitOfWork:          deps.TenantUnitOfWork,
		inventories:               deps.Inventories,
		inventoryUnitOfWork:       deps.InventoryUnitOfWork,
		inventoryAccess:           deps.InventoryAccess,
		inventoryAccessUnitOfWork: deps.InventoryAccessUnitOfWork,
		customAssetTypes:          deps.CustomAssetTypes,
		customAssetTypeUnitOfWork: deps.CustomAssetTypeUnitOfWork,
		customFields:              deps.CustomFields,
		customFieldUnitOfWork:     deps.CustomFieldUnitOfWork,
		assets:                    deps.Assets,
		assetUnitOfWork:           deps.AssetUnitOfWork,
		undoables:                 deps.Undoables,
		search:                    deps.Search,
		attachments:               deps.Attachments,
		attachmentUnitOfWork:      deps.AttachmentUnitOfWork,
		blobs:                     deps.Blobs,
		directUploads:             deps.DirectUploads,
		imageProcessor:            deps.ImageProcessor,
		blobDeletionOutbox:        deps.BlobDeletionOutbox,
		audit:                     deps.Audit,
		outbox:                    deps.Outbox,
		ids:                       ids,
		clock:                     clock,
		outboxDrainLimit:          deps.AuthorizationOutboxDrainLimit,
		outboxClaimLease:          deps.AuthorizationOutboxClaimLease,
		blobDeletionClaimLease:    normalizeOutboxClaimLease(deps.BlobDeletionOutboxClaimLease),
		blobDeletionMaxAttempts:   normalizeBlobDeletionMaxAttempts(deps.BlobDeletionOutboxMaxAttempts),
		directUploadTTL:           normalizeDirectUploadTTL(deps.DirectUploadTTL),
		invitationTTL:             normalizeInvitationTTL(deps.InvitationTTL),
		defaultPageLimit:          defaultPageLimit,
		maxPageLimit:              maxPageLimit,
		maxAttachmentBytes:        normalizeMaxAttachmentBytes(deps.MaxAttachmentBytes),
		speechToText:              deps.SpeechToText,
		languageInference:         deps.LanguageInference,
		textToSpeech:              deps.TextToSpeech,
	}
	app.assetService = assetapp.New(assetapp.Dependencies{
		Observer:         app.observer,
		Authorizer:       app.authorizer,
		Tenants:          app.tenants,
		Inventories:      app.inventories,
		CustomAssetTypes: app.customAssetTypes,
		CustomFields:     app.customFields,
		Assets:           app.assets,
		AssetUnitOfWork:  app.assetUnitOfWork,
		Undoables:        app.undoables,
		Audit:            app.audit,
		IDs:              app.ids,
		Clock:            app.clock,
		DefaultPageLimit: app.defaultPageLimit,
		MaxPageLimit:     app.maxPageLimit,
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
	return app
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
		return 5 * 1024 * 1024
	}
	return maxBytes
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

	return principal, nil
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
