import { afterEach, describe, expect, it, vi } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import type { CustomAssetType, CustomFieldDefinition } from '$lib/domain/inventory';
import InventoryCustomizationArchivePanel, { customizationArchiveFocusTarget, type InventoryCustomizationArchivePanelProps } from './InventoryCustomizationArchivePanel.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(async () => {
  document.body.querySelector<HTMLElement>('[role="alertdialog"]')?.dispatchEvent(
    new KeyboardEvent('keydown', { key: 'Escape', bubbles: true })
  );
  await new Promise((resolve) => window.setTimeout(resolve, 20));
  if (component) {
    await unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('InventoryCustomizationArchivePanel', () => {
  it('restores the exact surviving trigger and falls back when that trigger disappears', () => {
    const trigger = document.createElement('button');
    const heading = document.createElement('h2');
    document.body.append(trigger, heading);

    expect(customizationArchiveFocusTarget(trigger, heading)).toBe(trigger);
    trigger.remove();
    expect(customizationArchiveFocusTarget(trigger, heading)).toBe(heading);
  });

  it('does not dismiss the route after external teardown', async () => {
    vi.useFakeTimers();
    let dismissed = 0;
    let closeAutoFocused = 0;
    try {
      component = mount(InventoryCustomizationArchivePanel, {
        target: document.body,
        props: panelProps({
          assetType: customAssetType('medicine', 'Medicine'),
          onClose: (event) => event.preventDefault(),
          onCloseAutoFocus: () => { closeAutoFocused += 1; },
          onDismiss: () => { dismissed += 1; }
        })
      });
      await tick();

      link('Cancel').click();
      await tick();
      expect(closeAutoFocused).toBe(1);
      expect(vi.getTimerCount()).toBeGreaterThan(0);
      await unmount(component);
      component = null;
      vi.runAllTimers();

      expect(dismissed).toBe(0);
    } finally {
      vi.useRealTimers();
    }
  });

  it('renders an asset type archive confirmation with durable cancel href', async () => {
    const archivedIds: string[] = [];
    component = mount(InventoryCustomizationArchivePanel, {
      target: document.body,
      props: panelProps({
        assetType: customAssetType('medicine', 'Medicine'),
        onArchiveAssetType: async (assetType) => {
          archivedIds.push(assetType.id);
          return true;
        }
      })
    });
    await tick();

    expect(document.body.querySelector('[role="alertdialog"]')).not.toBeNull();
    expect(document.activeElement?.textContent).toBe('Cancel');
    expect(link('Cancel').classList.contains('min-h-11')).toBe(true);
    expect(button('Archive').classList.contains('min-h-11')).toBe(true);
    expect(document.body.textContent).toContain('Archive asset type');
    expect(document.body.textContent).toContain('Medicine');
    expect(link('Cancel').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/settings/fields');

    button('Archive').click();
    await tick();

    expect(archivedIds).toEqual(['medicine']);
  });

  it('renders field definition confirmation and respects scope permissions', async () => {
    component = mount(InventoryCustomizationArchivePanel, {
      target: document.body,
      props: panelProps({
        fieldDefinition: customFieldDefinition('expiration-date', 'Expiration date'),
        canArchiveScope: (scope) => scope === 'tenant'
      })
    });
    await tick();

    expect(document.body.textContent).toContain('Archive field definition');
    expect(document.body.textContent).toContain('Expiration date');
    expect(button('Archive').disabled).toBe(true);
  });

  it('renders unavailable state when the route target is missing', async () => {
    component = mount(InventoryCustomizationArchivePanel, {
      target: document.body,
      props: panelProps()
    });
    await tick();

    expect(document.body.textContent).toContain('Archive target unavailable');
    expect(link('Back to fields').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/settings/fields');
  });

  it('keeps an archive failure inside the open confirmation', async () => {
    component = mount(InventoryCustomizationArchivePanel, {
      target: document.body,
      props: panelProps({
        assetType: customAssetType('medicine', 'Medicine'),
        error: 'Archive not saved. Medicine is still active.'
      })
    });
    await tick();

    const dialog = document.body.querySelector('[role="alertdialog"]');
    expect(dialog?.textContent).toContain('Archive not saved. Medicine is still active.');
    expect(dialog?.querySelector('[role="alert"]')).not.toBeNull();
  });
});

function panelProps(overrides: Partial<InventoryCustomizationArchivePanelProps> = {}): InventoryCustomizationArchivePanelProps {
  return {
    assetType: null,
    fieldDefinition: null,
    busy: false,
    fieldsHref: '/tenants/tenant-one/inventories/inventory-one/settings/fields',
    error: '',
    canArchiveScope: () => true,
    onClose: () => {},
    onDismiss: () => {},
    onCloseAutoFocus: () => {},
    onArchiveAssetType: async () => true,
    onArchiveFieldDefinition: async () => true,
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
