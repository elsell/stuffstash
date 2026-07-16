import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../../../ui/navigation/AppServicesContext';
import { AssetHistoryDetailRouteScreen } from '../../../../ui/screens/AssetHistoryDetailRouteScreen';

export default function AssetHistoryDetailRoute() {
  const { assetActivityQuery, revertAssetChangeCommand } = useAppServices();
  const params = useLocalSearchParams<{ readonly assetId: string; readonly activityId: string; readonly tenantId: string; readonly inventoryId: string; readonly assetTitle?: string }>();
  return (
    <AssetHistoryDetailRouteScreen
      assetActivityQuery={assetActivityQuery}
      revertAssetChangeCommand={revertAssetChangeCommand}
      activityId={params.activityId}
      assetId={params.assetId}
      tenantId={params.tenantId}
      inventoryId={params.inventoryId}
      assetTitle={params.assetTitle ?? 'Item'}
    />
  );
}
