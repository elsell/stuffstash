import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import type { Asset, AssetKind, Inventory, Tenant, WorkspaceMode } from '$lib/domain/inventory';
import type { InventoryWorkspaceChromeProps } from './InventoryWorkspaceChrome.svelte';
import InventoryWorkspaceChromeHarness from './InventoryWorkspaceChromeHarness.test.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('InventoryWorkspaceChrome', () => {
  it('composes the workspace shell without owning route content', () => {
    component = mount(InventoryWorkspaceChromeHarness, {
      target: document.body,
      props: chromeProps()
    });

    const shell = requiredShell();
    expect(isInert(shell)).toBe(false);
    expect(shell.getAttribute('aria-hidden')).toBeNull();
    expect(document.body.querySelector('[aria-label="Workspace navigation"]')?.textContent).toContain('Stuff Stash');
    expect(document.body.querySelector('.workspace-header')?.textContent).toContain('Add');
    expect(document.body.querySelector('[aria-label="Mobile navigation"]')?.textContent).toContain('Browse');
    expect(document.body.querySelector('.workspace-main')?.textContent).toContain('Fixture workspace');
    expect(document.body.querySelectorAll('main.workspace-route-content')).toHaveLength(1);
  });

  it('hides the background shell from assistive technology while modal overlays are open', () => {
    component = mount(InventoryWorkspaceChromeHarness, {
      target: document.body,
      props: chromeProps({ modalOpen: true })
    });

    const shell = requiredShell();
    expect(isInert(shell)).toBe(true);
    expect(shell.getAttribute('aria-hidden')).toBe('true');
  });

  it('routes chrome actions through the coordinator callbacks', async () => {
    const selectedModes: WorkspaceMode[] = [];
    const addKinds: AssetKind[] = [];
    component = mount(InventoryWorkspaceChromeHarness, {
      target: document.body,
      props: chromeProps({
        onModeChange: (mode) => {
          selectedModes.push(mode);
        },
        onOpenAccountSettings: () => {
          selectedModes.push('settings');
        },
        onOpenAdd: (kind) => {
          addKinds.push(kind);
        }
      })
    });

    linkContaining('Browse').click();
    document.body.querySelector<HTMLAnchorElement>('a[aria-label="Add asset"]')?.click();
    expect(document.body.querySelector<HTMLButtonElement>('.header-add')?.getAttribute('aria-haspopup')).toBe('menu');

    expect(selectedModes).toContain('browse');
    expect(addKinds).toContain('item');

  });

  it('keeps search suggestions image-ready and updates the bound search query', async () => {
    component = mount(InventoryWorkspaceChromeHarness, {
      target: document.body,
      props: chromeProps({ searchQuery: 'drill' })
    });

    const search = document.body.querySelector<HTMLInputElement>('input[aria-label="Search this inventory"]');
    search?.focus();
    search?.dispatchEvent(new FocusEvent('focus'));
    await tick();

    expect(document.body.querySelector('[aria-label="Search suggestions"]')?.textContent).toContain('Cordless drill');
    expect(document.body.querySelector('img[alt="Cordless drill photo"]')).not.toBeNull();

    search!.value = 'saw';
    search?.dispatchEvent(new InputEvent('input', { bubbles: true }));
    await tick();

    expect(document.body.querySelector('[data-testid="bound-search-query"]')?.textContent).toBe('saw');
  });

  it('lets Browse own the search field', () => {
    component = mount(InventoryWorkspaceChromeHarness, {
      target: document.body,
      props: chromeProps({ mode: 'browse' })
    });

    expect(document.body.querySelector('input[aria-label="Search this inventory"]')).toBeNull();
    expect(document.body.querySelector('.workspace-header')?.textContent).toContain('Add');
  });

  it('makes route content and mobile navigation inert while the account sheet is open', async () => {
    component = mount(InventoryWorkspaceChromeHarness, {
      target: document.body,
      props: chromeProps()
    });

    const account = document.body.querySelector<HTMLButtonElement>('[aria-label="Open account menu"]');
    account?.click();
    await flush();

    const routeContent = document.body.querySelector<HTMLElement>('.workspace-route-content');
    const mobileNavigation = document.body.querySelector<HTMLElement>('.mobile-nav-shell');
    expect(routeContent?.getAttribute('aria-hidden')).toBe('true');
    expect(isInert(routeContent!)).toBe(true);
    expect(mobileNavigation?.getAttribute('aria-hidden')).toBe('true');
    expect(isInert(mobileNavigation!)).toBe(true);
  });

  it('suppresses bottom navigation on focused asset and place task surfaces', () => {
    component = mount(InventoryWorkspaceChromeHarness, {
      target: document.body,
      props: chromeProps({ mode: 'asset' })
    });

    expect(document.body.querySelector('[aria-label="Mobile navigation"]')).toBeNull();
    expect(document.body.querySelector('[aria-label="Workspace navigation"]')).not.toBeNull();
  });
});

function chromeProps(overrides: Partial<InventoryWorkspaceChromeProps> = {}): InventoryWorkspaceChromeProps {
  const tenants: Tenant[] = [{ id: 'tenant-one', name: 'Household', access: { relationship: 'owner', permissions: ['view'] } }];
  const inventories: Inventory[] = [
    {
      id: 'inventory-one',
      tenantId: 'tenant-one',
      name: 'Garage',
      access: { relationship: 'owner', permissions: ['view', 'create_asset'] }
    }
  ];
  const asset: Asset = {
    id: 'asset-one',
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    kind: 'item',
    title: 'Cordless drill',
    description: '',
    parentAssetId: null,
    lifecycleState: 'active',
    photo: {
      id: 'photo-one',
      assetId: 'asset-one',
      url: '/fixtures/drill.jpg',
      alt: 'Cordless drill photo'
    }
  };

  return {
    tenants,
    inventories,
    selectedTenantId: 'tenant-one',
    selectedInventoryId: 'inventory-one',
    selectedInventory: inventories[0],
    mode: 'home' as WorkspaceMode,
    settingsSection: 'overview' as const,
    userLabel: 'owner@example.com',
    searchSuggestions: [asset],
    searchQuery: '',
    canCreateAsset: true,
    disablePortal: true,
    onSelectTenant: () => {},
    onSelectInventory: () => {},
    onModeChange: () => {},
    onSearch: () => {},
    onOpenSearchAsset: () => {},
    onOpenAdd: () => {},
    onOpenAccountSettings: () => {},
    onSignOut: () => {},
    ...overrides
  };
}

function linkContaining(text: string): HTMLAnchorElement {
  const link = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!link) {
    throw new Error(`Missing link containing ${text}`);
  }
  return link;
}

function addMenuItemContaining(text: string): HTMLAnchorElement {
  const link = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('.add-menu a')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!link) {
    throw new Error(`Missing add menu item containing ${text}`);
  }
  return link;
}

function requiredShell(): HTMLElement {
  const shell = document.body.querySelector<HTMLElement>('.product-shell');
  if (!shell) {
    throw new Error('Missing product shell');
  }
  return shell;
}

function isInert(element: HTMLElement): boolean {
  const candidate = element as HTMLElement & { inert?: boolean };
  return typeof candidate.inert === 'boolean' ? candidate.inert : element.hasAttribute('inert');
}

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
}
