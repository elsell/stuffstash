import { expect, it } from 'vitest';
import { InventoryAssetsQuery } from '../../application/assets/InventoryAssetsQuery';
import { LocationAssetsQuery } from '../../application/locations/LocationAssetsQuery';
import { LocationsQuery } from '../../application/locations/LocationsQuery';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { MobileRenderHarness } from '../../test-support/render';
import { setScreenFocused } from '../../test-support/navigation';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { AppFeedbackProvider } from '../feedback/AppFeedback';
import { InventoryAssetsRouteScreen } from './InventoryAssetsRouteScreen';
import { LocationAssetsRouteScreen } from './LocationAssetsRouteScreen';
import { LocationsScreen } from './LocationsScreen';
import { assetId, type AssetSummary } from '../../domain/assets/AssetSummary';
import { inventoryId } from '../../domain/inventories/InventorySummary';

const asset: AssetSummary = { id: assetId('bowl'), title: 'Stored bowl', kind: 'item', lifecycleState: 'active', description: '', locationLabel: 'Cabinet', locationTrail: ['Cabinet'], parentLocationTrail: [], updatedAtLabel: 'Today', hasPhoto: false, tags: [] };
const location = { id: assetId('cabinet'), inventoryId: inventoryId('inventory'), title: 'Stored cabinet', description: '', containedAssetCount: 1, recentAssetTitles: ['Stored bowl'], hasPhoto: false };

for (const surface of ['assets', 'location', 'locations'] as const) {
  for (const visit of ['current', 'departed', 'returned'] as const) {
    it(`${surface} refresh reports failure only in its ${visit} visit and retains content`, async () => {
      const h = new MobileRenderHarness(); const client = createMobileQueryClient();
      let finish!: () => void; let fail = false;
      const read = async () => { if (fail) { await new Promise<void>(resolve => { finish = resolve; }); throw new Error('Connection unavailable'); } };
      const screen = surface === 'assets'
        ? <InventoryAssetsRouteScreen inventoryAssetsQuery={new InventoryAssetsQuery({ getInventoryAssetsSnapshot: async () => { await read(); return { inventoryName: 'Home', assets: [asset] }; } })} />
        : surface === 'location'
          ? <LocationAssetsRouteScreen locationId="cabinet" locationAssetsQuery={new LocationAssetsQuery({ getLocationAssetsSnapshot: async id => { await read(); return { locationId: id, locationTitle: 'Cabinet', inventoryName: 'Home', assets: [asset] }; } })} />
          : <LocationsScreen locationsQuery={new LocationsQuery({ getLocationsSnapshot: async () => { await read(); return { canAdd: true, tenantName: 'Household', inventoryName: 'Home', locations: [location] }; } })} />;
      const settle = () => h.run(() => new Promise(resolve => setTimeout(resolve, 20)));
      const pull = () => { const list = h.byType('FlatList')!; return list.props.refreshControl?.props ?? list.props; };
      try {
        setScreenFocused(true);
        await h.render(<MobileServerStateProvider client={client} scopeId="session" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
          <AppFeedbackProvider>{screen}</AppFeedbackProvider>
        </MobileServerStateProvider>);
        await settle(); await settle();
        expect(h.byType('FlatList')).toBeDefined();
        expect(h.byText(surface === 'locations' ? 'Stored cabinet' : 'Stored bowl')).toBeDefined();
        fail = true;
        await h.run(() => pull().onRefresh()); await settle();
        expect(pull().refreshing).toBe(true);
        if (visit !== 'current') await h.run(() => setScreenFocused(false));
        if (visit === 'returned') await h.run(() => setScreenFocused(true));
        await h.run(() => finish()); await settle();
        expect(Boolean(h.byText(`Could not refresh ${surface}`))).toBe(visit === 'current');
        expect(h.byText(surface === 'locations' ? 'Stored cabinet' : 'Stored bowl')).toBeDefined();
        expect(h.byType('FlatList')).toBeDefined();
        expect(pull().refreshing).toBe(false);
        if (visit === 'current') {
          fail = false;
          await h.run(() => pull().onRefresh()); await settle();
          expect(pull().refreshing).toBe(false);
          expect(h.byType('FlatList')).toBeDefined();
        }
      } finally { await h.unmount(); client.clear(); setScreenFocused(true); }
    });
  }
}
