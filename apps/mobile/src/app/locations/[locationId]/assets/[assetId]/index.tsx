import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../../../../ui/navigation/AppServicesContext';
import { AssetDetailRouteScreen } from '../../../../../ui/screens/AssetDetailRouteScreen';

export default function LocationAssetDetailRoute() {
  const {
    addAssetPhotosCommand,
    assetAuditHistoryQuery,
    assetDetailQuery,
    assetLifecycleCommand,
    createAssetCommand,
    deleteAssetPhotoCommand,
    moveAssetCommand,
    parentLookupQuery,
    photoSelectionQuery,
    updateAssetCommand
  } = useAppServices();
  const { assetId } = useLocalSearchParams<{ readonly assetId: string }>();

  return (
    <AssetDetailRouteScreen
      addAssetPhotosCommand={addAssetPhotosCommand}
      assetAuditHistoryQuery={assetAuditHistoryQuery}
      assetDetailQuery={assetDetailQuery}
      assetLifecycleCommand={assetLifecycleCommand}
      createAssetCommand={createAssetCommand}
      deleteAssetPhotoCommand={deleteAssetPhotoCommand}
      moveAssetCommand={moveAssetCommand}
      parentLookupQuery={parentLookupQuery}
      photoSelectionQuery={photoSelectionQuery}
      updateAssetCommand={updateAssetCommand}
      assetId={assetId}
    />
  );
}
