import { useEffect, useState } from 'react';
import { Image } from 'react-native';
import { useLocalSearchParams } from 'expo-router';
import { SearchScreen } from '../src/ui/screens/SearchScreen';
import { parseBrowseRouteParams } from '../src/ui/screens/BrowseRouteParams';
import { SearchAssetsQuery } from '../src/application/search/SearchAssetsQuery';
import { InventoryMapQuery } from '../src/application/assets/InventoryMapQuery';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { assetId, type AssetSummary } from '../src/domain/assets/AssetSummary';
import { inventoryId, tenantId } from '../src/domain/inventories/InventorySummary';

const assets: readonly AssetSummary[] = ['Garage', 'Kitchen', 'Camping tent', 'Toolbox', 'Coffee grinder', 'Garden hose', 'Picnic blanket', 'Travel bag', 'Lantern', 'Camp stove', 'Sleeping bag', 'Bicycle'].map((title, index) => ({
  id: assetId(`journey-${index}`), title, description: '', kind: index < 2 ? 'location' : 'item',
  lifecycleState: 'active', locationLabel: '', locationTrail: [], parentLocationTrail: [],
  updatedAtLabel: 'Updated today', hasPhoto: false,
  ...(index > 1 ? { parentAssetId: assetId('journey-0') } : {})
}));

/** Actual Browse screen and native header, with isolated read-only fixture data. */
export function BrowseJourneyFixture() {
  const params = useLocalSearchParams();
  const mixedPhotos = params.photoMix === 'true';
  const dense = params.dense === 'true';
  return <BrowseJourneyContent key={`${mixedPhotos}:${dense}`} params={params} mixedPhotos={mixedPhotos} dense={dense} />;
}

function BrowseJourneyContent({ params, mixedPhotos, dense }: {
  readonly params: ReturnType<typeof useLocalSearchParams>;
  readonly mixedPhotos: boolean;
  readonly dense: boolean;
}) {
  const [fixture] = useState(() => {
    const sourceAssets = dense ? [...assets, ...Array.from({ length: 24 }, (_, index) => ({
      ...assets[2]!, id: assetId(`journey-extra-${index}`), title: `Stored item ${index + 13}`
    }))] : assets;
    const fixtureAssets = mixedPhotos ? sourceAssets.map((asset, index) => ({
      ...asset,
      hasPhoto: index === 0 || index === 2,
      ...(index === 0 ? { photo: { uri: Image.resolveAssetSource(require('../assets/brand/stuff-stash-glyph.png')).uri } } : {})
    })) : sourceAssets;
    return {
      client: createMobileQueryClient(),
      search: new SearchAssetsQuery({ browseAssets: async input => ({
        assets: fixtureAssets.filter(asset => asset.title.toLowerCase().includes(input.query.toLowerCase())), hasMore: false
      }) }),
      map: new InventoryMapQuery({ listActiveInventoryMapAssets: async () => ({
        sessionScopeId: 'audit', tenantId: tenantId('audit-tenant'), inventoryId: inventoryId('audit-inventory'),
        inventoryName: 'Main Inventory', permissions: ['view', 'create_asset'], assets: fixtureAssets
      }) })
    };
  });
  useEffect(() => () => fixture.client.clear(), [fixture]);
  return <MobileServerStateProvider client={fixture.client} scopeId="audit" loadInventoryScope={async () => ({ tenantId: 'audit-tenant', inventoryId: 'audit-inventory' })}>
    <SearchScreen {...parseBrowseRouteParams(params)} searchAssetsQuery={fixture.search} inventoryMapQuery={fixture.map}
      inventoryContextQuery={{ execute: async () => ({ inventoryName: 'Main Inventory', canAdd: true }) }}
      inventoryAssetTagsQuery={{ execute: async () => [] }}
      locationsQuery={{ execute: async () => ({ inventoryName: 'Main Inventory', tenantName: 'Home', canAdd: true, locations: [] }) }} />
  </MobileServerStateProvider>;
}
