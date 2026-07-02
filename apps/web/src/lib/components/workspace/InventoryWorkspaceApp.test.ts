import { tick } from 'svelte';
import { mount, unmount } from 'svelte';
import { afterEach, describe, expect, it } from 'vitest';
import { SeededInventoryRepository } from '$lib/adapters/memory/seededInventoryRepository';
import type { WorkspaceSeed } from '$lib/ports/inventoryRepository';
import InventoryWorkspaceApp from './InventoryWorkspaceApp.svelte';

let component: ReturnType<typeof mount> | null = null;

const seed: WorkspaceSeed = {
  principal: { id: 'principal-one', email: 'owner@example.test' },
  tenants: [{ id: 'tenant-home', name: 'Home', access: { relationship: 'owner', permissions: ['view'] } }],
  inventories: [
    {
      id: 'inventory-household',
      tenantId: 'tenant-home',
      name: 'Household',
      access: { relationship: 'owner', permissions: ['view', 'create_asset', 'edit_asset'] }
    }
  ],
  customAssetTypes: [],
  customFieldDefinitions: [],
  assets: [
    {
      id: 'asset-home',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'item',
      title: 'Passport',
      description: 'Blue folder',
      parentAssetId: null,
      lifecycleState: 'active'
    },
    {
      id: 'asset-archived',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'item',
      title: 'Archived Passport',
      description: 'Old folder',
      parentAssetId: null,
      lifecycleState: 'archived'
    }
  ]
};

async function mountWorkspace(path: string): Promise<SeededInventoryRepository> {
  window.history.replaceState({}, '', path);
  const repository = new SeededInventoryRepository(structuredClone(seed));
  component = mount(InventoryWorkspaceApp, {
    target: document.body,
    props: {
      repository,
      initialData: await repository.loadWorkspace(),
      onSignOut: () => {}
    }
  });
  return repository;
}

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

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
  window.history.replaceState({}, '', '/');
});

describe('InventoryWorkspaceApp route application', () => {
  it('canonicalizes inventory-only asset aliases after loading the asset detail', async () => {
    await mountWorkspace('/inventories/inventory-household/assets/asset-home');

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/assets/asset-home');
      expect(document.body.textContent).toContain('Passport');
    });
  });

  it('shows a calm unavailable state for an inventory outside the visible workspace', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/not-visible');

    await waitFor(() => {
      expect(document.body.textContent).toContain('Workspace unavailable');
      expect(document.body.textContent).toContain('That inventory is not available in the current workspace.');
    });
  });

  it('normalizes unavailable asset action routes back to asset detail', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/assets/asset-archived/move?lifecycle=archived');

    await waitFor(() => {
      expect(document.body.textContent).toContain('Archived Passport');
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/assets/asset-archived');
    });
  });

  it('applies browser popstate route changes', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household');

    await waitFor(() => {
      expect(document.body.textContent).toContain('Home');
    });

    window.history.pushState({}, '', '/tenants/tenant-home/inventories/inventory-household/assets/asset-home');
    window.dispatchEvent(new PopStateEvent('popstate'));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Passport');
      expect(document.body.textContent).toContain('Blue folder');
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/assets/asset-home');
    });
  });
});
