import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import InventoryCustomizationManager from './InventoryCustomizationManager.svelte';
import type { CustomAssetType, Inventory, Tenant } from '$lib/domain/inventory';
import type { InventoryCustomizationRepository } from '$lib/ports/inventoryCustomizationRepository';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('InventoryCustomizationManager', () => {
  it('creates custom asset types and targeted custom fields through the customization port', async () => {
    const calls: string[] = [];
    const medicine: CustomAssetType = {
      id: 'type-medicine',
      tenantId: 'tenant-one',
      inventoryId: 'inventory-one',
      scope: 'inventory',
      key: 'medicine',
      displayName: 'Medicine',
      description: 'Medication',
      lifecycleState: 'active'
    };
    const schemaChanges: string[] = [];
    const repository: InventoryCustomizationRepository = {
      listInventoryCustomAssetTypes: async () => ({ items: [], pagination: page() }),
      createCustomAssetType: async (_tenantId, _inventoryId, draft) => {
        calls.push(`type:${draft.scope}:${draft.key}:${draft.displayName}`);
        return { ...medicine, key: draft.key, displayName: draft.displayName, description: draft.description };
      },
      archiveCustomAssetType: async () => medicine,
      listInventoryCustomFieldDefinitions: async () => ({ items: [], pagination: page() }),
      createCustomFieldDefinition: async (_tenantId, _inventoryId, draft) => {
        calls.push(`field:${draft.scope}:${draft.key}:${draft.type}:${draft.applicability}:${draft.customAssetTypeIds.join(',')}`);
        return {
          id: 'field-expiration',
          tenantId: 'tenant-one',
          inventoryId: 'inventory-one',
          scope: draft.scope,
          key: draft.key,
          displayName: draft.displayName,
          type: draft.type,
          enumOptions: draft.enumOptions,
          applicability: draft.applicability,
          customAssetTypeIds: draft.customAssetTypeIds,
          lifecycleState: 'active'
        };
      },
      archiveCustomFieldDefinition: async () => ({
        id: 'field-expiration',
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        scope: 'inventory',
        key: 'expiration-date',
        displayName: 'Expiration date',
        type: 'date',
        enumOptions: [],
        applicability: 'custom_asset_types',
        customAssetTypeIds: ['type-medicine'],
        lifecycleState: 'archived'
      })
    };

    component = mount(InventoryCustomizationManager, {
      target: document.body,
      props: {
        tenant: tenant(),
        inventory: inventory(),
        repository,
        initialAssetTypes: [medicine],
        initialFieldDefinitions: [],
        onSchemaChange: (assetTypes, fieldDefinitions) => {
          schemaChanges.push(`${assetTypes.length}:${fieldDefinitions.length}`);
        }
      }
    });
    await flush();

    input('#custom-type-key', 'medicine');
    input('#custom-type-name', 'Medicine');
    input('#custom-type-description', 'Medication');
    await flush();
    click('Create type');
    await flush();
    click('Types only');
    await flush();
    input('#custom-field-key', 'expiration-date');
    input('#custom-field-name', 'Expiration date');
    click('date');
    click('Medicine');
    await flush();
    click('Create field');
    await flush();

    expect(calls).toEqual([
      'type:inventory:medicine:Medicine',
      'field:inventory:expiration-date:date:custom_asset_types:type-medicine'
    ]);
    expect(schemaChanges).toEqual(['2:0', '2:1']);
  });
});

function tenant(): Tenant {
  return { id: 'tenant-one', name: 'Home', access: { relationship: 'owner', permissions: ['view', 'configure'] } };
}

function inventory(): Inventory {
  return {
    id: 'inventory-one',
    tenantId: 'tenant-one',
    name: 'Household',
    access: { relationship: 'owner', permissions: ['view', 'configure'] }
  };
}

function page() {
  return { limit: 50, nextCursor: null, hasMore: false };
}

function input(selector: string, value: string): void {
  const element = document.querySelector<HTMLInputElement | HTMLTextAreaElement>(selector);
  if (!element) throw new Error(`Missing input ${selector}`);
  element.value = value;
  element.dispatchEvent(new Event('input', { bubbles: true }));
}

function click(text: string): void {
  const button = Array.from(document.body.querySelectorAll('button')).find((candidate) => candidate.textContent?.includes(text));
  if (!button) throw new Error(`Missing button ${text}`);
  button.click();
}

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
}
