package customfields

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type Service struct {
	observer                  ports.Observer
	authorizer                ports.Authorizer
	tenants                   ports.TenantRepository
	inventories               ports.InventoryRepository
	customAssetTypes          ports.CustomAssetTypeRepository
	customAssetTypeUnitOfWork ports.CustomAssetTypeUnitOfWork
	customFields              ports.CustomFieldDefinitionRepository
	customFieldUnitOfWork     ports.CustomFieldDefinitionUnitOfWork
	audit                     ports.AuditRepository
	ids                       ports.IDGenerator
	clock                     ports.Clock
	defaultPageLimit          int
	maxPageLimit              int
}

type Dependencies struct {
	Observer                  ports.Observer
	Authorizer                ports.Authorizer
	Tenants                   ports.TenantRepository
	Inventories               ports.InventoryRepository
	CustomAssetTypes          ports.CustomAssetTypeRepository
	CustomAssetTypeUnitOfWork ports.CustomAssetTypeUnitOfWork
	CustomFields              ports.CustomFieldDefinitionRepository
	CustomFieldUnitOfWork     ports.CustomFieldDefinitionUnitOfWork
	Audit                     ports.AuditRepository
	IDs                       ports.IDGenerator
	Clock                     ports.Clock
	DefaultPageLimit          int
	MaxPageLimit              int
}

func New(deps Dependencies) Service {
	maxPageLimit := appsupport.NormalizeMaxPageLimit(deps.MaxPageLimit)
	observer := deps.Observer
	if observer == nil {
		observer = noopObserver{}
	}
	return Service{
		observer:                  observer,
		authorizer:                deps.Authorizer,
		tenants:                   deps.Tenants,
		inventories:               deps.Inventories,
		customAssetTypes:          deps.CustomAssetTypes,
		customAssetTypeUnitOfWork: deps.CustomAssetTypeUnitOfWork,
		customFields:              deps.CustomFields,
		customFieldUnitOfWork:     deps.CustomFieldUnitOfWork,
		audit:                     deps.Audit,
		ids:                       deps.IDs,
		clock:                     deps.Clock,
		defaultPageLimit:          appsupport.NormalizeDefaultPageLimit(deps.DefaultPageLimit, maxPageLimit),
		maxPageLimit:              maxPageLimit,
	}
}

type noopObserver struct{}

func (noopObserver) Record(context.Context, ports.Event) {}

func (s Service) ensureTenantExists(ctx context.Context, tenantID tenant.ID) error {
	exists, err := s.tenants.TenantExists(ctx, tenantID)
	if err != nil {
		return err
	}
	if !exists {
		return apperrors.ErrNotFound
	}
	return nil
}

func (s Service) ensureActiveInventoryAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID, permission ports.InventoryPermission) error {
	exists, err := s.tenants.TenantExists(ctx, tenantID)
	if err != nil {
		return err
	}
	if !exists {
		return apperrors.ErrNotFound
	}
	item, found, err := s.inventories.InventoryByID(ctx, tenantID, inventoryID)
	if err != nil {
		return err
	}
	if !found || !item.IsActive() {
		return apperrors.ErrNotFound
	}
	if err := s.authorizer.CheckInventory(ctx, principal, permission, inventoryID); err != nil {
		s.recordAuthorizationDenied(ctx, principal, tenantID)
		return err
	}
	return nil
}

func (s Service) recordAuthorizationDenied(ctx context.Context, principal identity.Principal, tenantID tenant.ID) {
	s.observer.Record(ctx, ports.Event{
		Name:    ports.EventAuthorizationDenied,
		Message: "authorization denied",
		Fields: map[string]string{
			"tenant_id":    tenantID.String(),
			"principal_id": principal.ID.String(),
		},
	})
}

func (s Service) newAuditRecord(input appsupport.AuditRecordInput) (audit.Record, error) {
	return appsupport.NewAuditRecord(s.ids, s.clock, input)
}

func (s Service) saveReadAuditRecord(ctx context.Context, input appsupport.AuditRecordInput) error {
	return appsupport.SaveReadAuditRecord(ctx, s.audit, s.ids, s.clock, input)
}
