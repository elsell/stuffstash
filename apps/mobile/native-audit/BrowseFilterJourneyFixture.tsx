import { createContext, useContext, useEffect, useState, type PropsWithChildren } from 'react';
import { useLocalSearchParams } from 'expo-router';
import { SearchScreen } from '../src/ui/screens/SearchScreen';
import { BrowseFiltersRouteScreen } from '../src/ui/screens/BrowseFiltersRouteScreen';
import { ExpirationRouteScreen } from '../src/ui/expiration/ExpirationRouteScreen';
import { AssetDetailRouteScreen } from '../src/ui/screens/AssetDetailRouteScreen';
import { parseBrowseRouteParams } from '../src/ui/screens/BrowseRouteParams';
import { SearchAssetsQuery } from '../src/application/search/SearchAssetsQuery';
import { InventoryMapQuery } from '../src/application/assets/InventoryMapQuery';
import { AssetCoreQuery } from '../src/application/assets/AssetCoreQuery';
import { AssetContentsQuery } from '../src/application/assets/AssetContentsQuery';
import { AssetPhotosQuery } from '../src/application/assets/AssetPhotosQuery';
import { PhotoSelectionQuery } from '../src/application/add/PhotoSelectionQuery';
import { ExpirationWorkspaceQuery } from '../src/application/expiration/ExpirationWorkspaceQuery';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { assetId, type AssetSummary } from '../src/domain/assets/AssetSummary';
import { inventoryId, tenantId } from '../src/domain/inventories/InventorySummary';

const scope = { tenantId: 'filter-tenant', inventoryId: 'filter-inventory' };
const unsupported = async (): Promise<never> => { throw new Error('Read-only Browse journey'); };
const campingAssets: readonly AssetSummary[] = Array.from({ length: 24 }, (_, index) => ({
  id: assetId(`filter-item-${index + 1}`), title: `Camping item ${String(index + 1).padStart(2, '0')}`,
  description: 'Inventory for the camping trip', kind: 'item', lifecycleState: 'active',
  locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: 'Updated today', hasPhoto: false,
  expiration: { date: index % 2 === 0 ? '2026-01-01' : '2050-01-01', precision: 'day' },
  expirationContext: { state: index % 2 === 0 ? 'expired' : 'upcoming', trackingEnabled: true, advanceDays: 30, timezone: 'UTC' },
  ...(index % 4 === 0 ? { currentCheckout: { id: `checkout-${index}`, state: 'checked_out', checkedOutAt: '2026-01-01T00:00:00Z', checkedOutByPrincipalId: 'fixture-person' } } : {})
}));
const assets: readonly AssetSummary[] = [...campingAssets, {
  ...campingAssets[1]!, id: assetId('filter-kitchen-item'), title: 'Kitchen item',
  expiration: { date: '2026-01-01', precision: 'day' },
  expirationContext: { state: 'expired', trackingEnabled: true, advanceDays: 30, timezone: 'UTC' }
}];
function matches(asset: AssetSummary, query = '', checkoutState = 'any', kind = 'all') {
  return asset.title.toLowerCase().includes(query.toLowerCase()) && (kind === 'all' || kind === asset.kind)
    && (checkoutState === 'any' || (checkoutState === 'checked_out') === !!asset.currentCheckout);
}
export function createBrowseFilterJourney() {
  const client = createMobileQueryClient();
  const tags = { execute: async () => [] };
  return {
    client, tags,
    search: new SearchAssetsQuery({ browseAssets: async input => ({
      assets: input.lifecycleState === 'archived' || input.tagIds?.length ? [] : assets.filter(asset => matches(asset, input.query, input.checkoutState, input.kind)), hasMore: false
    }) }),
    map: new InventoryMapQuery({ listActiveInventoryMapAssets: async () => ({ sessionScopeId: 'filter-journey',
      tenantId: tenantId(scope.tenantId), inventoryId: inventoryId(scope.inventoryId), inventoryName: 'Camping inventory', permissions: ['view'], assets }) }),
    expiration: new ExpirationWorkspaceQuery({ list: async (tenant, inventory, filter) => {
      if (tenant !== scope.tenantId || inventory !== scope.inventoryId) throw new Error('Foreign fixture scope');
      const candidates = filter.tagIds?.length || filter.typeId || filter.locationId ? [] : assets.filter(asset => matches(asset, filter.query, filter.checkoutState, filter.kind));
      const items = candidates.filter(asset => filter.mode === 'all' || (filter.mode === 'expired') === (asset.expirationContext?.state === 'expired'));
      return { items, counts: { expired: candidates.filter(asset => asset.expirationContext?.state === 'expired').length,
        soon: candidates.filter(asset => asset.expirationContext?.state === 'upcoming').length, all: candidates.length }, timezone: 'UTC', hasMore: false, nextCursor: null };
    } }, { record: () => {} }),
    core: new AssetCoreQuery({ getAssetCore: async id => {
      const asset = assets.find(candidate => candidate.id === id);
      if (!asset) throw new Error('Unknown fixture asset');
      return { tenantId: tenantId(scope.tenantId), inventoryId: inventoryId(scope.inventoryId), permissions: ['view'], revision: '1', asset };
    } }),
    contents: new AssetContentsQuery({ getAssetContents: async core => ({ asset: core.asset, allAssets: [] }) }),
    photos: new AssetPhotosQuery({ getAssetPhotos: async () => [] }),
    selection: new PhotoSelectionQuery({ selectFromLibrary: unsupported, captureFromCamera: unsupported })
  };
}
const JourneyContext = createContext<ReturnType<typeof createBrowseFilterJourney> | undefined>(undefined);
export function BrowseFilterJourneyProvider({ children }: PropsWithChildren) {
  const [journey] = useState(createBrowseFilterJourney);
  useEffect(() => () => journey.client.clear(), [journey]);
  return <JourneyContext.Provider value={journey}>{children}</JourneyContext.Provider>;
}
function useJourney() {
  const journey = useContext(JourneyContext);
  if (!journey) throw new Error('Browse journey requires fixture provider');
  return journey;
}
function JourneyState({ children }: PropsWithChildren) {
  const journey = useJourney();
  return <MobileServerStateProvider client={journey.client} scopeId="filter-journey" loadInventoryScope={async () => scope}>{children}</MobileServerStateProvider>;
}
export function BrowseFilterJourneySearch() {
  const journey = useJourney(); const params = useLocalSearchParams();
  return <JourneyState><SearchScreen {...parseBrowseRouteParams(params)} searchAssetsQuery={journey.search} inventoryMapQuery={journey.map}
    inventoryContextQuery={{ execute: async () => ({ inventoryName: 'Camping inventory', canAdd: false }) }} inventoryAssetTagsQuery={journey.tags}
    locationsQuery={{ execute: async () => ({ inventoryName: 'Camping inventory', tenantName: 'Home', canAdd: false, locations: [] }) }} /></JourneyState>;
}
export function BrowseFilterJourneyFilters() {
  const journey = useJourney();
  return <JourneyState><BrowseFiltersRouteScreen inventoryAssetTagsQuery={journey.tags} /></JourneyState>;
}
export function BrowseFilterJourneyExpiration() {
  const journey = useJourney();
  return <JourneyState><ExpirationRouteScreen expirationWorkspaceQuery={journey.expiration} /></JourneyState>;
}
export function BrowseFilterJourneyDetail() {
  const journey = useJourney(); const params = useLocalSearchParams<{ assetId: string }>();
  return <JourneyState><AssetDetailRouteScreen assetId={params.assetId} assetCoreQuery={journey.core}
    assetContentsQuery={journey.contents} assetPhotosQuery={journey.photos} photoSelectionQuery={journey.selection}
    assetCheckoutCommand={{ execute: unsupported }} assetLifecycleCommand={{ execute: unsupported }}
    undoAssetEditCommand={{ execute: unsupported }} deleteAssetPhotoCommand={{ execute: unsupported }}
    addAssetPhotosCommand={{ execute: unsupported }} /></JourneyState>;
}
