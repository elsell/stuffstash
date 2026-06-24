import { useAppServices } from '../../ui/navigation/AppServicesContext';
import { LocationsScreen } from '../../ui/screens/LocationsScreen';

export default function LocationsRoute() {
  const { locationsQuery } = useAppServices();

  return <LocationsScreen locationsQuery={locationsQuery} />;
}
