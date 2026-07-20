export type CustomizationScope = 'tenant' | 'inventory';
export type CustomizationLifecycle = 'active' | 'archived';
export type CustomizationKind = 'tag' | 'field' | 'asset-type';

export type AssetTagDefinition = {
  readonly kind: 'tag';
  readonly id: string;
  readonly key: string;
  readonly displayName: string;
  readonly color?: string;
};

export type CustomFieldType = 'text' | 'number' | 'boolean' | 'date' | 'url' | 'enum';
export type CustomFieldApplicability = 'all_assets' | 'custom_asset_types';

export type CustomFieldDefinition = {
  readonly kind: 'field';
  readonly id: string;
  readonly tenantId: string;
  readonly inventoryId?: string;
  readonly scope: CustomizationScope;
  readonly key: string;
  readonly displayName: string;
  readonly type: CustomFieldType;
  readonly enumOptions: readonly string[];
  readonly applicability: CustomFieldApplicability;
  readonly customAssetTypeIds: readonly string[];
  readonly lifecycle: CustomizationLifecycle;
};

export type CustomAssetTypeDefinition = {
  readonly kind: 'asset-type';
  readonly id: string;
  readonly tenantId: string;
  readonly inventoryId?: string;
  readonly scope: CustomizationScope;
  readonly key: string;
  readonly displayName: string;
  readonly description: string;
  readonly lifecycle: CustomizationLifecycle;
};

export type CustomDefinition = CustomFieldDefinition | CustomAssetTypeDefinition;

export function suggestedCustomizationKey(value: string): string {
  return value
    .normalize('NFKD')
    .replace(/[\u0300-\u036f]/g, '')
    .toLocaleLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .replace(/^[^a-z]+/, '')
    .slice(0, 80);
}

export const customizationKeyValidationMessage = 'Key must start with a letter and use lowercase letters, numbers, or hyphens.';

export function customizationKeyIsValid(value: string): boolean {
  return /^[a-z][a-z0-9-]{0,79}$/.test(value.trim());
}

export function customizationNameOrder<T extends { readonly displayName: string; readonly id: string }>(
  items: readonly T[]
): readonly T[] {
  return [...items].sort((left, right) => {
    const name = left.displayName.localeCompare(right.displayName, undefined, { sensitivity: 'base' });
    return name || left.id.localeCompare(right.id);
  });
}

export function normalizeTagColor(value: string | undefined): string | undefined {
  const raw = value?.trim();
  if (!raw) return undefined;
  const color = raw.startsWith('#') ? raw : `#${raw}`;
  return /^#[0-9a-fA-F]{6}$/.test(color) ? color.toUpperCase() : undefined;
}
