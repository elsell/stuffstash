import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../../ui/navigation/AppServicesContext';
import { AssetCheckoutHistorySheetRouteScreen } from '../../../ui/screens/AssetNativeActionSheetScreens';

export default function AssetCheckoutHistoryRoute() {
  const { assetCheckoutHistoryQuery, assetDetailQuery } = useAppServices();
  const { assetId } = useLocalSearchParams<{ readonly assetId: string }>();

  return (
    <AssetCheckoutHistorySheetRouteScreen
      assetCheckoutHistoryQuery={assetCheckoutHistoryQuery}
      assetDetailQuery={assetDetailQuery}
      assetId={assetId}
    />
  );
}
