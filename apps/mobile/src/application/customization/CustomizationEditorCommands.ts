import type { ManageCustomAssetTypes } from './ManageCustomAssetTypes';
import type { ManageCustomFields } from './ManageCustomFields';
import type { ManageTags } from './ManageTags';
import type { CustomizationContext } from './CustomizationRepository';
import type { CustomFieldDefinition, CustomizationKind, CustomizationScope } from '../../domain/customization/Customization';
import type { CustomizationEditorDraft } from './CustomizationEditorModel';

export type CustomizationEditorManagers = {
  readonly assetTypes: ManageCustomAssetTypes;
  readonly fields: ManageCustomFields;
  readonly tags: ManageTags;
};

export async function saveCustomizationEditor(input: {
  readonly context: CustomizationContext;
  readonly draft: CustomizationEditorDraft;
  readonly kind: CustomizationKind;
  readonly managers: CustomizationEditorManagers;
  readonly mode: 'create' | 'edit';
  readonly record?: CustomFieldDefinition;
  readonly resourceId?: string;
  readonly scope: CustomizationScope;
}): Promise<void> {
  const { context, draft, kind, managers, mode, resourceId, scope } = input;
  if (kind === 'tag') {
    if (mode === 'create') await managers.tags.create(context, { displayName: draft.name, color: draft.color });
    else await managers.tags.update(context, requiredId(resourceId), { displayName: draft.name, color: draft.color });
    return;
  }
  if (kind === 'asset-type') {
    if (mode === 'create') await managers.assetTypes.create(context, scope, { key: draft.key, displayName: draft.name, description: draft.description });
    else await managers.assetTypes.update(address(context, scope, requiredId(resourceId)), { displayName: draft.name.trim(), description: draft.description.trim() });
    return;
  }
  if (mode === 'create') {
    await managers.fields.create(context, scope, { key: draft.key, displayName: draft.name, type: draft.fieldType, enumOptions: draft.enumOptions, applicability: draft.applicability, customAssetTypeIds: draft.targetIds });
  } else {
    await managers.fields.update(address(context, scope, requiredId(resourceId)), input.record as CustomFieldDefinition, { displayName: draft.name.trim(), enumOptions: draft.enumOptions, applicability: draft.applicability, customAssetTypeIds: draft.targetIds });
  }
}

type CustomizationLifecycleCommand = {
  readonly context: CustomizationContext;
  readonly managers: CustomizationEditorManagers;
  readonly resourceId: string;
  readonly scope: CustomizationScope;
} & (
  | { readonly action: 'archive'; readonly kind: 'tag' }
  | { readonly action: 'archive' | 'restore' | 'delete'; readonly kind: 'field' | 'asset-type' }
);

export async function runCustomizationLifecycle(input: CustomizationLifecycleCommand): Promise<void> {
  const target = address(input.context, input.scope, input.resourceId);
  if (input.kind === 'tag') {
    if ((input as { readonly action: string }).action !== 'archive') throw new Error('Tags support archive only.');
    await input.managers.tags.archive(input.context, input.resourceId);
  }
  else if (input.kind === 'field') await input.managers.fields[input.action](target);
  else await input.managers.assetTypes[input.action](target);
}

export async function runCustomizationLifecycleIntent(input: {
  readonly action: 'archive' | 'restore' | 'delete';
  readonly context: CustomizationContext;
  readonly kind: CustomizationKind;
  readonly managers: CustomizationEditorManagers;
  readonly resourceId: string;
  readonly scope: CustomizationScope;
}): Promise<void> {
  if (input.kind === 'tag') {
    if (input.action !== 'archive') throw new Error('Tags support archive only.');
    return runCustomizationLifecycle({ ...input, kind: 'tag', action: 'archive' });
  }
  return runCustomizationLifecycle({ ...input, kind: input.kind, action: input.action });
}

function address(context: CustomizationContext, scope: CustomizationScope, id: string) {
  return { scope, tenantId: context.tenantId, inventoryId: scope === 'inventory' ? context.inventoryId : undefined, id } as const;
}

function requiredId(value: string | undefined): string {
  if (!value) throw new Error('Customization resource ID is required.');
  return value;
}
