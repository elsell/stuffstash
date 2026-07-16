import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../../../../ui/navigation/AppServicesContext';
import { AssetDetailRouteScreen } from '../../../../../ui/screens/AssetDetailRouteScreen';

export default function LocationAssetDetailRoute() {
  const {
    addAssetPhotosCommand,
    assetCheckoutCommand,
    assetDetailQuery,
    assetLifecycleCommand,
    deleteAssetPhotoCommand,
    photoSelectionQuery,
    undoAssetEditCommand
  } = useAppServices();
  const { assetId } = useLocalSearchParams<{ readonly assetId: string }>();

  return (
    <AssetDetailRouteScreen
      addAssetPhotosCommand={addAssetPhotosCommand}
      assetCheckoutCommand={assetCheckoutCommand}
      assetDetailQuery={assetDetailQuery}
      assetLifecycleCommand={assetLifecycleCommand}
      deleteAssetPhotoCommand={deleteAssetPhotoCommand}
      photoSelectionQuery={photoSelectionQuery}
      undoAssetEditCommand={undoAssetEditCommand}
      assetId={assetId}
    />
  );
}
