import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../../ui/navigation/AppServicesContext';
import { AssetEditSheetRouteScreen } from '../../../ui/screens/AssetNativeActionSheetScreens';

export default function AssetEditRoute() {
  const { assetCoreQuery, inventoryAssetTagsQuery, updateAssetCommand } = useAppServices();
  const { assetId } = useLocalSearchParams<{ readonly assetId: string }>();

  return (
    <AssetEditSheetRouteScreen
      assetCoreQuery={assetCoreQuery}
      assetId={assetId}
      inventoryAssetTagsQuery={inventoryAssetTagsQuery}
      updateAssetCommand={updateAssetCommand}
    />
  );
}
