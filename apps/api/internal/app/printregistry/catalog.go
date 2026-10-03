package printregistry

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type CatalogService struct {
	Authorizer  ports.Authorizer
	Inventories ports.InventoryRepository
	Catalog     ports.PrinterCatalog
}

func (s CatalogService) List(ctx context.Context, principal identity.Principal, scope printing.Scope) ([]printing.PrinterProfile, error) {
	if principal.ID == "" {
		return nil, apperrors.ErrUnauthenticated
	}
	if scope.TenantID == "" || scope.InventoryID == "" {
		return nil, apperrors.ErrInvalidInput
	}
	item, found, err := s.Inventories.InventoryByID(ctx, tenant.ID(scope.TenantID), inventory.InventoryID(scope.InventoryID))
	if err != nil {
		return nil, err
	}
	if !found || !item.IsActive() {
		return nil, apperrors.ErrNotFound
	}
	if err := s.Authorizer.CheckInventory(ctx, principal, ports.InventoryPermissionView, inventory.InventoryID(scope.InventoryID)); err != nil {
		return nil, err
	}
	if s.Catalog == nil {
		return nil, errors.New("printer catalog unavailable")
	}
	return s.Catalog.ListPrinterProfiles(), nil
}
