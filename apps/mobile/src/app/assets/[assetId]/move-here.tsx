import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../../ui/navigation/AppServicesContext';
import { AssetMoveHereSheetRouteScreen } from '../../../ui/screens/AssetNativeActionSheetScreens';

export default function AssetMoveHereRoute() {
  const {
    assetCoreQuery,
    moveAssetCommand,
    parentLookupQuery
  } = useAppServices();
  const { assetId } = useLocalSearchParams<{ readonly assetId: string }>();

  return (
    <AssetMoveHereSheetRouteScreen
      assetCoreQuery={assetCoreQuery}
      assetId={assetId}
      moveAssetCommand={moveAssetCommand}
      parentLookupQuery={parentLookupQuery}
    />
  );
}
