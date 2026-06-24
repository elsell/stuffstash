import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../../../../ui/navigation/AppServicesContext';
import { AssetDetailRouteScreen } from '../../../../../ui/screens/AssetDetailRouteScreen';

export default function LocationAssetDetailRoute() {
  const { assetDetailQuery } = useAppServices();
  const { assetId } = useLocalSearchParams<{ readonly assetId: string }>();

  return <AssetDetailRouteScreen assetDetailQuery={assetDetailQuery} assetId={assetId} />;
}
