import type {
  CustomAssetType,
  CustomFieldApplicability,
  CustomFieldDefinition,
  CustomFieldType
} from '$lib/domain/inventory';
import type { Pagination } from './pagination';

export interface CustomizationPage<T> {
  items: T[];
  pagination: Pagination;
}

export interface CustomAssetTypeDraft {
  scope: 'tenant' | 'inventory';
  key: string;
  displayName: string;
  description: string;
}

export interface CustomFieldDefinitionDraft {
  scope: 'tenant' | 'inventory';
  key: string;
  displayName: string;
  type: CustomFieldType;
  enumOptions: string[];
  applicability: CustomFieldApplicability;
  customAssetTypeIds: string[];
}

export interface InventoryCustomizationRepository {
  listInventoryCustomAssetTypes(tenantId: string, inventoryId: string, cursor?: string): Promise<CustomizationPage<CustomAssetType>>;
  createCustomAssetType(tenantId: string, inventoryId: string, draft: CustomAssetTypeDraft): Promise<CustomAssetType>;
  archiveCustomAssetType(tenantId: string, inventoryId: string, customAssetTypeId: string, scope: 'tenant' | 'inventory'): Promise<CustomAssetType>;
  listInventoryCustomFieldDefinitions(
    tenantId: string,
    inventoryId: string,
    cursor?: string
  ): Promise<CustomizationPage<CustomFieldDefinition>>;
  createCustomFieldDefinition(
    tenantId: string,
    inventoryId: string,
    draft: CustomFieldDefinitionDraft
  ): Promise<CustomFieldDefinition>;
  archiveCustomFieldDefinition(
    tenantId: string,
    inventoryId: string,
    definitionId: string,
    scope: 'tenant' | 'inventory'
  ): Promise<CustomFieldDefinition>;
}
