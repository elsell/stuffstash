package app

import (
	"github.com/stuffstash/stuff-stash/internal/app/printregistry"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) WithPrinterRegistry(repository ports.PrinterRepository, catalog ports.PrinterCatalog) App {
	a.printerRepository = repository
	a.printerCatalog = catalog
	return a
}
func (a App) PrinterRegistry() printregistry.Service {
	return printregistry.Service{Authorizer: a.authorizer, Inventories: a.inventories, Printers: a.printerRepository, Catalog: a.printerCatalog, Audit: a.audit, IDs: a.ids, Clock: a.clock, Observer: a.observer}
}
