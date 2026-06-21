package app

import (
	"context"

	customfieldapp "github.com/stuffstash/stuff-stash/internal/app/customfields"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
)

type CreateCustomFieldDefinitionInput = customfieldapp.CreateCustomFieldDefinitionInput
type ListCustomFieldDefinitionsInput = customfieldapp.ListCustomFieldDefinitionsInput
type GetCustomFieldDefinitionInput = customfieldapp.GetCustomFieldDefinitionInput
type UpdateCustomFieldDefinitionLifecycleInput = customfieldapp.UpdateCustomFieldDefinitionLifecycleInput
type UpdateCustomFieldDefinitionInput = customfieldapp.UpdateCustomFieldDefinitionInput
type ListCustomFieldDefinitionsResult = customfieldapp.ListCustomFieldDefinitionsResult

type CreateCustomAssetTypeInput = customfieldapp.CreateCustomAssetTypeInput
type ListCustomAssetTypesInput = customfieldapp.ListCustomAssetTypesInput
type GetCustomAssetTypeInput = customfieldapp.GetCustomAssetTypeInput
type UpdateCustomAssetTypeInput = customfieldapp.UpdateCustomAssetTypeInput
type ArchiveCustomAssetTypeInput = customfieldapp.ArchiveCustomAssetTypeInput
type ListCustomAssetTypesResult = customfieldapp.ListCustomAssetTypesResult

func (a App) CreateTenantCustomFieldDefinition(ctx context.Context, input CreateCustomFieldDefinitionInput) (customfield.Definition, error) {
	return a.customFieldService.CreateTenantCustomFieldDefinition(ctx, input)
}

func (a App) CreateInventoryCustomFieldDefinition(ctx context.Context, input CreateCustomFieldDefinitionInput) (customfield.Definition, error) {
	return a.customFieldService.CreateInventoryCustomFieldDefinition(ctx, input)
}

func (a App) UpdateTenantCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionInput) (customfield.Definition, error) {
	return a.customFieldService.UpdateTenantCustomFieldDefinition(ctx, input)
}

func (a App) UpdateInventoryCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionInput) (customfield.Definition, error) {
	return a.customFieldService.UpdateInventoryCustomFieldDefinition(ctx, input)
}

func (a App) GetTenantCustomFieldDefinition(ctx context.Context, input GetCustomFieldDefinitionInput) (customfield.Definition, error) {
	return a.customFieldService.GetTenantCustomFieldDefinition(ctx, input)
}

func (a App) GetInventoryCustomFieldDefinition(ctx context.Context, input GetCustomFieldDefinitionInput) (customfield.Definition, error) {
	return a.customFieldService.GetInventoryCustomFieldDefinition(ctx, input)
}

func (a App) ArchiveTenantCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput) (customfield.Definition, error) {
	return a.customFieldService.ArchiveTenantCustomFieldDefinition(ctx, input)
}

func (a App) ArchiveInventoryCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput) (customfield.Definition, error) {
	return a.customFieldService.ArchiveInventoryCustomFieldDefinition(ctx, input)
}

func (a App) RestoreTenantCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput) (customfield.Definition, error) {
	return a.customFieldService.RestoreTenantCustomFieldDefinition(ctx, input)
}

func (a App) RestoreInventoryCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput) (customfield.Definition, error) {
	return a.customFieldService.RestoreInventoryCustomFieldDefinition(ctx, input)
}

func (a App) DeleteTenantCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput) error {
	return a.customFieldService.DeleteTenantCustomFieldDefinition(ctx, input)
}

func (a App) DeleteInventoryCustomFieldDefinition(ctx context.Context, input UpdateCustomFieldDefinitionLifecycleInput) error {
	return a.customFieldService.DeleteInventoryCustomFieldDefinition(ctx, input)
}

func (a App) ListTenantCustomFieldDefinitions(ctx context.Context, input ListCustomFieldDefinitionsInput) (ListCustomFieldDefinitionsResult, error) {
	return a.customFieldService.ListTenantCustomFieldDefinitions(ctx, input)
}

func (a App) ListInventoryCustomFieldDefinitions(ctx context.Context, input ListCustomFieldDefinitionsInput) (ListCustomFieldDefinitionsResult, error) {
	return a.customFieldService.ListInventoryCustomFieldDefinitions(ctx, input)
}

func (a App) CreateTenantCustomAssetType(ctx context.Context, input CreateCustomAssetTypeInput) (customfield.AssetType, error) {
	return a.customFieldService.CreateTenantCustomAssetType(ctx, input)
}

func (a App) CreateInventoryCustomAssetType(ctx context.Context, input CreateCustomAssetTypeInput) (customfield.AssetType, error) {
	return a.customFieldService.CreateInventoryCustomAssetType(ctx, input)
}

func (a App) UpdateTenantCustomAssetType(ctx context.Context, input UpdateCustomAssetTypeInput) (customfield.AssetType, error) {
	return a.customFieldService.UpdateTenantCustomAssetType(ctx, input)
}

func (a App) UpdateInventoryCustomAssetType(ctx context.Context, input UpdateCustomAssetTypeInput) (customfield.AssetType, error) {
	return a.customFieldService.UpdateInventoryCustomAssetType(ctx, input)
}

func (a App) ArchiveTenantCustomAssetType(ctx context.Context, input ArchiveCustomAssetTypeInput) (customfield.AssetType, error) {
	return a.customFieldService.ArchiveTenantCustomAssetType(ctx, input)
}

func (a App) ArchiveInventoryCustomAssetType(ctx context.Context, input ArchiveCustomAssetTypeInput) (customfield.AssetType, error) {
	return a.customFieldService.ArchiveInventoryCustomAssetType(ctx, input)
}

func (a App) RestoreTenantCustomAssetType(ctx context.Context, input ArchiveCustomAssetTypeInput) (customfield.AssetType, error) {
	return a.customFieldService.RestoreTenantCustomAssetType(ctx, input)
}

func (a App) RestoreInventoryCustomAssetType(ctx context.Context, input ArchiveCustomAssetTypeInput) (customfield.AssetType, error) {
	return a.customFieldService.RestoreInventoryCustomAssetType(ctx, input)
}

func (a App) DeleteTenantCustomAssetType(ctx context.Context, input ArchiveCustomAssetTypeInput) error {
	return a.customFieldService.DeleteTenantCustomAssetType(ctx, input)
}

func (a App) DeleteInventoryCustomAssetType(ctx context.Context, input ArchiveCustomAssetTypeInput) error {
	return a.customFieldService.DeleteInventoryCustomAssetType(ctx, input)
}

func (a App) GetTenantCustomAssetType(ctx context.Context, input GetCustomAssetTypeInput) (customfield.AssetType, error) {
	return a.customFieldService.GetTenantCustomAssetType(ctx, input)
}

func (a App) GetInventoryCustomAssetType(ctx context.Context, input GetCustomAssetTypeInput) (customfield.AssetType, error) {
	return a.customFieldService.GetInventoryCustomAssetType(ctx, input)
}

func (a App) ListTenantCustomAssetTypes(ctx context.Context, input ListCustomAssetTypesInput) (ListCustomAssetTypesResult, error) {
	return a.customFieldService.ListTenantCustomAssetTypes(ctx, input)
}

func (a App) ListInventoryCustomAssetTypes(ctx context.Context, input ListCustomAssetTypesInput) (ListCustomAssetTypesResult, error) {
	return a.customFieldService.ListInventoryCustomAssetTypes(ctx, input)
}
