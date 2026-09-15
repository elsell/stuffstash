import React from 'react';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { HomeDashboardQuery } from '../../application/home/HomeDashboardQuery';
import type { AssetCheckoutResult, HomeDashboardSnapshot, HomeDashboardSnapshotRepository } from '../../application/home/InventorySummaryRepository';
import { AssetCheckoutCommand } from '../../application/assets/AssetCheckoutCommand';
import { assetId, type AssetSummary } from '../../domain/assets/AssetSummary';
import { inventoryId, tenantId } from '../../domain/inventories/InventorySummary';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { MobileRenderHarness } from '../../test-support/render';
import { dispatchedActions, resetNavigation, setScreenFocused } from '../../test-support/navigation';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { AppFeedbackProvider } from '../feedback/AppFeedback';
import { HomeReturnTaskProvider, useHomeReturnTask } from '../navigation/HomeReturnTaskPresentation';
import HomeReturnDetailsRoute from '../../app/home-return-details';
import { router } from 'expo-router';
import { HomeScreen } from './HomeScreen';

function ReturnRouteHost() { return useHomeReturnTask() ? <HomeReturnDetailsRoute /> : null; }

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>(done => { resolve = done; });
  return { promise, resolve };
}
const recent: AssetSummary = {
  id: assetId('asset-recent'), title: 'Recent bowl', kind: 'item', lifecycleState: 'active',
  description: '', locationLabel: 'Cabinet', locationTrail: ['Kitchen', 'Cabinet'],
  parentLocationTrail: [{ id: assetId('asset-kitchen'), title: 'Kitchen' }],
  updatedAtLabel: 'Updated today', hasPhoto: false,
  tags: [{ id: 'tag-kitchen', key: 'kitchen', displayName: 'Kitchen supplies' }]
};
const checkedOut: AssetSummary = {
  ...recent, id: assetId('asset-checked-out'), title: 'Cordless drill',
  tags: [{ id: 'tag-tools', key: 'tools', displayName: 'Power tools' }],
  currentCheckout: { id: 'checkout-one', state: 'open', checkedOutAt: '2026-09-15T00:00:00Z', checkedOutByPrincipalId: 'principal' }
};
function snapshot(checkedOutAssets: readonly AssetSummary[] = [checkedOut]): HomeDashboardSnapshot {
  return { checkedOutAssets, workspace: {
    tenants: [{ id: tenantId('tenant-home'), name: 'Home' }], defaultInventoryId: inventoryId('inventory-home'),
    inventories: [{ id: inventoryId('inventory-home'), tenantId: tenantId('tenant-home'), name: 'Home Inventory',
      role: 'owner', permissions: ['view', 'create_asset', 'edit_asset'], description: '', updatedAtLabel: 'Updated today',
      locationCount: 1, locations: [], assets: [recent] }]
  } };
}
class DashboardRepository implements HomeDashboardSnapshotRepository {
  reads = 0;
  load: () => Promise<HomeDashboardSnapshot> = async () => snapshot();
  async getHomeDashboardSnapshot() { this.reads++; return this.load(); }
}

describe('Home interactions through mounted components', () => {
  let h: MobileRenderHarness;
  let repository: DashboardRepository;
  let client: ReturnType<typeof createMobileQueryClient>;
  let returns: string[];
  let updates: string[];
  let updateDetails: () => Promise<AssetCheckoutResult>;
  let undos: string[];
  let returnResult: () => Promise<AssetCheckoutResult>;
  const settle = async () => { await h.run(() => new Promise(resolve => setTimeout(resolve, 10))); };
  const refresh = () => h.byType('ScrollView')!.props.refreshControl.props;
  async function render() {
    const command = new AssetCheckoutCommand({
      returnAsset: async id => { returns.push(id); return returnResult(); },
      updateReturnedCheckoutDetails: async (_id, _checkout, input) => { updates.push(input?.details ?? ''); return updateDetails(); },
      undoInventoryOperation: async id => { undos.push(id); }
    });
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant-home', inventoryId: 'inventory-home' })}>
      <AppFeedbackProvider><HomeReturnTaskProvider><HomeScreen dashboardQuery={new HomeDashboardQuery(repository)} assetCheckoutCommand={command} /><ReturnRouteHost /></HomeReturnTaskProvider></AppFeedbackProvider>
    </MobileServerStateProvider>);
    await settle(); await settle();
  }
  beforeEach(() => {
    resetNavigation(); setScreenFocused(true); h = new MobileRenderHarness(); repository = new DashboardRepository(); client = createMobileQueryClient(); returns = []; updates = []; undos = [];
    updateDetails = async () => ({ id: 'checkout-one', assetId: checkedOut.id });
    returnResult = async () => ({ id: 'checkout-one', assetId: checkedOut.id, undoableOperationId: 'operation-one' });
  });
  afterEach(async () => { await h.unmount(); client.clear(); resetNavigation(); setScreenFocused(true); });

  it('opens recent and checked-out items and their parent location', async () => {
    await render();
    await h.press(h.byLabel('Open asset Recent bowl'));
    await h.press(h.byLabel('Open asset Cordless drill'));
    await h.press(h.byLabel('Open location Kitchen'));
    expect(dispatchedActions()).toEqual(['asset-recent', 'asset-checked-out', 'asset-kitchen'].map(id => ({ type: 'push', href: { pathname: '/assets/[assetId]', params: { assetId: id } } })));
  });
  it('names inventory context and opens inventory, account, and Add', async () => {
    await render();
    await h.press(h.byLabel('Current inventory Home Inventory, tenant Home. Switch inventory'));
    await h.press(h.byLabel('Open account and settings'));
    await h.press(h.byLabel('Add an asset'));
    expect(dispatchedActions()).toEqual(['/tenant-switcher', '/settings', '/add'].map(href => ({ type: 'push', href })));
  });
  it('retains location breadcrumbs without a permanent Locations section', async () => {
    await render();
    expect(h.byLabel('Open location Kitchen')).toBeDefined();
    expect(h.byText('Locations')).toBeUndefined();
    expect(h.byText('Kitchen supplies')).toBeUndefined();
    expect(h.byText('Power tools')).toBeUndefined();
  });
  it('exposes update context and section navigation', async () => {
    await render();
    expect(h.byText('Updated today')).toBeDefined();
    expect(h.byType('ScrollView')?.props.contentInsetAdjustmentBehavior).toBe('automatic');
    expect(h.byType('ScrollView')?.props.horizontal).not.toBe(true);
    expect(h.byText('Recently changed')?.props.accessibilityRole).toBe('header');
    expect(h.allByType('Text').some(node => node.children.includes('Checked out') && node.props.accessibilityRole === 'header')).toBe(true);
    await h.press(h.byLabel('View all recently changed assets'));
    await h.press(h.byLabel('View all checked-out assets'));
    expect(dispatchedActions()).toEqual([{ type: 'push', href: '/assets' }, { type: 'navigate', href: { pathname: '/search', params: { checkoutState: 'checked_out' } } }]);
  });
  it('omits the checked-out section when empty', async () => {
    repository.load = async () => snapshot([]); await render();
    expect(h.byLabel('View all checked-out assets')).toBeUndefined();
    expect(h.byText('Checked out')).toBeUndefined();
    expect(h.byText('Nothing checked out.')).toBeUndefined();
  });
  it('returns the named asset through its command', async () => {
    await render(); await h.press(h.byLabel('Return Cordless drill')); await settle();
    expect(returns).toEqual(['asset-checked-out']);
    expect(h.byLabel('Optional return details')).toBeDefined();
  });
  it('does not display a pull indicator during background refresh', async () => {
    await render(); const pending = deferred<HomeDashboardSnapshot>(); repository.load = () => pending.promise;
    let finished!: Promise<void>;
    await h.run(() => { finished = client.invalidateQueries(); }); await settle();
    expect(repository.reads).toBeGreaterThan(1); expect(refresh().refreshing).toBe(false);
    await h.run(async () => { pending.resolve(snapshot()); await finished; });
  });
  it('disables Return while the command is pending', async () => {
    const pending = deferred<AssetCheckoutResult>(); returnResult = () => pending.promise;
    await render(); await h.press(h.byLabel('Return Cordless drill'));
    expect(h.byLabel('Return Cordless drill')?.props.disabled).toBe(true);
    expect(h.byText('Returning...')).toBeDefined(); expect(h.byLabel('Optional return details')).toBeUndefined();
    await h.run(() => pending.resolve({ id: 'checkout-one', assetId: checkedOut.id, undoableOperationId: 'operation-one' })); await settle();
    expect(h.byLabel('Optional return details')).toBeDefined();
  });
  it('shows return details before background reconciliation completes', async () => {
    await render(); const pending = deferred<HomeDashboardSnapshot>(); repository.load = () => pending.promise;
    await h.press(h.byLabel('Return Cordless drill')); await settle();
    expect(repository.reads).toBeGreaterThan(1); expect(h.byLabel('Optional return details')).toBeDefined();
    expect(h.byText('Cancel return')).toBeDefined(); expect(refresh().refreshing).toBe(false);
    await h.run(() => pending.resolve(snapshot([]))); await settle();
  });
  it('still reconciles a successful return without undo and explains the limitation', async () => {
    returnResult = async () => ({ id: 'checkout-one', assetId: checkedOut.id });
    await render(); const pending = deferred<HomeDashboardSnapshot>(); repository.load = () => pending.promise;
    await h.press(h.byLabel('Return Cordless drill')); await settle();
    expect(repository.reads).toBeGreaterThan(1); expect(h.byLabel('Optional return details')).toBeDefined();
    expect(h.byText('Return completed without undo')).toBeDefined(); expect(h.byText('Close')).toBeDefined();
    expect(h.byText('Cancel return')).toBeUndefined();
    await h.run(() => pending.resolve(snapshot([]))); await settle();
  });
  it('recovers an initial load failure using Retry', async () => {
    repository.load = async () => { throw new Error('Connection failed'); }; await render();
    expect(h.byLabel('Retry loading Home')).toBeDefined();
    repository.load = async () => snapshot(); await h.press(h.byLabel('Retry loading Home')); await settle();
    expect(h.byLabel('Open asset Recent bowl')).toBeDefined(); expect(h.byLabel('Retry loading Home')).toBeUndefined();
  });
  it('rejects repeated Return callbacks before rendering disabled state and while details are open', async () => {
    const pending = deferred<AssetCheckoutResult>(); returnResult = () => pending.promise;
    await render(); const press = h.byLabel('Return Cordless drill')!.props.onPress;
    await h.run(() => { press(); press(); });
    expect(returns).toEqual(['asset-checked-out']);
    await h.run(() => pending.resolve({ id: 'checkout-one', assetId: checkedOut.id, undoableOperationId: 'operation-one' })); await settle();
    await h.run(press);
    expect(returns).toEqual(['asset-checked-out']);
  });
  it('reconciles but does not open details when Return finishes after blur and refocus', async () => {
    const pending = deferred<AssetCheckoutResult>(); returnResult = () => pending.promise;
    await render(); await h.press(h.byLabel('Return Cordless drill'));
    await h.run(() => setScreenFocused(false)); await h.run(() => setScreenFocused(true));
    await h.run(() => pending.resolve({ id: 'checkout-one', assetId: checkedOut.id, undoableOperationId: 'operation-one' })); await settle();
    expect(repository.reads).toBeGreaterThan(1);
    expect(h.byLabel('Optional return details')).toBeUndefined();
    expect(h.byLabel('Return Cordless drill')?.props.disabled).toBe(true);
  });

  it('guards Save and Cancel together, preserves failed details, and permits retry', async () => {
    await render(); await h.press(h.byLabel('Return Cordless drill')); await settle();
    await h.changeText(h.byType('TextInput'), 'All accessories included');
    const pending = deferred<AssetCheckoutResult>();
    updateDetails = async () => { await pending.promise; throw new Error('Try again'); };
    const save = h.byText('Save')!.parent!.props.onPress;
    const cancel = h.byText('Cancel return')!.parent!.props.onPress;
    await h.run(() => { save(); save(); cancel(); });
    expect(updates).toEqual(['All accessories included']); expect(undos).toEqual([]);
    await h.run(() => pending.resolve({ id: 'checkout-one', assetId: checkedOut.id })); await settle();
    expect(h.byLabel('Optional return details')).toBeDefined();
    expect(h.byText('Could not save return details')).toBeDefined();
    expect(h.byLabel('Return details error')).toBeDefined();
    updateDetails = async () => ({ id: 'checkout-one', assetId: checkedOut.id });
    await h.run(save); await settle();
    expect(updates).toEqual(['All accessories included', 'All accessories included']);
    expect(h.byLabel('Optional return details')).toBeUndefined();
    await h.press(h.byLabel('Return Cordless drill'));
    expect(returns).toHaveLength(1);
  });
  it('presents return details as a native task and routes Back through undo', async () => {
    await render(); await h.press(h.byLabel('Return Cordless drill')); await settle();
    expect(dispatchedActions()).toContainEqual({ type: 'push', href: '/home-return-details' });
    expect(h.byLabel('Optional return details')).toBeDefined();
    await h.run(() => router.back()); await settle();
    expect(undos).toEqual(['operation-one']);
    expect(h.byLabel('Optional return details')).toBeUndefined();
  });
  it('keeps the draft but rejects stale commands when permission is revoked', async () => {
    await render(); await h.press(h.byLabel('Return Cordless drill')); await settle();
    await h.changeText(h.byLabel('Optional return details'), 'Retain this draft');
    const save = h.byLabel('Save')!.props.onPress;
    const cancel = h.byLabel('Cancel return')!.props.onPress;
    repository.load = async () => {
      const value = snapshot();
      return { ...value, workspace: { ...value.workspace, inventories: value.workspace.inventories.map(inventory => ({ ...inventory, permissions: ['view'] })) } };
    };
    await h.run(() => client.invalidateQueries()); await settle();
    await h.run(() => { save(); cancel(); }); await settle();
    expect(updates).toEqual([]); expect(undos).toEqual([]);
    expect(h.byLabel('Optional return details')).toBeDefined();
    expect(h.byLabel('Optional return details')?.props.editable).toBe(false);
    expect(h.byLabel('Save')).toBeUndefined();
    await h.press(h.byLabel('Close')); await settle();
    expect(h.byLabel('Optional return details')).toBeUndefined();
  });
  it('rejects callbacks from an earlier return editor after another Return', async () => {
    await render(); await h.press(h.byLabel('Return Cordless drill')); await settle();
    const oldSave = h.byLabel('Save')!.props.onPress;
    await h.press(h.byLabel('Cancel return')); await settle();
    await h.press(h.byLabel('Return Cordless drill')); await settle();
    expect(returns).toHaveLength(2);
    expect(h.byLabel('Optional return details')).toBeDefined();
    expect(dispatchedActions().filter(action => action.type === 'push' && action.href === '/home-return-details')).toHaveLength(2);
    await h.run(oldSave); await settle();
    expect(updates).toEqual([]);
    expect(h.byLabel('Optional return details')).toBeDefined();
  });
  it('permits a restored checkout after Cancel return', async () => {
    await render(); await h.press(h.byLabel('Return Cordless drill')); await settle();
    await h.press(h.byText('Cancel return')!.parent ?? undefined); await settle();
    expect(undos).toEqual(['operation-one']); expect(h.byLabel('Optional return details')).toBeUndefined();
    await h.press(h.byLabel('Return Cordless drill')); await settle();
    expect(returns).toHaveLength(2);
  });
  it('permits a later checkout after reconciliation removes the original card', async () => {
    await render(); await h.press(h.byLabel('Return Cordless drill')); await settle();
    repository.load = async () => snapshot([]);
    await h.press(h.byText('Save')!.parent ?? undefined); await settle();
    expect(h.byLabel('Return Cordless drill')).toBeUndefined();
    repository.load = async () => snapshot();
    await h.run(() => client.invalidateQueries()); await settle();
    await h.press(h.byLabel('Return Cordless drill')); await settle();
    expect(returns).toHaveLength(2);
  });

  it('keeps failed reconciliation silent after leaving Home', async () => {
    const pending = deferred<AssetCheckoutResult>(); returnResult = () => pending.promise;
    await render(); await h.press(h.byLabel('Return Cordless drill'));
    await h.run(() => setScreenFocused(false));
    repository.load = async () => { throw new Error('Unavailable'); };
    await h.run(() => pending.resolve({ id: 'checkout-one', assetId: checkedOut.id })); await settle();
    expect(h.byText('Could not refresh Home')).toBeUndefined();
    expect(h.byLabel('Optional return details')).toBeUndefined();
  });
  it('permits a new checkout of the same asset without an intermediate empty snapshot', async () => {
    await render(); await h.press(h.byLabel('Return Cordless drill')); await settle();
    await h.press(h.byText('Save')!.parent ?? undefined); await settle();
    repository.load = async () => snapshot([{ ...checkedOut, currentCheckout: { ...checkedOut.currentCheckout!, id: 'checkout-two' } }]);
    await h.run(() => client.invalidateQueries()); await settle();
    expect(h.byLabel('Return Cordless drill')?.props.disabled).toBe(false);
    await h.press(h.byLabel('Return Cordless drill')); await settle();
    expect(returns).toHaveLength(2);
  });

  it.each([['view'], ['view', 'create_asset']])('keeps checkout status but hides Return without edit permission: %j', async (...permissions) => {
    repository.load = async () => {
      const value = snapshot();
      return { ...value, workspace: { ...value.workspace, inventories: value.workspace.inventories.map(inventory => ({ ...inventory, permissions })) } };
    };
    await render();
    expect(h.byLabel('Open asset Cordless drill')).toBeDefined();
    expect(h.byLabel('View all checked-out assets')).toBeDefined();
    expect(h.byLabel('Return Cordless drill')).toBeUndefined(); expect(returns).toEqual([]);
  });
  it('rejects a stale Return callback after permission revocation', async () => {
    await render(); const press = h.byLabel('Return Cordless drill')!.props.onPress;
    repository.load = async () => {
      const value = snapshot();
      return { ...value, workspace: { ...value.workspace, inventories: value.workspace.inventories.map(inventory => ({ ...inventory, permissions: ['view'] })) } };
    };
    await h.run(() => client.invalidateQueries()); await settle();
    expect(h.byLabel('Return Cordless drill')).toBeUndefined();
    await h.run(press); expect(returns).toEqual([]);
  });

  it('offers Return with edit permission even without create permission', async () => {
    repository.load = async () => {
      const value = snapshot();
      return { ...value, workspace: { ...value.workspace, inventories: value.workspace.inventories.map(inventory => ({ ...inventory, permissions: ['view', 'edit_asset'] })) } };
    };
    await render();
    expect(h.byLabel('Add an asset')).toBeUndefined();
    await h.press(h.byLabel('Return Cordless drill')); await settle();
    expect(returns).toEqual(['asset-checked-out']);
  });

});
