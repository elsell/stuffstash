package httpapi

import "github.com/stuffstash/stuff-stash/cli/internal/ports"

func validateDefinitionScope(s ports.DefinitionScope) error {
	if s.Level != ports.HouseholdDefinition && s.Level != ports.InventoryDefinition {
		return ports.Failure("usage", "Select household or inventory scope.")
	}
	if s.Scope.Tenant == "" || s.Level == ports.InventoryDefinition && s.Scope.Inventory == "" {
		return ports.Failure("usage", "Supply the household and, for inventory scope, the inventory ID.")
	}
	return nil
}
