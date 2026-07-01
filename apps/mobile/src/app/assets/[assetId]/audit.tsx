import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../../ui/navigation/AppServicesContext';
import { AssetAuditSheetRouteScreen } from '../../../ui/screens/AssetNativeActionSheetScreens';

export default function AssetAuditRoute() {
  const { assetAuditHistoryQuery, assetDetailQuery } = useAppServices();
  const { assetId } = useLocalSearchParams<{ readonly assetId: string }>();

  return (
    <AssetAuditSheetRouteScreen
      assetAuditHistoryQuery={assetAuditHistoryQuery}
      assetDetailQuery={assetDetailQuery}
      assetId={assetId}
    />
  );
}
