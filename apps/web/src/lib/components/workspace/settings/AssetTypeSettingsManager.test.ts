import { expect, it, vi } from 'vitest';
import { mount, unmount } from 'svelte';
import { writable, fromStore } from 'svelte/store';
import { SeededInventoryRepository } from '$lib/adapters/memory/seededInventoryRepository';
import type { Tenant, Inventory, CustomAssetType } from '$lib/domain/inventory';
import AssetTypeSettingsManager from './AssetTypeSettingsManager.svelte';

it('resets discarded expiration changes when reopening the same type', async () => {
  const tenant: Tenant = { id: 'home', name: 'Home', access: { relationship: 'owner', permissions: ['view', 'configure'] } };
  const inventory: Inventory = { id: 'main', tenantId: 'home', name: 'Main', access: { relationship: 'owner', permissions: ['view', 'configure'] } };
  const type: CustomAssetType = { id: 'medicine', tenantId: 'home', inventoryId: 'main', scope: 'inventory', key: 'medicine', displayName: 'Medicine', description: '', lifecycleState: 'active', expirationEnabled: false };
  const repository = new SeededInventoryRepository({ principal: { id: 'owner', email: 'owner@example.test' }, tenants: [tenant], inventories: [inventory], assets: [], customAssetTypes: [type], customFieldDefinitions: [] });
  const action = writable<'edit' | null>('edit');
  const reactiveAction = fromStore(action);
  const component = mount(AssetTypeSettingsManager, { target: document.body, props: { level: 'inventory', tenant, inventory, repository, observer: { record() {} }, canonicalItems: [type], resourceId: type.id, get action() { return reactiveAction.current; }, onNavigate() { action.set(null); }, onSchemaChange() {}, async onPermissionDenied() {} } });
  try {
    const checkbox = () => document.querySelector<HTMLInputElement>('#asset-type-expiration');
    await vi.waitFor(() => expect(checkbox()).not.toBeNull());
    checkbox()!.click();
    await vi.waitFor(() => expect(checkbox()!.checked).toBe(true));
    action.set(null);
    await vi.waitFor(() => expect(checkbox()).toBeNull());
    action.set('edit');
    await vi.waitFor(() => expect(checkbox()?.checked).toBe(false));
  } finally { await unmount(component); document.body.innerHTML = ''; }
});
