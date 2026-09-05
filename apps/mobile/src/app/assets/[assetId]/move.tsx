import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../../ui/navigation/AppServicesContext';
import { AssetMoveSheetRouteScreen } from '../../../ui/screens/AssetNativeActionSheetScreens';

export default function AssetMoveRoute() {
  const {
    assetCoreQuery,
    assetPlacementQuery,
    createAssetCommand,
    moveAssetCommand,
    parentLookupQuery
  } = useAppServices();
  const { assetId } = useLocalSearchParams<{ readonly assetId: string }>();

  return (
    <AssetMoveSheetRouteScreen
      assetCoreQuery={assetCoreQuery}
      assetPlacementQuery={assetPlacementQuery}
      assetId={assetId}
      createAssetCommand={createAssetCommand}
      moveAssetCommand={moveAssetCommand}
      parentLookupQuery={parentLookupQuery}
    />
  );
}
