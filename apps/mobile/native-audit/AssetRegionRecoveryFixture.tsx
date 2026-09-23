import { useEffect, useState } from 'react';
import { Image } from 'react-native';
import { useLocalSearchParams } from 'expo-router';
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

type DetailVariant = 'default' | 'photo' | 'checked-out' | 'read-only' | 'place';
const detailTitles: Record<DetailVariant, string> = {
  default: 'Garage shelves and seasonal storage', photo: 'Camping tent',
  'checked-out': 'Camping gear', 'read-only': 'Spare camping gear', place: 'Garage'
};
export function AssetDetailCommandsFixture() {
  const { variant } = useLocalSearchParams<{ variant?: string }>();
  const selected: DetailVariant = variant === 'photo' || variant === 'checked-out'
    || variant === 'read-only' || variant === 'place' ? variant : 'default';
  return <AssetDetailFixture key={selected} mode="commands" variant={selected} />;
}

function AssetDetailFixture({ mode, variant = 'default' }: { readonly mode: 'recovery' | 'search' | 'commands'; readonly variant?: DetailVariant }) {
  const [fixture] = useState(() => {
    const client = createMobileQueryClient();
    const defaults = client.getDefaultOptions();
    client.setDefaultOptions({ ...defaults, queries: { ...defaults.queries, retry: false } });
    const asset = { id: assetId('audit-place'), title: mode === 'commands' ? detailTitles[variant] : 'Audit place',
      kind: variant === 'photo' ? 'item' as const : mode === 'commands' && variant !== 'place' ? 'container' as const : 'location' as const,
      lifecycleState: 'active' as const, description: '', locationLabel: '', locationTrail: [],
      parentLocationTrail: [], updatedAtLabel: '', hasPhoto: variant === 'photo',
      currentCheckout: variant === 'checked-out' || variant === 'read-only' ? {
        id: 'audit-checkout', state: 'checked_out', checkedOutAt: '2026-09-01T10:00:00Z',
        checkedOutByPrincipalId: 'Alex'
      } : undefined };
    let contentsReads = 0; let photoReads = 0;
    return { client,
      core: new AssetCoreQuery({ getAssetCore: async () => ({
        tenantId: tenantId('audit-tenant'), inventoryId: inventoryId('audit-inventory'),
        permissions: mode === 'commands' && variant !== 'read-only' ? ['view', 'edit_asset', 'create_asset'] : ['view'], revision: 'audit', asset
      }) }),
      contents: new AssetContentsQuery({ getAssetContents: async () => {
        if (++contentsReads === 1 && mode === 'recovery') throw new Error('Audit contents unavailable');
        return { asset, allAssets: mode === 'search' || variant === 'place' ? Array.from({ length: 20 }, (_, index) => ({
          ...asset, id: assetId(`audit-item-${index}`), title: `Tool ${index}`, kind: 'item' as const, parentAssetId: asset.id
        })) : [] };
      } }),
      photos: new AssetPhotosQuery({ getAssetPhotos: async () => {
        if (++photoReads === 1 && mode === 'recovery') throw new Error('Audit photos unavailable');
        return variant === 'photo' ? [{ id: 'audit-detail-photo', fileName: 'Fixture image.png',
          uri: Image.resolveAssetSource(require('../assets/brand/stuff-stash-glyph.png')).uri }] : [];
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
