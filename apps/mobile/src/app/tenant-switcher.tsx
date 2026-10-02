import { useRouter } from 'expo-router';
import { useAppServices } from '../ui/navigation/AppServicesContext';
import { TenantSwitcherSheetScreen } from '../ui/screens/TenantSwitcherSheetScreen';

export default function TenantSwitcherRoute() {
  const router = useRouter();
  const { homeDashboardQuery, selectInventoryCommand, createWorkspace } = useAppServices();

  return (
    <TenantSwitcherSheetScreen
      onRestore={tenantId => router.push({ pathname: "/inventory-archive", params: { tenantId } })}
      dashboardQuery={homeDashboardQuery}
      createWorkspace={createWorkspace}
      selectInventoryCommand={selectInventoryCommand}
    />
  );
}
