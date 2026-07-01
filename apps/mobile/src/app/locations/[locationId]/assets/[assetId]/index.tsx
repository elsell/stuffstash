import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../../../../ui/navigation/AppServicesContext';
import { AssetDetailRouteScreen } from '../../../../../ui/screens/AssetDetailRouteScreen';

export default function LocationAssetDetailRoute() {
  const {
    addAssetPhotosCommand,
    assetDetailQuery,
    assetLifecycleCommand,
    deleteAssetPhotoCommand,
    photoSelectionQuery
  } = useAppServices();
  const { assetId } = useLocalSearchParams<{ readonly assetId: string }>();

  return (
    <AssetDetailRouteScreen
      addAssetPhotosCommand={addAssetPhotosCommand}
      assetDetailQuery={assetDetailQuery}
      assetLifecycleCommand={assetLifecycleCommand}
      deleteAssetPhotoCommand={deleteAssetPhotoCommand}
      photoSelectionQuery={photoSelectionQuery}
      assetId={assetId}
    />
  );
}
