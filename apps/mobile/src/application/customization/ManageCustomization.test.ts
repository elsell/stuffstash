import { describe, expect, it } from 'vitest';
import type { CustomFieldDefinition } from '../../domain/customization/Customization';
import type { CustomizationRepository } from './CustomizationRepository';
import { ManageTags } from './ManageTags';
import { ManageCustomFields } from './ManageCustomFields';
import { ManageCustomAssetTypes } from './ManageCustomAssetTypes';
import { BufferedCustomizationObservability, noCustomizationObservability } from './CustomizationObservability';

describe('mobile customization commands', () => {
  it('normalizes a tag color and clears it with the API empty-color contract', async () => {
    const updates: unknown[] = [];
    const repository = {
      updateTag: async (_context: never, _id: string, input: unknown) => { updates.push(input); return { kind: 'tag', id: 'tag-1', key: 'tools', displayName: 'Tools' }; }
    } as unknown as CustomizationRepository;
    const command = new ManageTags(repository, noCustomizationObservability);
    await command.update({} as never, 'tag-1', { displayName: ' Tools ', color: '2f80ed' });
    await command.update({} as never, 'tag-1', { displayName: 'Tools', color: '' });
    expect(updates).toEqual([
      { displayName: 'Tools', color: '#2F80ED' },
      { displayName: 'Tools', color: '' }
    ]);
  });

  it('rejects a nonempty invalid color without calling the adapter or clearing the existing color', async () => {
    const updates: unknown[] = [];
    const repository = { updateTag: async (_context: never, _id: string, input: unknown) => { updates.push(input); return {} as never; } } as unknown as CustomizationRepository;
    const command = new ManageTags(repository, noCustomizationObservability);

    await expect(command.update({} as never, 'tag-1', { displayName: 'Tools', color: '#12345G' })).rejects.toThrow('six-digit hex');
    expect(updates).toEqual([]);
  });

  it('emits safe mutation requested, succeeded, and failed events', async () => {
    const events = new BufferedCustomizationObservability();
    let fail = false;
    const repository = { archiveTag: async () => { if (fail) throw new Error('unsafe tag name'); } } as unknown as CustomizationRepository;
    const command = new ManageTags(repository, events);
    await command.archive({} as never, 'tag-secret');
    fail = true;
    await expect(command.archive({} as never, 'tag-secret')).rejects.toThrow();

    expect(events.events()).toEqual([
      { name: 'customization.mutation_requested', resource: 'tag', scope: 'inventory', action: 'archive' },
      { name: 'customization.mutation_succeeded', resource: 'tag', scope: 'inventory', action: 'archive' },
      { name: 'customization.mutation_requested', resource: 'tag', scope: 'inventory', action: 'archive' },
      { name: 'customization.mutation_failed', resource: 'tag', scope: 'inventory', action: 'archive' }
    ]);
    expect(JSON.stringify(events.events())).not.toContain('tag-secret');
  });

  it('rejects incompatible field narrowing and option removal before the adapter', async () => {
    let updates = 0;
    const repository = { updateField: async () => { updates += 1; return field; } } as unknown as CustomizationRepository;
    const command = new ManageCustomFields(repository, noCustomizationObservability);
    await expect(command.update({} as never, field, { applicability: 'custom_asset_types' })).rejects.toThrow('cannot be narrowed');
    await expect(command.update({} as never, { ...field, applicability: 'custom_asset_types' }, { enumOptions: ['first'] })).rejects.toThrow('cannot be renamed');
    expect(updates).toBe(0);
  });

  it('provides create, edit, and lifecycle command parity for every supported resource', async () => {
    const calls: string[] = [];
    const repository = Object.fromEntries([
      'createTag', 'updateTag', 'archiveTag',
      'createField', 'updateField', 'archiveField', 'restoreField', 'deleteField',
      'createAssetType', 'updateAssetType', 'archiveAssetType', 'restoreAssetType', 'deleteAssetType'
    ].map((name) => [name, async () => { calls.push(name); return name.includes('Field') ? field : name.includes('AssetType') ? assetType : tag; }])) as unknown as CustomizationRepository;
    const tags = new ManageTags(repository, noCustomizationObservability);
    const fields = new ManageCustomFields(repository, noCustomizationObservability);
    const types = new ManageCustomAssetTypes(repository, noCustomizationObservability);
    const context = {} as never;
    const fieldAddress = { scope: 'inventory', tenantId: 'tenant-1', inventoryId: 'inventory-1', id: field.id } as const;
    const typeAddress = { ...fieldAddress, id: assetType.id };

    await tags.create(context, { displayName: 'Tools' });
    await tags.update(context, tag.id, { displayName: 'Hand tools', color: '' });
    await tags.archive(context, tag.id);
    await fields.create(context, 'inventory', field);
    await fields.update(fieldAddress, field, { displayName: 'Updated priority' });
    await fields.archive(fieldAddress); await fields.restore(fieldAddress); await fields.delete(fieldAddress);
    await types.create(context, 'inventory', assetType);
    await types.update(typeAddress, { displayName: 'Updated appliance', description: 'Updated' });
    await types.archive(typeAddress); await types.restore(typeAddress); await types.delete(typeAddress);

    expect(calls).toEqual([
      'createTag', 'updateTag', 'archiveTag',
      'createField', 'updateField', 'archiveField', 'restoreField', 'deleteField',
      'createAssetType', 'updateAssetType', 'archiveAssetType', 'restoreAssetType', 'deleteAssetType'
    ]);
  });
});

const field: CustomFieldDefinition = {
  kind: 'field', id: 'field-1', tenantId: 'tenant-1', scope: 'tenant', key: 'priority', displayName: 'Priority',
  type: 'enum', enumOptions: ['first', 'second'], applicability: 'all_assets', customAssetTypeIds: [], lifecycle: 'active'
};
const tag = { kind: 'tag', id: 'tag-1', key: 'tools', displayName: 'Tools' } as const;
const assetType = { kind: 'asset-type', id: 'type-1', tenantId: 'tenant-1', inventoryId: 'inventory-1', scope: 'inventory', key: 'appliance', displayName: 'Appliance', description: '', lifecycle: 'active' } as const;
