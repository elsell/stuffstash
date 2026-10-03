package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/printregistry"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) WithPrinterCatalog(catalog ports.PrinterCatalog) App {
	a.printerCatalog = catalog
	return a
}
func (a App) ListPrinterProfiles(ctx context.Context, principal identity.Principal, scope printing.Scope) ([]printing.PrinterProfile, error) {
	return (printregistry.CatalogService{Authorizer: a.authorizer, Inventories: a.inventories, Catalog: a.printerCatalog}).List(ctx, principal, scope)
}
