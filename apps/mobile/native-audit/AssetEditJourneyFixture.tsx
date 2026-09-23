import { QueryClientInventoryMutationObserver } from '../src/adapters/serverState/QueryClientInventoryMutationObserver';
import { createContext, useContext, useEffect, useState, type PropsWithChildren } from 'react';
import { useLocalSearchParams } from 'expo-router';
import { AssetDetailRouteScreen } from '../src/ui/screens/AssetDetailRouteScreen';
import { AssetEditSheetRouteScreen } from '../src/ui/screens/AssetNativeActionSheetScreens';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { AssetCoreQuery } from '../src/application/assets/AssetCoreQuery';
import { AssetContentsQuery } from '../src/application/assets/AssetContentsQuery';
import { AssetPhotosQuery } from '../src/application/assets/AssetPhotosQuery';
import { UpdateAssetCommand } from '../src/application/assets/UpdateAssetCommand';
import { PhotoSelectionQuery } from '../src/application/add/PhotoSelectionQuery';
import { assetId, assetTagKeyFromDisplayName, type AssetSummary, type AssetTagSummary } from '../src/domain/assets/AssetSummary';
import { tenantId, inventoryId } from '../src/domain/inventories/InventorySummary';

const unsupported = async () => { throw new Error('Outside the isolated Edit journey'); };
const scope = { tenantId: 'audit-tenant', inventoryId: 'audit-inventory' };

/** In-memory port implementation shared by the actual detail and editor routes. */
export function createAssetEditJourney() {
  let asset: AssetSummary = { id: assetId('audit-edit-item'), title: 'Camping tent', description: '', kind: 'item',
    lifecycleState: 'active', locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false, tags: [] };
  const tags: AssetTagSummary[] = [];
  let writes = 0;
  const client = createMobileQueryClient();
  const mutations = new QueryClientInventoryMutationObserver(client, 'edit-journey');
  return {
    client,
    core: new AssetCoreQuery({ getAssetCore: async id => {
      if (id !== asset.id) throw new Error('Unknown journey asset');
      return { tenantId: tenantId(scope.tenantId), inventoryId: inventoryId(scope.inventoryId),
        permissions: ['view', 'edit_asset'], revision: String(writes), asset };
    } }),
    contents: new AssetContentsQuery({ getAssetContents: async () => ({ asset, allAssets: [] }) }),
    photos: new AssetPhotosQuery({ getAssetPhotos: async () => [] }),
    selection: new PhotoSelectionQuery({ selectFromLibrary: unsupported, captureFromCamera: unsupported }),
    tags: { execute: async () => tags.map(tag => ({ ...tag, label: tag.displayName })) },
    update: new UpdateAssetCommand({
      createAssetTag: async input => {
        const tag = { ...input, id: `journey-tag-${tags.length + 1}`, key: assetTagKeyFromDisplayName(input.displayName) };
        tags.push(tag);
        mutations.onInventoryMutation({ ...scope, kind: 'asset_tag_created' });
        return { id: tag.id };
      },
      updateAsset: async input => {
        if (input.assetId !== asset.id) throw new Error('Unknown journey asset');
        asset = { ...asset, title: input.title ?? asset.title, description: input.description ?? asset.description,
          expiration: input.expiration === undefined ? asset.expiration : input.expiration ?? undefined,
          tags: input.tagIds === undefined ? asset.tags : tags.filter(tag => input.tagIds!.includes(tag.id)) };
        writes++;
        mutations.onInventoryMutation({ ...scope, kind: 'asset_updated', assetId: asset.id });
        return asset;
      }
    }),
    writeCount: () => writes
  };
}

const JourneyContext = createContext<ReturnType<typeof createAssetEditJourney> | undefined>(undefined);
export function AssetEditJourneyProvider({ children }: PropsWithChildren) {
  const [journey] = useState(createAssetEditJourney);
  useEffect(() => () => journey.client.clear(), [journey]);
  return <JourneyContext.Provider value={journey}>{children}</JourneyContext.Provider>;
}
function useJourney() {
  const journey = useContext(JourneyContext);
  if (!journey) throw new Error('Edit journey requires its fixture provider');
  return journey;
}
function JourneyServerState({ children }: PropsWithChildren) {
  const journey = useJourney();
  return <MobileServerStateProvider client={journey.client} scopeId="edit-journey" loadInventoryScope={async () => scope}>{children}</MobileServerStateProvider>;
}
export function AssetEditJourneyDetailFixture() {
  const journey = useJourney();
  return <JourneyServerState><AssetDetailRouteScreen assetId="audit-edit-item" assetCoreQuery={journey.core}
    assetContentsQuery={journey.contents} assetPhotosQuery={journey.photos} photoSelectionQuery={journey.selection}
    assetCheckoutCommand={{ execute: unsupported }} assetLifecycleCommand={{ execute: unsupported }}
    undoAssetEditCommand={{ execute: unsupported }} deleteAssetPhotoCommand={{ execute: unsupported }}
    addAssetPhotosCommand={{ execute: unsupported }} /></JourneyServerState>;
}
export function AssetEditJourneyEditorFixture() {
  const journey = useJourney();
  const params = useLocalSearchParams<{ assetId: string }>();
  return <JourneyServerState><AssetEditSheetRouteScreen assetId={params.assetId} assetCoreQuery={journey.core}
    inventoryAssetTypesQuery={{ execute: async () => [] }} inventoryAssetTagsQuery={journey.tags}
    updateAssetCommand={journey.update} /></JourneyServerState>;
}
