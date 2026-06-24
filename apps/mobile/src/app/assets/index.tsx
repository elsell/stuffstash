import { useAppServices } from '../../ui/navigation/AppServicesContext';
import { InventoryAssetsRouteScreen } from '../../ui/screens/InventoryAssetsRouteScreen';

export default function InventoryAssetsRoute() {
  const { inventoryAssetsQuery } = useAppServices();

  return <InventoryAssetsRouteScreen inventoryAssetsQuery={inventoryAssetsQuery} />;
}
