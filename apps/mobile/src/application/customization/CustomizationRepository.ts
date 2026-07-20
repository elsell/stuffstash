import type {
  AssetTagDefinition,
  CustomAssetTypeDefinition,
  CustomFieldApplicability,
  CustomFieldDefinition,
  CustomFieldType,
  CustomizationLifecycle,
  CustomizationScope
} from '../../domain/customization/Customization';

export type CustomizationContext = {
  readonly tenantId: string;
  readonly tenantName: string;
  readonly tenantPermissions: readonly string[];
  readonly inventoryId: string;
  readonly inventoryName: string;
  readonly inventoryPermissions: readonly string[];
};

export type CustomizationPage<T> = {
  readonly items: readonly T[];
  readonly nextCursor?: string;
};

export type DefinitionAddress = {
  readonly scope: CustomizationScope;
  readonly tenantId: string;
  readonly inventoryId?: string;
  readonly id: string;
};

export interface CustomizationRepository {
  listTags(context: CustomizationContext, cursor?: string): Promise<CustomizationPage<AssetTagDefinition>>;
  createTag(context: CustomizationContext, input: { readonly displayName: string; readonly color?: string }): Promise<AssetTagDefinition>;
  updateTag(context: CustomizationContext, id: string, input: { readonly displayName?: string; readonly color?: string }): Promise<AssetTagDefinition>;
  archiveTag(context: CustomizationContext, id: string): Promise<void>;

  listFields(context: CustomizationContext, scope: CustomizationScope, lifecycle: CustomizationLifecycle, cursor?: string): Promise<CustomizationPage<CustomFieldDefinition>>;
  createField(context: CustomizationContext, scope: CustomizationScope, input: CreateCustomFieldInput): Promise<CustomFieldDefinition>;
  updateField(address: DefinitionAddress, input: UpdateCustomFieldInput): Promise<CustomFieldDefinition>;
  archiveField(address: DefinitionAddress): Promise<void>;
  restoreField(address: DefinitionAddress): Promise<void>;
  deleteField(address: DefinitionAddress): Promise<void>;

  listAssetTypes(context: CustomizationContext, scope: CustomizationScope, lifecycle: CustomizationLifecycle, cursor?: string): Promise<CustomizationPage<CustomAssetTypeDefinition>>;
  createAssetType(context: CustomizationContext, scope: CustomizationScope, input: CreateCustomAssetTypeInput): Promise<CustomAssetTypeDefinition>;
  updateAssetType(address: DefinitionAddress, input: UpdateCustomAssetTypeInput): Promise<CustomAssetTypeDefinition>;
  archiveAssetType(address: DefinitionAddress): Promise<void>;
  restoreAssetType(address: DefinitionAddress): Promise<void>;
  deleteAssetType(address: DefinitionAddress): Promise<void>;
}

export type CreateCustomFieldInput = {
  readonly key: string;
  readonly displayName: string;
  readonly type: CustomFieldType;
  readonly enumOptions: readonly string[];
  readonly applicability: CustomFieldApplicability;
  readonly customAssetTypeIds: readonly string[];
};

export type UpdateCustomFieldInput = {
  readonly displayName?: string;
  readonly enumOptions?: readonly string[];
  readonly applicability?: CustomFieldApplicability;
  readonly customAssetTypeIds?: readonly string[];
};

export type CreateCustomAssetTypeInput = {
  readonly key: string;
  readonly displayName: string;
  readonly description: string;
};

export type UpdateCustomAssetTypeInput = {
  readonly displayName?: string;
  readonly description?: string;
};
