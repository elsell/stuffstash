import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import type { CustomAssetType, CustomFieldDefinition } from '$lib/domain/inventory';
import InventoryCustomizationArchivePanel, { type InventoryCustomizationArchivePanelProps } from './InventoryCustomizationArchivePanel.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('InventoryCustomizationArchivePanel', () => {
  it('renders an asset type archive confirmation with durable cancel href', async () => {
    const archivedIds: string[] = [];
    component = mount(InventoryCustomizationArchivePanel, {
      target: document.body,
      props: panelProps({
        assetType: customAssetType('medicine', 'Medicine'),
        onArchiveAssetType: async (assetType) => {
          archivedIds.push(assetType.id);
        }
      })
    });

    expect(document.body.textContent).toContain('Archive asset type');
    expect(document.body.textContent).toContain('Medicine');
    expect(link('Cancel').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/settings/fields');

    button('Archive').click();
    await tick();

    expect(archivedIds).toEqual(['medicine']);
  });

  it('renders field definition confirmation and respects scope permissions', () => {
    component = mount(InventoryCustomizationArchivePanel, {
      target: document.body,
      props: panelProps({
        fieldDefinition: customFieldDefinition('expiration-date', 'Expiration date'),
        canArchiveScope: (scope) => scope === 'tenant'
      })
    });

    expect(document.body.textContent).toContain('Archive field definition');
    expect(document.body.textContent).toContain('Expiration date');
    expect(button('Archive').disabled).toBe(true);
  });

  it('renders unavailable state when the route target is missing', () => {
    component = mount(InventoryCustomizationArchivePanel, {
      target: document.body,
      props: panelProps()
    });

    expect(document.body.textContent).toContain('Archive target unavailable');
    expect(link('Back to fields').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/settings/fields');
  });
});

function panelProps(overrides: Partial<InventoryCustomizationArchivePanelProps> = {}): InventoryCustomizationArchivePanelProps {
  return {
    assetType: null,
    fieldDefinition: null,
    busy: false,
    fieldsHref: '/tenants/tenant-one/inventories/inventory-one/settings/fields',
    panelElement: null,
    canArchiveScope: () => true,
    onClose: () => {},
    onArchiveAssetType: async () => {},
    onArchiveFieldDefinition: async () => {},
    ...overrides
  };
}

function customAssetType(id: string, displayName: string): CustomAssetType {
  return {
    id,
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    scope: 'inventory',
    key: id,
    displayName,
    description: '',
    lifecycleState: 'active'
  };
}

function customFieldDefinition(id: string, displayName: string): CustomFieldDefinition {
  return {
    id,
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    scope: 'inventory',
    key: id,
    displayName,
    type: 'date',
    enumOptions: [],
    applicability: 'all_assets',
    customAssetTypeIds: [],
    lifecycleState: 'active'
  };
}

function button(text: string): HTMLButtonElement {
  const target = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find((candidate) => candidate.textContent?.trim() === text);
  if (!target) {
    throw new Error(`Missing button ${text}`);
  }
  return target;
}

function link(text: string): HTMLAnchorElement {
  const target = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) => candidate.textContent?.trim() === text);
  if (!target) {
    throw new Error(`Missing link ${text}`);
  }
  return target;
}
