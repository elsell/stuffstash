import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import type { InventoryWorkspaceOverlaysProps } from './InventoryWorkspaceOverlays.svelte';
import InventoryWorkspaceOverlays from './InventoryWorkspaceOverlays.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('InventoryWorkspaceOverlays', () => {
  it('renders add tray only when creation is allowed', async () => {
    component = mount(InventoryWorkspaceOverlays, {
      target: document.body,
      props: overlaysProps({ addOpen: true, createAssetAllowed: false })
    });

    await tick();
    expect(document.body.textContent).not.toContain('Add item');

    unmount(component);
    component = mount(InventoryWorkspaceOverlays, {
      target: document.body,
      props: overlaysProps({ addOpen: true, createAssetAllowed: true })
    });

    await tick();
    expect(document.body.textContent).toContain('Add item');
  });

  it('renders success and error feedback as fixed workspace toasts', () => {
    component = mount(InventoryWorkspaceOverlays, {
      target: document.body,
      props: overlaysProps({ message: 'Saved Drill.', error: 'Move not saved.' })
    });

    const toasts = document.body.querySelectorAll('.toast');
    expect(toasts).toHaveLength(2);
    expect(toasts[0]?.textContent).toContain('Saved Drill.');
    expect(toasts[1]?.textContent).toContain('Move not saved.');
  });
});

function overlaysProps(overrides: Partial<InventoryWorkspaceOverlaysProps> = {}): InventoryWorkspaceOverlaysProps {
  return {
    addOpen: false,
    createAssetAllowed: true,
    addKind: 'item',
    addParentAssetId: null,
    addCloseHref: '/tenants/tenant-one/inventories/inventory-one',
    parentTargets: [],
    mediaPolicy: { supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'], maxBytes: 1024 },
    customAssetTypes: [],
    customFieldDefinitions: [],
    saving: false,
    message: '',
    error: '',
    onAddClose: () => {},
    onAddSave: async () => ({ saved: true }),
    ...overrides
  };
}
