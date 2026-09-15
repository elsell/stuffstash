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
  return <AssetDetailFixture mode="recovery" />;
}

export function AssetContentsSearchFixture() {
  return <AssetDetailFixture mode="search" />;
}

export function AssetDetailCommandsFixture() {
  return <AssetDetailFixture mode="commands" />;
}

function AssetDetailFixture({ mode }: { readonly mode: 'recovery' | 'search' | 'commands' }) {
  const [fixture] = useState(() => {
    const client = createMobileQueryClient();
    const defaults = client.getDefaultOptions();
    client.setDefaultOptions({ ...defaults, queries: { ...defaults.queries, retry: false } });
    const asset = { id: assetId('audit-place'), title: mode === 'commands' ? 'Garage shelves and seasonal storage' : 'Audit place',
      kind: mode === 'commands' ? 'container' as const : 'location' as const,
      lifecycleState: 'active' as const, description: '', locationLabel: '', locationTrail: [],
      parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
    let contentsReads = 0; let photoReads = 0;
    return { client,
      core: new AssetCoreQuery({ getAssetCore: async () => ({
        tenantId: tenantId('audit-tenant'), inventoryId: inventoryId('audit-inventory'),
        permissions: mode === 'commands' ? ['view', 'edit_asset', 'create_asset'] : ['view'], revision: 'audit', asset
      }) }),
      contents: new AssetContentsQuery({ getAssetContents: async () => {
        if (++contentsReads === 1 && mode === 'recovery') throw new Error('Audit contents unavailable');
        return { asset, allAssets: mode === 'search' ? Array.from({ length: 20 }, (_, index) => ({
          ...asset, id: assetId(`audit-item-${index}`), title: `Tool ${index}`, kind: 'item' as const, parentAssetId: asset.id
        })) : [] };
      } }),
      photos: new AssetPhotosQuery({ getAssetPhotos: async () => {
        if (++photoReads === 1 && mode === 'recovery') throw new Error('Audit photos unavailable');
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
