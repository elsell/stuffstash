import type { CustomizationKind, CustomizationScope } from '../../domain/customization/Customization';

export function customizationEditorTarget(kind: CustomizationKind, scope: CustomizationScope, resourceId: string, lifecycle: 'active' | 'archived', inherited: boolean, canManageInherited: boolean) {
  const kindSegment = kind === 'asset-type' ? 'asset-types' : kind === 'field' ? 'fields' : 'tags';
  const localBase = `/settings/${scope === 'tenant' ? 'household' : 'inventory'}/${kindSegment}`;
  return {
    pathname: `${localBase}/[resourceId]`,
    params: { resourceId, lifecycle, inherited: inherited ? 'true' : 'false' }
  } as const;
}
