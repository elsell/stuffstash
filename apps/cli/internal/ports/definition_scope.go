package ports

type DefinitionLevel string

const (
	HouseholdDefinition DefinitionLevel = "household"
	InventoryDefinition DefinitionLevel = "inventory"
)

type DefinitionScope struct {
	Level DefinitionLevel
	Scope Scope
}
