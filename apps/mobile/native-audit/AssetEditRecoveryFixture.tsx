import { useEffect, useState } from 'react';
import { AssetEditSheetRouteScreen } from '../src/ui/screens/AssetNativeActionSheetScreens';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { AssetCoreQuery } from '../src/application/assets/AssetCoreQuery';
import { assetId } from '../src/domain/assets/AssetSummary';
import { tenantId, inventoryId } from '../src/domain/inventories/InventorySummary';

/** Runner-only metadata failures; no production credentials or mutations. */
export function AssetEditRecoveryFixture() {
  const [fixture] = useState(() => {
    let typeReads = 0; let tagReads = 0;
    const client = createMobileQueryClient();
    const defaults = client.getDefaultOptions();
    client.setDefaultOptions({ ...defaults, queries: { ...defaults.queries, retry: false } });
    return { client,
      types: { execute: async () => { if (++typeReads === 1) throw new Error('Audit types unavailable'); return []; } },
      tags: { execute: async () => { if (++tagReads === 1) throw new Error('Audit tags unavailable'); return []; } },
      core: new AssetCoreQuery({ getAssetCore: async () => ({
        tenantId: tenantId('audit-tenant'), inventoryId: inventoryId('audit-inventory'), permissions: ['edit_asset'], revision: 'audit',
        asset: { id: assetId('audit-tent'), title: 'Audit tent', kind: 'item', lifecycleState: 'active', description: '',
          locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false }
      }) })
    };
  });
  useEffect(() => () => fixture.client.clear(), [fixture]);
  return <MobileServerStateProvider client={fixture.client} scopeId="audit" loadInventoryScope={async () => ({ tenantId: 'audit-tenant', inventoryId: 'audit-inventory' })}>
    <AssetEditSheetRouteScreen assetId="audit-tent" assetCoreQuery={fixture.core}
      inventoryAssetTypesQuery={fixture.types} inventoryAssetTagsQuery={fixture.tags}
      updateAssetCommand={{ execute: async () => { throw new Error('Audit does not save changes'); } }} />
  </MobileServerStateProvider>;
}
