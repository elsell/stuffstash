import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../../ui/navigation/AppServicesContext';
import { AssetMoveHereSheetRouteScreen } from '../../../ui/screens/AssetNativeActionSheetScreens';

export default function AssetMoveHereRoute() {
  const {
    assetDetailQuery,
    moveAssetCommand,
    parentLookupQuery
  } = useAppServices();
  const { assetId } = useLocalSearchParams<{ readonly assetId: string }>();

  return (
    <AssetMoveHereSheetRouteScreen
      assetDetailQuery={assetDetailQuery}
      assetId={assetId}
      moveAssetCommand={moveAssetCommand}
      parentLookupQuery={parentLookupQuery}
    />
  );
}
