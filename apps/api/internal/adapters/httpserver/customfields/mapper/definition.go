package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customfields/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
)

func DefinitionToResponse(definition customfield.Definition) dto.DefinitionResponse {
	options := make([]string, 0, len(definition.EnumOptions))
	for _, option := range definition.EnumOptions {
		options = append(options, option.String())
	}
	return dto.DefinitionResponse{
		ID:          definition.ID.String(),
		TenantID:    definition.TenantID.String(),
		InventoryID: definition.InventoryID.String(),
		Scope:       definition.Scope.String(),
		Key:         definition.Key.String(),
		DisplayName: definition.DisplayName.String(),
		Type:        definition.Type.String(),
		EnumOptions: options,
	}
}

func DefinitionsToResponse(definitions []customfield.Definition) []dto.DefinitionResponse {
	data := make([]dto.DefinitionResponse, 0, len(definitions))
	for _, definition := range definitions {
		data = append(data, DefinitionToResponse(definition))
	}
	return data
}
