import React from 'react';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { HomeDashboardQuery } from '../../application/home/HomeDashboardQuery';
import type { AssetCheckoutResult, HomeDashboardSnapshot, HomeDashboardSnapshotRepository } from '../../application/home/InventorySummaryRepository';
import { AssetCheckoutCommand } from '../../application/assets/AssetCheckoutCommand';
import { assetId, type AssetSummary } from '../../domain/assets/AssetSummary';
import { inventoryId, tenantId } from '../../domain/inventories/InventorySummary';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { MobileRenderHarness } from '../../test-support/render';
import { dispatchedActions, resetNavigation } from '../../test-support/navigation';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { AppFeedbackProvider } from '../feedback/AppFeedback';
import { HomeScreen } from './HomeScreen';

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
      role: 'owner', permissions: ['view', 'create_asset'], description: '', updatedAtLabel: 'Updated today',
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
  let returnResult: () => Promise<AssetCheckoutResult>;
  const settle = async () => { await h.run(() => new Promise(resolve => setTimeout(resolve, 10))); };
  const refresh = () => h.byType('ScrollView')!.props.refreshControl.props;
  async function render() {
    const command = new AssetCheckoutCommand({ returnAsset: async id => { returns.push(id); return returnResult(); } });
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant-home', inventoryId: 'inventory-home' })}>
      <AppFeedbackProvider><HomeScreen dashboardQuery={new HomeDashboardQuery(repository)} assetCheckoutCommand={command} /></AppFeedbackProvider>
    </MobileServerStateProvider>);
    await settle(); await settle();
  }
  beforeEach(() => {
    resetNavigation(); h = new MobileRenderHarness(); repository = new DashboardRepository(); client = createMobileQueryClient(); returns = [];
    returnResult = async () => ({ id: 'checkout-one', assetId: checkedOut.id, undoableOperationId: 'operation-one' });
  });
  afterEach(async () => { await h.unmount(); client.clear(); resetNavigation(); });

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
    expect(h.byText('Return details')).toBeDefined();
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
    expect(h.byText('Returning...')).toBeDefined(); expect(h.byText('Return details')).toBeUndefined();
    await h.run(() => pending.resolve({ id: 'checkout-one', assetId: checkedOut.id, undoableOperationId: 'operation-one' })); await settle();
    expect(h.byText('Return details')).toBeDefined();
  });
  it('shows return details before background reconciliation completes', async () => {
    await render(); const pending = deferred<HomeDashboardSnapshot>(); repository.load = () => pending.promise;
    await h.press(h.byLabel('Return Cordless drill')); await settle();
    expect(repository.reads).toBeGreaterThan(1); expect(h.byText('Return details')).toBeDefined();
    expect(h.byText('Cancel return')).toBeDefined(); expect(refresh().refreshing).toBe(false);
    await h.run(() => pending.resolve(snapshot([]))); await settle();
  });
  it('still reconciles a successful return without undo and explains the limitation', async () => {
    returnResult = async () => ({ id: 'checkout-one', assetId: checkedOut.id });
    await render(); const pending = deferred<HomeDashboardSnapshot>(); repository.load = () => pending.promise;
    await h.press(h.byLabel('Return Cordless drill')); await settle();
    expect(repository.reads).toBeGreaterThan(1); expect(h.byText('Return details')).toBeDefined();
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
});
