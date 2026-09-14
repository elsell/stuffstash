import { expect, it } from 'vitest';
import { InventoryAssetsQuery } from '../../application/assets/InventoryAssetsQuery';
import { LocationAssetsQuery } from '../../application/locations/LocationAssetsQuery';
import { LocationsQuery } from '../../application/locations/LocationsQuery';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { InventoryAssetsRouteScreen } from './InventoryAssetsRouteScreen';
import { LocationAssetsRouteScreen } from './LocationAssetsRouteScreen';
import { LocationsScreen } from './LocationsScreen';

for (const surface of ['inventory', 'location-content', 'locations'] as const) {
  it(`retries an initial ${surface} failure in place without a pull indicator`, async () => {
    const h = new MobileRenderHarness(); const client = createMobileQueryClient();
    client.setDefaultOptions({ queries: { retry: false } });
    let calls = 0; let finish: (() => void) | undefined;
    const read = async () => {
      calls++;
      if (calls === 1) throw new Error('Connection unavailable');
      await new Promise<void>(resolve => { finish = resolve; });
      if (calls === 2) throw new Error('Connection still unavailable');
    };
    const screen = surface === 'inventory'
      ? <InventoryAssetsRouteScreen inventoryAssetsQuery={new InventoryAssetsQuery({ getInventoryAssetsSnapshot: async () => { await read(); return { inventoryName: 'Home', assets: [] }; } })} />
      : surface === 'location-content'
        ? <LocationAssetsRouteScreen locationId="cabinet" locationAssetsQuery={new LocationAssetsQuery({ getLocationAssetsSnapshot: async id => { expect(id).toBe('cabinet'); await read(); return { locationId: id, locationTitle: 'Cabinet', inventoryName: 'Home', assets: [] }; } })} />
        : <LocationsScreen locationsQuery={new LocationsQuery({ getLocationsSnapshot: async () => { await read(); return { canAdd: true, tenantName: 'Household', inventoryName: 'Home', locations: [] }; } })} />;
    const settle = () => h.run(() => new Promise(resolve => setTimeout(resolve, 20)));
    try {
      await h.render(<MobileServerStateProvider client={client} scopeId="session" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>{screen}</MobileServerStateProvider>);
      await settle(); await settle();
      expect(h.byText('Could not load')).toBeDefined();
      await h.press(h.byLabel('Retry'));
      await settle();
      expect(calls).toBe(2);
      expect(h.byType('FlatList')).toBeUndefined();
      await h.run(() => finish?.()); await settle();
      expect(h.byText('Could not load')).toBeDefined();
      await h.press(h.byLabel('Retry')); await settle();
      expect(calls).toBe(3);
      await h.run(() => finish?.()); await settle();
      expect(h.byText('Could not load')).toBeUndefined();
      expect(h.byType('FlatList')).toBeDefined();
      expect(h.byType('FlatList')?.props.refreshing ?? h.byType('RefreshControl')?.props.refreshing ?? false).toBe(false);
    } finally { await h.unmount(); client.clear(); }
  });
}

it('recovers an initial inventory-scope failure before loading resource rows', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  client.setDefaultOptions({ queries: { retry: false } });
  let scopeAvailable = false; let resourceReads = 0;
  const query = new InventoryAssetsQuery({ getInventoryAssetsSnapshot: async () => {
    resourceReads++; return { inventoryName: 'Home', assets: [] };
  } });
  const settle = () => h.run(() => new Promise(resolve => setTimeout(resolve, 20)));
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="session" loadInventoryScope={async () => {
      if (!scopeAvailable) throw new Error('Inventory unavailable');
      return { tenantId: 'tenant', inventoryId: 'inventory' };
    }}><InventoryAssetsRouteScreen inventoryAssetsQuery={query} /></MobileServerStateProvider>);
    await settle(); await settle();
    expect(h.byText('Could not load')).toBeDefined();
    expect(resourceReads).toBe(0);
    scopeAvailable = true;
    await h.press(h.byLabel('Retry')); await settle(); await settle();
    expect(h.byText('Could not load')).toBeUndefined();
    expect(h.byType('FlatList')).toBeDefined();
    expect(resourceReads).toBeGreaterThan(0);
    expect(h.byType('FlatList')?.props.refreshing).toBe(false);
  } finally { await h.unmount(); client.clear(); }
});
