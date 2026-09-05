import type { AssetCoreQuery } from '../../application/assets/AssetCoreQuery';
import { assetContentsIdentity } from '../../application/assets/AssetCoreQuery';
import type { AssetContentsQuery } from '../../application/assets/AssetContentsQuery';
import type { AssetPhotosQuery } from '../../application/assets/AssetPhotosQuery';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { mergeProgressiveAssetDetail } from '../screens/AssetDetailProgressivePresentation';
import { useMobileInventoryServerQuery } from './useMobileInventoryServerQuery';

export type ProgressiveAssetDetailQueries = {
  readonly assetCoreQuery: Pick<AssetCoreQuery, 'execute'>;
  readonly assetContentsQuery: Pick<AssetContentsQuery, 'execute'>;
  readonly assetPhotosQuery: Pick<AssetPhotosQuery, 'execute'>;
};

export function useProgressiveAssetDetail(assetId: string | undefined, queries: ProgressiveAssetDetailQueries) {
  const coreAsset = useMobileInventoryServerQuery({
    key: (scope, tenant, inventory) => mobileQueryKeys.assetCore(scope, tenant, inventory, assetId ?? 'unselected'),
    query: (signal) => queries.assetCoreQuery.execute(assetId!, { signal }),
    enabled: Boolean(assetId)
  });
  const assetContents = useMobileInventoryServerQuery({
    key: (scope, tenant, inventory) => mobileQueryKeys.assetContents(scope, tenant, inventory, assetId ?? 'unselected',
      coreAsset.data ? assetContentsIdentity(coreAsset.data.snapshot) : 'pending'),
    query: (signal) => queries.assetContentsQuery.execute(coreAsset.data!.snapshot, { signal }),
    enabled: Boolean(assetId && coreAsset.data)
  });
  const assetPhotos = useMobileInventoryServerQuery({
    key: (scope, tenant, inventory) => mobileQueryKeys.assetPhotos(scope, tenant, inventory, assetId ?? 'unselected'),
    query: (signal) => queries.assetPhotosQuery.execute(coreAsset.data!.snapshot, { signal }),
    enabled: Boolean(assetId && coreAsset.data)
  });
  async function refresh(): Promise<void> {
    const priorIdentity = coreAsset.data ? assetContentsIdentity(coreAsset.data.snapshot) : undefined;
    const nextCore = await coreAsset.refetch({ throwOnError: true, cancelRefetch: false });
    if (!nextCore.data) throw new Error('Asset could not be loaded.');
    await Promise.all([
      assetPhotos.refetch({ throwOnError: true, cancelRefetch: false }),
      ...(assetContentsIdentity(nextCore.data.snapshot) === priorIdentity
        ? [assetContents.refetch({ throwOnError: true, cancelRefetch: false })]
        : [])
    ]);
  }
  return {
    coreAsset, assetContents, assetPhotos, refresh,
    asset: coreAsset.data ? mergeProgressiveAssetDetail(coreAsset.data.view, assetContents.data, assetPhotos.data) : undefined
  };
}
