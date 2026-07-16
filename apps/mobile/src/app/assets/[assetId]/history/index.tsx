import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../../../ui/navigation/AppServicesContext';
import { AssetHistoryRouteScreen } from '../../../../ui/screens/AssetHistoryRouteScreen';

export default function AssetHistoryRoute() {
  const { assetActivityQuery } = useAppServices();
  const params = useLocalSearchParams<{
    readonly assetId: string;
    readonly tenantId: string;
    readonly inventoryId: string;
    readonly assetTitle?: string;
  }>();
  return (
    <AssetHistoryRouteScreen
      assetActivityQuery={assetActivityQuery}
      tenantId={params.tenantId}
      inventoryId={params.inventoryId}
      assetId={params.assetId}
      assetTitle={params.assetTitle ?? 'Item'}
    />
  );
}
