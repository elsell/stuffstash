import { afterEach, describe, expect, it, vi } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import { Toaster } from '$lib/components/ui/sonner/index.js';
import type { InventoryWorkspaceOverlaysProps } from './InventoryWorkspaceOverlays.svelte';
import InventoryWorkspaceOverlays from './InventoryWorkspaceOverlays.svelte';
import InventoryWorkspaceOverlaysHarness from './InventoryWorkspaceOverlays.test-harness.svelte';

const gotoMock = vi.hoisted(() => vi.fn());

vi.mock('$app/navigation', () => ({
  goto: gotoMock
}));

let component: ReturnType<typeof mount> | null = null;
let toaster: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  if (toaster) {
    unmount(toaster);
    toaster = null;
  }
  document.body.innerHTML = '';
  gotoMock.mockClear();
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

  it('renders success, action, and error feedback as fixed workspace toasts', async () => {
    installMatchMedia();
    toaster = mount(Toaster, { target: document.body });
    component = mount(InventoryWorkspaceOverlays, {
      target: document.body,
      props: overlaysProps({
        notification: {
          kind: 'success',
          title: 'Saved Drill.',
          action: { label: 'View location', href: '/tenants/tenant-one/inventories/inventory-one/locations/location-one' }
        },
        error: 'Move not saved.'
      })
    });

    await waitFor(() => {
      const toasts = document.body.querySelectorAll('.stuffstash-toast');
      expect(toasts).toHaveLength(2);
      expect(document.body.textContent).toContain('Saved Drill.');
      expect(document.body.textContent).toContain('View location');
      expect(document.body.textContent).toContain('Move not saved.');
    });

    button('View location').click();
    expect(gotoMock).toHaveBeenCalledWith('/tenants/tenant-one/inventories/inventory-one/locations/location-one');
  });

  it('shows the same error again after the parent clears it', async () => {
    installMatchMedia();
    toaster = mount(Toaster, { target: document.body });
    component = mount(InventoryWorkspaceOverlaysHarness, { target: document.body });

    component.showError('Move not saved.');
    await waitForToastCount(1);

    component.showError('');
    await tick();
    component.showError('Move not saved.');

    await waitForToastCount(2);
  });

  it('shows same-copy success notifications when the action destination changes', async () => {
    installMatchMedia();
    toaster = mount(Toaster, { target: document.body });
    component = mount(InventoryWorkspaceOverlaysHarness, { target: document.body });

    component.showNotification({
      kind: 'success',
      title: 'Moved Drill.',
      action: { label: 'View location', href: '/tenants/tenant-one/inventories/inventory-one/locations/shed' }
    });
    await waitForToastCount(1);

    component.showNotification({
      kind: 'success',
      title: 'Moved Drill.',
      action: { label: 'View location', href: '/tenants/tenant-one/inventories/inventory-one/locations/garage' }
    });

    await waitForToastCount(2);
  });
});

async function waitFor(assertion: () => void): Promise<void> {
  let lastError: unknown;
  for (let attempt = 0; attempt < 30; attempt += 1) {
    await tick();
    await new Promise((resolve) => window.setTimeout(resolve, 0));
    try {
      assertion();
      return;
    } catch (caught) {
      lastError = caught;
    }
  }
  throw lastError;
}

async function waitForToastCount(count: number): Promise<void> {
  await waitFor(() => {
    expect(document.body.querySelectorAll('.stuffstash-toast')).toHaveLength(count);
  });
}

function installMatchMedia(): void {
  Object.defineProperty(window, 'matchMedia', {
    configurable: true,
    writable: true,
    value: vi.fn().mockImplementation((query: string) => ({
      matches: false,
      media: query,
      onchange: null,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn()
    }))
  });
}

function button(text: string): HTMLButtonElement {
  const target = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find(
    (candidate) => candidate.textContent?.trim() === text
  );
  if (!target) {
    throw new Error(`Missing button ${text}`);
  }
  return target;
}

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
    notification: null,
    error: '',
    onAddClose: () => {},
    onAddSave: async () => ({ saved: true }),
    ...overrides
  };
}
