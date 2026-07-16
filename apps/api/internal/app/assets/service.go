package assets

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type Service struct {
	observer            ports.Observer
	authorizer          ports.Authorizer
	tenants             ports.TenantRepository
	inventories         ports.InventoryRepository
	customAssetTypes    ports.CustomAssetTypeRepository
	customFields        ports.CustomFieldDefinitionRepository
	assets              ports.AssetRepository
	checkouts           ports.AssetCheckoutRepository
	attachments         ports.AttachmentRepository
	assetTags           ports.AssetTagRepository
	assetUnitOfWork     ports.AssetUnitOfWork
	assetTagUnitOfWork  ports.AssetTagUnitOfWork
	assetEditUnitOfWork ports.AssetEditUnitOfWork
	undoables           ports.UndoableOperationRepository
	audit               ports.AuditRepository
	ids                 ports.IDGenerator
	clock               ports.Clock
	defaultPageLimit    int
	maxPageLimit        int
}

type Dependencies struct {
	Observer            ports.Observer
	Authorizer          ports.Authorizer
	Tenants             ports.TenantRepository
	Inventories         ports.InventoryRepository
	CustomAssetTypes    ports.CustomAssetTypeRepository
	CustomFields        ports.CustomFieldDefinitionRepository
	Assets              ports.AssetRepository
	Checkouts           ports.AssetCheckoutRepository
	Attachments         ports.AttachmentRepository
	AssetTags           ports.AssetTagRepository
	AssetUnitOfWork     ports.AssetUnitOfWork
	AssetTagUnitOfWork  ports.AssetTagUnitOfWork
	AssetEditUnitOfWork ports.AssetEditUnitOfWork
	Undoables           ports.UndoableOperationRepository
	Audit               ports.AuditRepository
	IDs                 ports.IDGenerator
	Clock               ports.Clock
	DefaultPageLimit    int
	MaxPageLimit        int
}

func New(deps Dependencies) Service {
	maxPageLimit := appsupport.NormalizeMaxPageLimit(deps.MaxPageLimit)
	observer := deps.Observer
	if observer == nil {
		observer = noopObserver{}
	}
	assetEditUnitOfWork := deps.AssetEditUnitOfWork
	if assetEditUnitOfWork == nil {
		assetEditUnitOfWork, _ = deps.AssetUnitOfWork.(ports.AssetEditUnitOfWork)
	}
	return Service{
		observer:            observer,
		authorizer:          deps.Authorizer,
		tenants:             deps.Tenants,
		inventories:         deps.Inventories,
		customAssetTypes:    deps.CustomAssetTypes,
		customFields:        deps.CustomFields,
		assets:              deps.Assets,
		checkouts:           deps.Checkouts,
		attachments:         deps.Attachments,
		assetTags:           deps.AssetTags,
		assetUnitOfWork:     deps.AssetUnitOfWork,
		assetTagUnitOfWork:  deps.AssetTagUnitOfWork,
		assetEditUnitOfWork: assetEditUnitOfWork,
		undoables:           deps.Undoables,
		audit:               deps.Audit,
		ids:                 deps.IDs,
		clock:               deps.Clock,
		defaultPageLimit:    appsupport.NormalizeDefaultPageLimit(deps.DefaultPageLimit, maxPageLimit),
		maxPageLimit:        maxPageLimit,
	}
}

type noopObserver struct{}

func (noopObserver) Record(context.Context, ports.Event) {}

func (s Service) ensureInventoryAccessDependencies() error {
	if s.tenants == nil || s.inventories == nil || s.authorizer == nil {
		return apperrors.ErrInvalidInput
	}
	return nil
}

func (s Service) ensureAssetRepository() error {
	if s.assets == nil {
		return apperrors.ErrInvalidInput
	}
	return nil
}

func (s Service) ensureCheckoutDependencies() error {
	if s.assets == nil || s.checkouts == nil || s.assetUnitOfWork == nil || s.undoables == nil {
		return apperrors.ErrInvalidInput
	}
	return nil
}

type defaultClock struct{}

func (defaultClock) Now() time.Time {
	return time.Now().UTC()
}

func (s Service) now() time.Time {
	if s.clock == nil {
		return defaultClock{}.Now()
	}
	return s.clock.Now()
}

func (s Service) newID() string {
	if s.ids == nil {
		return ""
	}
	return s.ids.NewID()
}

func pageLimit(defaultLimit int, maxLimit int, requested int) int {
	return appsupport.PageLimit(defaultLimit, maxLimit, requested)
}
