import { useEffect, useState } from 'react';
import { AssetDetailRouteScreen } from '../src/ui/screens/AssetDetailRouteScreen';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { AssetCoreQuery } from '../src/application/assets/AssetCoreQuery';
import { AssetContentsQuery } from '../src/application/assets/AssetContentsQuery';
import { AssetPhotosQuery } from '../src/application/assets/AssetPhotosQuery';
import { PhotoSelectionQuery } from '../src/application/add/PhotoSelectionQuery';
import { assetId } from '../src/domain/assets/AssetSummary';
import { tenantId, inventoryId } from '../src/domain/inventories/InventorySummary';

const noMutation = async () => { throw new Error('This audit fixture does not mutate assets'); };

/** Real progressive detail route with isolated, independently failing queries. */
export function AssetRegionRecoveryFixture() {
  const [fixture] = useState(() => {
    const client = createMobileQueryClient();
    const defaults = client.getDefaultOptions();
    client.setDefaultOptions({ ...defaults, queries: { ...defaults.queries, retry: false } });
    const asset = { id: assetId('audit-place'), title: 'Audit place', kind: 'location' as const,
      lifecycleState: 'active' as const, description: '', locationLabel: '', locationTrail: [],
      parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
    let contentsReads = 0; let photoReads = 0;
    return { client,
      core: new AssetCoreQuery({ getAssetCore: async () => ({
        tenantId: tenantId('audit-tenant'), inventoryId: inventoryId('audit-inventory'),
        permissions: ['view'], revision: 'audit', asset
      }) }),
      contents: new AssetContentsQuery({ getAssetContents: async () => {
        if (++contentsReads === 1) throw new Error('Audit contents unavailable');
        return { asset, allAssets: [] };
      } }),
      photos: new AssetPhotosQuery({ getAssetPhotos: async () => {
        if (++photoReads === 1) throw new Error('Audit photos unavailable');
        return [];
      } }),
      selection: new PhotoSelectionQuery({ selectFromLibrary: noMutation, captureFromCamera: noMutation })
    };
  });
  useEffect(() => () => fixture.client.clear(), [fixture]);
  return <MobileServerStateProvider client={fixture.client} scopeId="audit" loadInventoryScope={async () => ({ tenantId: 'audit-tenant', inventoryId: 'audit-inventory' })}>
    <AssetDetailRouteScreen assetId="audit-place" assetCoreQuery={fixture.core}
      assetContentsQuery={fixture.contents} assetPhotosQuery={fixture.photos}
      photoSelectionQuery={fixture.selection} assetCheckoutCommand={{ execute: noMutation }}
      assetLifecycleCommand={{ execute: noMutation }} undoAssetEditCommand={{ execute: noMutation }}
      deleteAssetPhotoCommand={{ execute: noMutation }} addAssetPhotosCommand={{ execute: noMutation }} />
  </MobileServerStateProvider>;
}
