import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import HomeWorkspace from './HomeWorkspace.svelte';
import type { AssetViewModel } from '$lib/domain/inventory';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('HomeWorkspace', () => {
  it('shows containment context once for descriptionless recent assets', () => {
    const asset: AssetViewModel = {
      id: 'tape',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'item',
      title: 'Tape measure',
      description: '',
      parentAssetId: 'garage',
      lifecycleState: 'active',
      containmentTrail: 'Garage'
    };

    component = mount(HomeWorkspace, {
      target: document.body,
      props: {
        lifecycleState: 'active',
        locations: [],
        recentAssets: [asset],
        archivedAssets: [],
        onOpenLocation: () => {},
        onOpenAsset: () => {},
        onOpenAdd: () => {},
        onSelectLifecycle: () => {}
      }
    });

    const row = Array.from(document.body.querySelectorAll('button')).find((button) => button.textContent?.includes('Tape measure'));
    expect(row?.textContent?.match(/Garage/g)).toHaveLength(1);
  });
});
