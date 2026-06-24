import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../../ui/navigation/AppServicesContext';
import { LocationAssetsRouteScreen } from '../../../ui/screens/LocationAssetsRouteScreen';

export default function LocationAssetsRoute() {
  const { locationAssetsQuery } = useAppServices();
  const { locationId } = useLocalSearchParams<{ readonly locationId: string }>();

  return (
    <LocationAssetsRouteScreen
      locationAssetsQuery={locationAssetsQuery}
      locationId={locationId}
    />
  );
}
