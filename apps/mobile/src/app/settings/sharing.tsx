import { useAppServices } from '../../ui/navigation/AppServicesContext';
import { InventorySharingGuard } from '../../ui/navigation/InventorySharingGuard';
import { InventorySharingScreen } from '../../ui/screens/InventorySharingScreen';

export default function InventorySharingRoute() {
  const services = useAppServices();
  return (
    <InventorySharingGuard settingsQuery={services.settingsQuery}>
      {(scope) => (
        <InventorySharingScreen
          cancelCommand={services.cancelInventoryInvitationCommand}
          createCommand={services.createInventoryInvitationCommand}
          linkActions={services.invitationLinkActions}
          listQuery={services.listInventoryInvitationsQuery}
          scope={scope}
        />
      )}
    </InventorySharingGuard>
  );
}
