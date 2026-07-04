import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import type { AssetAttachment, AssetViewModel, CustomFieldDefinition, ParentTargetViewModel, UpdateAssetDraft } from '$lib/domain/inventory';
import AssetDetailActionPanel, { type AssetDetailActionPanelProps } from './AssetDetailActionPanel.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('AssetDetailActionPanel', () => {
  it('renders the edit panel with bound fields and custom field controls', async () => {
    const customChanges: Array<[string, string]> = [];
    component = mount(AssetDetailActionPanel, {
      target: document.body,
      props: panelProps({
        panel: 'edit',
        title: 'Ibuprofen',
        description: 'Pain relief',
        applicableFields: [customFieldDefinition('expiration-date', 'Expiration date')],
        customFieldValues: { 'expiration-date': '2027-01-01' },
        onCustomFieldValueChange: (key, value) => {
          customChanges.push([key, value]);
        }
      })
    });

    expect(document.body.querySelector('[aria-labelledby="edit-asset-panel-title"]')).not.toBeNull();
    expect(input('edit-asset-title').value).toBe('Ibuprofen');
    expect(input('edit-custom-field-expiration-date').value).toBe('2027-01-01');

    input('edit-custom-field-expiration-date').value = '2028-01-01';
    input('edit-custom-field-expiration-date').dispatchEvent(new Event('input', { bubbles: true }));
    await tick();

    expect(customChanges).toEqual([['expiration-date', '2028-01-01']]);
  });

  it('renders the move panel with the shared searchable parent picker', async () => {
    let selectedParent: string | null = null;
    component = mount(AssetDetailActionPanel, {
      target: document.body,
      props: panelProps({
        panel: 'move',
        parentTargets: [parentTarget('parent-two', 'Garage shelf', 'Garage')],
        onParentSelect: (id) => {
          selectedParent = id;
        }
      })
    });

    expect(document.body.textContent).toContain('Move asset');
    expect(document.body.textContent).toContain('Garage shelf');

    button('Garage shelf').click();
    await tick();

    expect(selectedParent).toBe('parent-two');
  });

  it('renders destructive confirmation panels with durable cancel links', async () => {
    const deleted: string[] = [];
    component = mount(AssetDetailActionPanel, {
      target: document.body,
      props: panelProps({
        panel: 'attachment-delete',
        selectedAttachment: attachment('manual', 'manual.pdf'),
        onDeleteAttachment: async () => {
          deleted.push('attachment');
        }
      })
    });

    expect(document.body.textContent).toContain('Delete manual.pdf permanently?');
    expect(link('Cancel').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/assets/asset-one');

    button('Delete').click();
    await tick();

    expect(deleted).toEqual(['attachment']);
  });

  it('routes asset delete confirmation through the asset delete callback', async () => {
    const deleted: string[] = [];
    component = mount(AssetDetailActionPanel, {
      target: document.body,
      props: panelProps({
        panel: 'delete',
        onDelete: async () => {
          deleted.push('asset');
        }
      })
    });

    expect(document.body.textContent).toContain('Delete Ibuprofen permanently?');

    button('Delete').click();
    await tick();

    expect(deleted).toEqual(['asset']);
  });
});

function panelProps(overrides: Partial<AssetDetailActionPanelProps> = {}): AssetDetailActionPanelProps {
  return {
    panel: 'none',
    panelElement: null,
    asset: asset(),
    parentTargets: [],
    selectedAttachment: null,
    saving: false,
    saveError: '',
    detailHref: '/tenants/tenant-one/inventories/inventory-one/assets/asset-one',
    applicableFields: [],
    title: '',
    description: '',
    parentAssetId: null,
    moveParentSearch: '',
    customFieldValues: {},
    onClose: () => {},
    onSave: async (_draft?: UpdateAssetDraft) => {},
    onArchive: async () => {},
    onRestore: async () => {},
    onDelete: async () => {},
    onDeleteAttachment: async () => {},
    onParentSelect: () => {},
    onCustomFieldValueChange: () => {},
    ...overrides
  };
}

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
    containmentTrail: 'Hall closet'
  };
}

function parentTarget(id: string, title: string, containmentTrail: string): ParentTargetViewModel {
  return {
    ...asset(),
    id,
    title,
    kind: 'container',
    parentAssetId: null,
    lifecycleState: 'active',
    containmentTrail
  };
}

function customFieldDefinition(key: string, displayName: string): CustomFieldDefinition {
  return {
    id: `field-${key}`,
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    scope: 'inventory',
    key,
    displayName,
    type: 'date',
    enumOptions: [],
    applicability: 'all_assets',
    customAssetTypeIds: [],
    lifecycleState: 'active'
  };
}

function attachment(id: string, fileName: string): AssetAttachment {
  return {
    id,
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    assetId: 'asset-one',
    fileName,
    contentType: 'application/pdf',
    sizeBytes: 12,
    lifecycleState: 'active'
  };
}

function input(id: string): HTMLInputElement {
  const target = document.body.querySelector<HTMLInputElement>(`#${id}`);
  if (!target) {
    throw new Error(`Missing input ${id}`);
  }
  return target;
}

function button(text: string): HTMLButtonElement {
  const target = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find((candidate) => candidate.textContent?.includes(text));
  if (!target) {
    throw new Error(`Missing button ${text}`);
  }
  return target;
}

function link(text: string): HTMLAnchorElement {
  const target = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) => candidate.textContent?.includes(text));
  if (!target) {
    throw new Error(`Missing link ${text}`);
  }
  return target;
}
