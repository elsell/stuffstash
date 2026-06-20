// @vitest-environment jsdom

import { mount, unmount, type ComponentProps } from 'svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import InventoryAssetFlow from './InventoryAssetFlow.svelte';

type InventoryAssetFlowProps = ComponentProps<typeof InventoryAssetFlow>;

const garageInventory = {
  id: 'inventory-one',
  tenantId: 'tenant-one',
  name: 'Garage'
};

const medicineInventory = {
  id: 'inventory-two',
  tenantId: 'tenant-one',
  name: 'Medicine'
};

const activeAsset = {
  id: 'asset-one',
  tenantId: 'tenant-one',
  inventoryId: 'inventory-one',
  kind: 'item' as const,
  title: 'Cordless drill',
  description: 'Shelf A',
  lifecycleState: 'active' as const,
  customFields: {}
};

const archivedAsset = {
  ...activeAsset,
  lifecycleState: 'archived' as const,
  title: 'Old paint'
};

let mounted: Parameters<typeof unmount>[0] | undefined;

afterEach(() => {
  document.body.innerHTML = '';
  if (mounted) {
    unmount(mounted);
    mounted = undefined;
  }
});

describe('InventoryAssetFlow', () => {
  it('sends inventory and active lifecycle actions through callbacks', () => {
    const onSelectInventory = vi.fn();
    const onSelectAssetLifecycle = vi.fn();
    const onArchiveAsset = vi.fn();
    const onDeleteAsset = vi.fn();

    mounted = mount(InventoryAssetFlow, {
      target: document.body,
      props: baseProps({
        assetLifecycleState: 'active',
        assets: [activeAsset],
        onSelectInventory,
        onSelectAssetLifecycle,
        onArchiveAsset,
        onDeleteAsset
      })
    });

    clickButton('Medicine');
    clickButton('Archived');
    clickButton('Archive');
    clickButton('Delete');

    expect(onSelectInventory).toHaveBeenCalledWith('inventory-two');
    expect(onSelectAssetLifecycle).toHaveBeenCalledWith('archived');
    expect(onArchiveAsset).toHaveBeenCalledWith(activeAsset);
    expect(onDeleteAsset).toHaveBeenCalledWith(activeAsset);
  });

  it('shows archived assets with restore actions and no creation form', () => {
    const onRestoreAsset = vi.fn();
    const onDeleteAsset = vi.fn();

    mounted = mount(InventoryAssetFlow, {
      target: document.body,
      props: baseProps({
        assetLifecycleState: 'archived',
        assets: [archivedAsset],
        onRestoreAsset,
        onDeleteAsset
      })
    });

    expect(document.body.textContent).toContain('Old paint');
    expect(document.body.textContent).not.toContain('Add asset');

    clickButton('Restore');
    clickButton('Delete');

    expect(onRestoreAsset).toHaveBeenCalledWith(archivedAsset);
    expect(onDeleteAsset).toHaveBeenCalledWith(archivedAsset);
  });
});

function baseProps(overrides: Partial<InventoryAssetFlowProps> = {}): InventoryAssetFlowProps {
  return {
    tenantName: 'Home',
    inventoryName: 'Garage',
    assetKind: 'item' as const,
    assetTitle: '',
    assetDescription: '',
    inventories: [garageInventory, medicineInventory],
    selectedInventory: garageInventory,
    assetLifecycleState: 'active',
    assets: [activeAsset],
    busy: false,
    onCreateInventory: vi.fn(),
    onSelectInventory: vi.fn(),
    onSelectAssetLifecycle: vi.fn(),
    onRefreshAssets: vi.fn(),
    onCreateAsset: vi.fn(),
    onArchiveAsset: vi.fn(),
    onRestoreAsset: vi.fn(),
    onDeleteAsset: vi.fn(),
    ...overrides
  };
}

function clickButton(label: string): void {
  const button = Array.from(document.querySelectorAll('button')).find((candidate) => candidate.textContent === label);
  if (!button) {
    throw new Error(`Button not found: ${label}`);
  }
  button.click();
}
