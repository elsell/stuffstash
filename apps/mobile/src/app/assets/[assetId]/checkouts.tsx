import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../../ui/navigation/AppServicesContext';
import { AssetCheckoutHistorySheetRouteScreen } from '../../../ui/screens/AssetCheckoutHistoryScreen';

export default function AssetCheckoutHistoryRoute() {
  const { assetCheckoutHistoryQuery, assetCoreQuery } = useAppServices();
  const { assetId } = useLocalSearchParams<{ readonly assetId: string }>();

  return (
    <AssetCheckoutHistorySheetRouteScreen
      assetCheckoutHistoryQuery={assetCheckoutHistoryQuery}
      assetCoreQuery={assetCoreQuery}
      assetId={assetId}
    />
  );
}
