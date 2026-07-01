import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../../ui/navigation/AppServicesContext';
import { AssetEditSheetRouteScreen } from '../../../ui/screens/AssetNativeActionSheetScreens';

export default function AssetEditRoute() {
  const { assetDetailQuery, updateAssetCommand } = useAppServices();
  const { assetId } = useLocalSearchParams<{ readonly assetId: string }>();

  return (
    <AssetEditSheetRouteScreen
      assetDetailQuery={assetDetailQuery}
      assetId={assetId}
      updateAssetCommand={updateAssetCommand}
    />
  );
}
