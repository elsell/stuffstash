import { useLocalSearchParams, useRouter } from 'expo-router';
import { useEffect } from 'react';
import { useAppConnectionActions, useAppServices } from '../../ui/navigation/AppServicesContext';
import { useInventoryInvitationLink } from '../../ui/navigation/InventoryInvitationLinkContext';
import { InventoryInvitationScreen } from '../../ui/screens/InventoryInvitationScreen';

export default function InventoryInvitationRoute() {
  const router = useRouter();
  const routeParams = useLocalSearchParams();
  const { signOut } = useAppConnectionActions();
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
  const dismiss = () => {
    link.clear();
    router.replace('/');
  };
  const openInventory = async (inventoryId: string) => {
    await selectInventoryCommand.execute(inventoryId);
    dismiss();
  };
  return (
    <InventoryInvitationScreen
      acceptCommand={acceptInventoryInvitationCommand}
      initialized={link.initialized}
      invalidLink={link.invalid}
      onAccepted={openInventory}
      onDismiss={dismiss}
      onSwitchAccount={() => void signOut()}
      previewQuery={previewInventoryInvitationQuery}
      reference={link.reference}
    />
  );
}
