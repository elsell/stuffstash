import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../../ui/navigation/AppServicesContext';
import { AssetMoveSheetRouteScreen } from '../../../ui/screens/AssetNativeActionSheetScreens';

export default function AssetMoveRoute() {
  const {
    assetDetailQuery,
    createAssetCommand,
    moveAssetCommand,
    parentLookupQuery
  } = useAppServices();
  const { assetId } = useLocalSearchParams<{ readonly assetId: string }>();

  return (
    <AssetMoveSheetRouteScreen
      assetDetailQuery={assetDetailQuery}
      assetId={assetId}
      createAssetCommand={createAssetCommand}
      moveAssetCommand={moveAssetCommand}
      parentLookupQuery={parentLookupQuery}
    />
  );
}
