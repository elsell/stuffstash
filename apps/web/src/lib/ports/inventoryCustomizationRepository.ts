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

export interface CustomAssetTypeUpdate {
  displayName: string;
  description: string;
}

export interface CustomFieldDefinitionUpdate {
  displayName: string;
  enumOptions?: string[];
  applicability?: CustomFieldApplicability;
  customAssetTypeIds?: string[];
}

export type CustomizationLifecycleFilter = 'active' | 'archived';

export interface InventoryCustomizationRepository {
  listTenantCustomAssetTypes(tenantId: string, cursor?: string, lifecycleState?: CustomizationLifecycleFilter): Promise<CustomizationPage<CustomAssetType>>;
  listInventoryCustomAssetTypes(tenantId: string, inventoryId: string, cursor?: string, lifecycleState?: CustomizationLifecycleFilter): Promise<CustomizationPage<CustomAssetType>>;
  createCustomAssetType(tenantId: string, inventoryId: string, draft: CustomAssetTypeDraft): Promise<CustomAssetType>;
  updateCustomAssetType(tenantId: string, inventoryId: string, customAssetTypeId: string, scope: 'tenant' | 'inventory', update: CustomAssetTypeUpdate): Promise<CustomAssetType>;
  archiveCustomAssetType(tenantId: string, inventoryId: string, customAssetTypeId: string, scope: 'tenant' | 'inventory'): Promise<CustomAssetType>;
  restoreCustomAssetType(tenantId: string, inventoryId: string, customAssetTypeId: string, scope: 'tenant' | 'inventory'): Promise<CustomAssetType>;
  deleteCustomAssetType(tenantId: string, inventoryId: string, customAssetTypeId: string, scope: 'tenant' | 'inventory'): Promise<void>;
  listTenantCustomFieldDefinitions(
    tenantId: string,
    cursor?: string,
    lifecycleState?: CustomizationLifecycleFilter
  ): Promise<CustomizationPage<CustomFieldDefinition>>;
  listInventoryCustomFieldDefinitions(
    tenantId: string,
    inventoryId: string,
    cursor?: string,
    lifecycleState?: CustomizationLifecycleFilter
  ): Promise<CustomizationPage<CustomFieldDefinition>>;
  createCustomFieldDefinition(
    tenantId: string,
    inventoryId: string,
    draft: CustomFieldDefinitionDraft
  ): Promise<CustomFieldDefinition>;
  updateCustomFieldDefinition(
    tenantId: string,
    inventoryId: string,
    definitionId: string,
    scope: 'tenant' | 'inventory',
    update: CustomFieldDefinitionUpdate
  ): Promise<CustomFieldDefinition>;
  archiveCustomFieldDefinition(
    tenantId: string,
    inventoryId: string,
    definitionId: string,
    scope: 'tenant' | 'inventory'
  ): Promise<CustomFieldDefinition>;
  restoreCustomFieldDefinition(
    tenantId: string,
    inventoryId: string,
    definitionId: string,
    scope: 'tenant' | 'inventory'
  ): Promise<CustomFieldDefinition>;
  deleteCustomFieldDefinition(
    tenantId: string,
    inventoryId: string,
    definitionId: string,
    scope: 'tenant' | 'inventory'
  ): Promise<void>;
}
