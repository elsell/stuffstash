import { useEffect, useState } from 'react';
import { InventoryAssetsQuery } from '../src/application/assets/InventoryAssetsQuery';
import { assetId } from '../src/domain/assets/AssetSummary';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { InventoryAssetsRouteScreen } from '../src/ui/screens/InventoryAssetsRouteScreen';

export function InventoryCollectionFixture() {
  const [fixture] = useState(() => ({
    client: createMobileQueryClient(),
    query: new InventoryAssetsQuery({ getInventoryAssetsSnapshot: async () => ({
      inventoryName: 'Household inventory',
      assets: Array.from({ length: 24 }, (_, index) => ({
        id: assetId(`collection-${index + 1}`), title: `Inventory item ${index + 1}`,
        kind: 'item' as const, lifecycleState: 'active' as const, description: '',
        locationLabel: '', locationTrail: [], parentLocationTrail: [],
        updatedAtLabel: 'Updated today', hasPhoto: false,
        tags: [{ id: `tag-${index}`, key: `collection-tag-${index}`,
          displayName: index === 23 ? 'Final inventory tag' : 'Household' }]
      }))
    }) })
  }));
  useEffect(() => () => fixture.client.clear(), [fixture]);
  return <MobileServerStateProvider client={fixture.client} scopeId="inventory-clearance"
    loadInventoryScope={async () => ({ tenantId: 'audit-tenant', inventoryId: 'audit-inventory' })}>
    <InventoryAssetsRouteScreen inventoryAssetsQuery={fixture.query} />
  </MobileServerStateProvider>;
}
