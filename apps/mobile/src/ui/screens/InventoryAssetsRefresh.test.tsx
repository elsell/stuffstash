import { expect, it } from 'vitest';
import { InventoryAssetsQuery, type InventoryAssetsSnapshot } from '../../application/assets/InventoryAssetsQuery';
import { createMobileQueryClient, mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { MobileRenderHarness } from '../../test-support/render';
import { setScreenFocused } from '../../test-support/navigation';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { InventoryAssetsRouteScreen } from './InventoryAssetsRouteScreen';
import { AppFeedbackProvider } from '../feedback/AppFeedback';

it('reconciles inventory assets silently and shows refresh only for the current pull session', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  const scope = { tenantId: 'tenant', inventoryId: 'inventory' };
  const key = mobileQueryKeys.inventoryAssets('session', scope.tenantId, scope.inventoryId);
  const snapshot: InventoryAssetsSnapshot = { inventoryName: 'Home', assets: [] };
  const finishes: (() => void)[] = [];
  const query = new InventoryAssetsQuery({ getInventoryAssetsSnapshot: () => new Promise(resolve => finishes.push(() => resolve(snapshot))) });
  client.setQueryData(mobileQueryKeys.inventoryScope('session'), scope);
  client.setQueryData(key, snapshot);
  const settle = () => h.run(() => new Promise(resolve => setTimeout(resolve, 10)));
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="session" loadInventoryScope={async () => scope}>
      <AppFeedbackProvider><InventoryAssetsRouteScreen inventoryAssetsQuery={query} /></AppFeedbackProvider>
    </MobileServerStateProvider>);
    await h.run(() => { void client.invalidateQueries({ queryKey: key }); });
    await settle();
    expect(client.isFetching({ queryKey: key })).toBe(1);
    expect(h.byType('FlatList')?.props.refreshing).toBe(false);
    await h.run(() => finishes[0]()); await settle();
    await h.run(() => h.byType('FlatList')?.props.onRefresh()); await settle();
    expect(h.byType('FlatList')?.props.refreshing).toBe(true);
    await h.run(() => setScreenFocused(false));
    expect(h.byType('FlatList')?.props.refreshing).toBe(false);
    await h.run(() => finishes[1]()); await settle();
    await h.run(() => setScreenFocused(true));
    expect(h.byType('FlatList')?.props.refreshing).toBe(false);
  } finally { await h.unmount(); setScreenFocused(true); }
});
