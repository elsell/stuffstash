import { useInvitationRouteActions } from '../../ui/navigation/useInvitationRouteActions';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { useEffect } from 'react';
import { useAppConnectionActions, useAppServices } from '../../ui/navigation/AppServicesContext';
import { useInventoryInvitationLink } from '../../ui/navigation/InventoryInvitationLinkContext';
import { InventoryInvitationScreen } from '../../ui/screens/InventoryInvitationScreen';

export default function InventoryInvitationRoute() {
  const router = useRouter();
  const routeParams = useLocalSearchParams();
  const { signOut, changeServer } = useAppConnectionActions();
  const link = useInventoryInvitationLink();
  const {
    acceptInventoryInvitationCommand,
    previewInventoryInvitationQuery,
    selectInventoryCommand
  } = useAppServices();
  useEffect(() => {
    if (link.initialized && Object.keys(routeParams).length > 0) {
      router.replace('/invitations/accept');
    }
  }, [link.initialized, routeParams, router]);
  const actions = useInvitationRouteActions({
    reference: link.reference,
    clear: link.clear,
    goHome: () => router.replace('/'),
    selectInventory: id => selectInventoryCommand.execute(id),
    startOver: changeServer
  });
  return (
    <InventoryInvitationScreen
      acceptCommand={acceptInventoryInvitationCommand}
      initialized={link.initialized}
      invalidLink={link.invalid}
      onAccepted={actions.openInventory}
      onDismiss={actions.dismiss}
      onSwitchAccount={() => void signOut()}
      onStartOver={actions.startOver}
      previewQuery={previewInventoryInvitationQuery}
      reference={link.reference}
    />
  );
}
