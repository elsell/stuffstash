import { useAppServices } from '../ui/navigation/AppServicesContext';
import { TenantSwitcherSheetScreen } from '../ui/screens/TenantSwitcherSheetScreen';

export default function TenantSwitcherRoute() {
  const { homeDashboardQuery, selectInventoryCommand } = useAppServices();

  return (
    <TenantSwitcherSheetScreen
      dashboardQuery={homeDashboardQuery}
      selectInventoryCommand={selectInventoryCommand}
    />
  );
}
