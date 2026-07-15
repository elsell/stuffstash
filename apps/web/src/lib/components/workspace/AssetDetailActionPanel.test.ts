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
  it('uses verb-form copy throughout the check-out sheet', async () => {
    component = mount(AssetDetailActionPanel, {
      target: document.body,
      props: panelProps({ panel: 'checkout' })
    });
    await tick();

    expect(document.body.textContent).toContain('Check out asset');
    expect(button('Check out')).not.toBeNull();
    expect(document.body.textContent).not.toContain('Checkout asset');
  });

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
    await tick();

    expect(document.body.querySelector('[role="dialog"]')).not.toBeNull();
    expect(input('edit-asset-title').value).toBe('Ibuprofen');
    expect(input('edit-custom-field-expiration-date').value).toBe('2027-01-01');

    input('edit-custom-field-expiration-date').value = '2028-01-01';
    input('edit-custom-field-expiration-date').dispatchEvent(new Event('input', { bubbles: true }));
    await tick();

    expect(customChanges).toEqual([['expiration-date', '2028-01-01']]);
  });

  it('does not offer no-op edit or move submissions as primary actions', async () => {
    component = mount(AssetDetailActionPanel, {
      target: document.body,
      props: panelProps({
        panel: 'edit',
        title: 'Ibuprofen',
        description: '',
        parentAssetId: 'parent-one'
      })
    });
    await tick();

    expect(button('Save').disabled).toBe(true);
    input('edit-asset-title').value = 'Ibuprofen tablets';
    input('edit-asset-title').dispatchEvent(new Event('input', { bubbles: true }));
    await tick();
    expect(button('Save').disabled).toBe(false);

    unmount(component);
    component = mount(AssetDetailActionPanel, {
      target: document.body,
      props: panelProps({
        panel: 'move',
        title: 'Ibuprofen',
        parentAssetId: 'parent-one'
      })
    });
    await tick();

    expect(button('Move').disabled).toBe(true);
  });

  it('keeps empty custom fields discoverable without turning the edit sheet into a wall of controls', async () => {
    component = mount(AssetDetailActionPanel, {
      target: document.body,
      props: panelProps({
        panel: 'edit',
        applicableFields: [
          customFieldDefinition('serial-number', 'Serial number'),
          customFieldDefinition('purchase-store', 'Purchase store')
        ],
        customFieldValues: { 'serial-number': '', 'purchase-store': '' }
      })
    });
    await tick();

    const disclosure = document.body.querySelector<HTMLDetailsElement>('.edit-empty-fields');
    expect(disclosure?.open).toBe(false);
    expect(disclosure?.querySelector('summary')?.textContent).toContain('Show 2 empty fields');
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
    await tick();

    expect(document.body.textContent).toContain('Move asset');
    expect(document.body.textContent).toContain('Garage shelf');

    button('Garage shelf').click();
    await tick();

    expect(selectedParent).toBe('parent-two');
  });

  it('labels the shared move sheet as a place workflow for locations', async () => {
    component = mount(AssetDetailActionPanel, {
      target: document.body,
      props: panelProps({ panel: 'move', asset: { ...asset(), kind: 'location' } })
    });
    await tick();

    expect(document.body.textContent).toContain('Move place');
    expect(document.body.textContent).not.toContain('Move asset');
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
    await tick();

    expect(document.body.textContent).toContain('Delete manual.pdf permanently?');
    expect(link('Cancel').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/assets/asset-one');

    button('Delete').click();
    await tick();

    expect(deleted).toEqual(['attachment']);
  });

  it('focuses the safe Cancel action for asset and attachment confirmations', async () => {
    for (const testCase of [
      { panel: 'archive' as const, selectedAttachment: null },
      { panel: 'delete' as const, selectedAttachment: null },
      { panel: 'attachment-delete' as const, selectedAttachment: attachment('manual', 'manual.pdf') }
    ]) {
      component = mount(AssetDetailActionPanel, {
        target: document.body,
        props: panelProps(testCase)
      });
      await tick();

      expect((document.activeElement as HTMLElement | null)?.textContent).toContain('Cancel');

      unmount(component);
      component = null;
      document.body.innerHTML = '';
    }
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
    await tick();

    expect(document.body.textContent).toContain('Delete Ibuprofen permanently?');

    button('Delete').click();
    await tick();

    expect(deleted).toEqual(['asset']);
  });
});

function panelProps(overrides: Partial<AssetDetailActionPanelProps> = {}): AssetDetailActionPanelProps {
  return {
    panel: 'none',
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
    checkoutDetails: '',
    customFieldValues: {},
    onClose: () => {},
    onDismiss: () => {},
    onCloseAutoFocus: () => {},
    onSave: async (_draft?: UpdateAssetDraft) => {},
    onArchive: async () => {},
    onRestore: async () => {},
    onDelete: async () => {},
    onCheckout: async () => {},
    onReturn: async () => {},
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
