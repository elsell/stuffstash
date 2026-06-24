import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import AssetDetail from './AssetDetail.svelte';
import type { AssetViewModel, UpdateAssetDraft } from '$lib/domain/inventory';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('AssetDetail', () => {
  it('preserves custom field values when moving an asset', async () => {
    let savedDraft: UpdateAssetDraft | null = null;
    component = mount(AssetDetail, {
      target: document.body,
      props: {
        asset: asset(),
        canEdit: true,
        parentTargets: [],
        customFieldDefinitions: [
          {
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
            lifecycleState: 'active'
          }
        ],
        saving: false,
        attachments: [],
        onBack: () => {},
        onSave: async (draft) => {
          savedDraft = draft;
        },
        onArchive: async () => {},
        onRestore: async () => {},
        onDelete: async () => {},
        onArchiveAttachment: async () => {},
        onDeleteAttachment: async () => {}
      }
    });

    clickFirst('Move');
    await flush();
    clickFirst('Inventory root');
    clickLast('Move');
    await flush();

    expect(savedDraft).toMatchObject({
      parentAssetId: null,
      customFields: { 'expiration-date': '2027-01-01' }
    });
  });
});

function asset(): AssetViewModel {
  return {
    id: 'asset-one',
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    kind: 'item',
    title: 'Ibuprofen',
    description: '',
    parentAssetId: 'parent-one',
    lifecycleState: 'active',
    customAssetTypeId: 'type-medicine',
    customAssetTypeLabel: 'Medicine',
    customFields: { 'expiration-date': '2027-01-01' },
    containmentTrail: 'Hall closet'
  };
}

function clickFirst(text: string): void {
  const button = buttons(text)[0];
  if (!button) throw new Error(`Missing button ${text}`);
  button.click();
}

function clickLast(text: string): void {
  const matching = buttons(text);
  const button = matching[matching.length - 1];
  if (!button) throw new Error(`Missing button ${text}`);
  button.click();
}

function buttons(text: string): HTMLButtonElement[] {
  return Array.from(document.body.querySelectorAll('button')).filter((candidate) => candidate.textContent?.includes(text));
}

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
}
