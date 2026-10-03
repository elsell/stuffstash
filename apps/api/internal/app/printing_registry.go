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
	return printregistry.Service{Health: a.printConnectorRepository, PrintingAuthorization: a.printConnectorAuthorization, ReportMaxAge: a.printConnectorPolicy.ReportMaxAge, Authorizer: a.authorizer, Inventories: a.inventories, Printers: a.printerRepository, Catalog: a.printerCatalog, Audit: a.audit, IDs: a.ids, Clock: a.clock, Observer: a.observer}
}

func (a App) WithPrintConnectors(repository ports.ConnectorRepository, authorization ports.PrintingAuthorization, secrets ports.PairingSecrets, policy printregistry.ConnectorPolicy) App {
	a.printConnectorRepository = repository
	a.printConnectorAuthorization = authorization
	a.printPairingSecrets = secrets
	a.printConnectorPolicy = policy
	return a
}
func (a App) PrintConnectors() printregistry.ConnectorService {
	return printregistry.ConnectorService{Registry: a.PrinterRegistry(), Repository: a.printConnectorRepository, Authorization: a.printConnectorAuthorization, Secrets: a.printPairingSecrets, Policy: a.printConnectorPolicy}
}
func (a App) PrintConnectorsConfigured() bool {
	return a.printConnectorRepository != nil && a.printConnectorAuthorization != nil && a.printPairingSecrets != nil && a.printConnectorPolicy.Valid()
}
