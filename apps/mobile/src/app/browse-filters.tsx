import { useAppServices } from '../ui/navigation/AppServicesContext';
import { BrowseFiltersRouteScreen } from '../ui/screens/BrowseFiltersRouteScreen';

export default function BrowseFiltersRoute() {
  const services = useAppServices();
  return <BrowseFiltersRouteScreen inventoryAssetTagsQuery={services.inventoryAssetTagsQuery} />;
}
