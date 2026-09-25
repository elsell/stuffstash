import { useAppServices } from '../ui/navigation/AppServicesContext';
import { TenantSwitcherSheetScreen } from '../ui/screens/TenantSwitcherSheetScreen';

export default function TenantSwitcherRoute() {
  const { homeDashboardQuery, selectInventoryCommand, createWorkspace } = useAppServices();

  return (
    <TenantSwitcherSheetScreen
      dashboardQuery={homeDashboardQuery}
      createWorkspace={createWorkspace}
      selectInventoryCommand={selectInventoryCommand}
    />
  );
}
