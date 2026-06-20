package app

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type App struct {
	observer           ports.Observer
	auth               ports.Authenticator
	authorizer         ports.Authorizer
	tenants            ports.TenantRepository
	inventories        ports.InventoryRepository
	inventoryAccess    ports.InventoryAccessRepository
	customAssetTypes   ports.CustomAssetTypeRepository
	customFields       ports.CustomFieldDefinitionRepository
	assets             ports.AssetRepository
	undoables          ports.UndoableOperationRepository
	search             ports.AssetSearchRepository
	attachments        ports.AttachmentRepository
	blobs              ports.BlobStorage
	audit              ports.AuditRepository
	outbox             ports.AuthorizationOutbox
	ids                ports.IDGenerator
	outboxDrainLimit   int
	outboxClaimLease   time.Duration
	invitationTTL      time.Duration
	defaultPageLimit   int
	maxPageLimit       int
	maxAttachmentBytes int
}

type Dependencies struct {
	Observer                      ports.Observer
	Auth                          ports.Authenticator
	Authorizer                    ports.Authorizer
	Tenants                       ports.TenantRepository
	Inventories                   ports.InventoryRepository
	InventoryAccess               ports.InventoryAccessRepository
	CustomAssetTypes              ports.CustomAssetTypeRepository
	CustomFields                  ports.CustomFieldDefinitionRepository
	Assets                        ports.AssetRepository
	Undoables                     ports.UndoableOperationRepository
	Search                        ports.AssetSearchRepository
	Attachments                   ports.AttachmentRepository
	Blobs                         ports.BlobStorage
	Audit                         ports.AuditRepository
	Outbox                        ports.AuthorizationOutbox
	IDs                           ports.IDGenerator
	AuthorizationOutboxDrainLimit int
	AuthorizationOutboxClaimLease time.Duration
	InvitationTTL                 time.Duration
	DefaultPageLimit              int
	MaxPageLimit                  int
	MaxAttachmentBytes            int
}

func New(deps Dependencies) App {
	maxPageLimit := normalizeMaxPageLimit(deps.MaxPageLimit)
	ids := deps.IDs
	if ids == nil {
		ids = defaultIDGenerator{}
	}
	return App{
		observer:           deps.Observer,
		auth:               deps.Auth,
		authorizer:         deps.Authorizer,
		tenants:            deps.Tenants,
		inventories:        deps.Inventories,
		inventoryAccess:    deps.InventoryAccess,
		customAssetTypes:   deps.CustomAssetTypes,
		customFields:       deps.CustomFields,
		assets:             deps.Assets,
		undoables:          deps.Undoables,
		search:             deps.Search,
		attachments:        deps.Attachments,
		blobs:              deps.Blobs,
		audit:              deps.Audit,
		outbox:             deps.Outbox,
		ids:                ids,
		outboxDrainLimit:   deps.AuthorizationOutboxDrainLimit,
		outboxClaimLease:   deps.AuthorizationOutboxClaimLease,
		invitationTTL:      normalizeInvitationTTL(deps.InvitationTTL),
		defaultPageLimit:   normalizeDefaultPageLimit(deps.DefaultPageLimit, maxPageLimit),
		maxPageLimit:       maxPageLimit,
		maxAttachmentBytes: normalizeMaxAttachmentBytes(deps.MaxAttachmentBytes),
	}
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

func normalizeDefaultPageLimit(defaultLimit int, maxLimit int) int {
	if defaultLimit <= 0 {
		return 50
	}
	if defaultLimit > maxLimit {
		return maxLimit
	}
	return defaultLimit
}

func normalizeMaxPageLimit(maxLimit int) int {
	if maxLimit <= 0 {
		return 100
	}
	return maxLimit
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
