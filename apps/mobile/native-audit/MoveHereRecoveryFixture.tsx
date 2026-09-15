import { useEffect, useState } from 'react';
import { AssetMoveHereSheetRouteScreen } from '../src/ui/screens/AssetNativeActionSheetScreens';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { AssetCoreQuery } from '../src/application/assets/AssetCoreQuery';
import { assetId } from '../src/domain/assets/AssetSummary';
import { tenantId, inventoryId } from '../src/domain/inventories/InventorySummary';

export function MoveHereRecoveryFixture() {
  const [fixture] = useState(() => {
    const reads = new Map<string, number>();
    const client = createMobileQueryClient();
    const defaults = client.getDefaultOptions();
    client.setDefaultOptions({ ...defaults, queries: { ...defaults.queries, retry: false } });
    return { client,
      lookup: { execute: async (query: string) => {
        const count = (reads.get(query) ?? 0) + 1; reads.set(query, count);
        if (count === 1) throw new Error('Audit suggestions unavailable');
        return [{ id: 'audit-tent', title: 'Audit tent', kind: 'item' as const, subtitle: 'Item', pathLabel: 'Garage', selectionHint: 'Item', willPromoteToContainer: false }];
      } },
      core: new AssetCoreQuery({ getAssetCore: async () => ({
        tenantId: tenantId('audit-tenant'), inventoryId: inventoryId('audit-inventory'), permissions: ['edit_asset'], revision: 'audit',
        asset: { id: assetId('audit-box'), title: 'Camping box', kind: 'container', lifecycleState: 'active', description: '',
          locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false }
      }) })
    };
  });
  useEffect(() => () => fixture.client.clear(), [fixture]);
  return <MobileServerStateProvider client={fixture.client} scopeId="audit" loadInventoryScope={async () => ({ tenantId: 'audit-tenant', inventoryId: 'audit-inventory' })}>
    <AssetMoveHereSheetRouteScreen assetId="audit-box" assetCoreQuery={fixture.core} parentLookupQuery={fixture.lookup}
      moveAssetCommand={{ execute: async () => { throw new Error('Audit does not move assets'); } }} />
  </MobileServerStateProvider>;
}
