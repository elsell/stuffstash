import { t } from '../../presentation/localization';
import { Share } from 'react-native';
import type { InvitationLinkActions } from '../../application/sharing/InventorySharing';

export class ExpoInvitationLinkActions implements InvitationLinkActions {
  async copy(link: string): Promise<void> {
    const clipboard = await import('expo-clipboard');
    await clipboard.setStringAsync(link);
  }

  async share(input: { readonly link: string; readonly inventoryName: string }): Promise<void> {
    await Share.share({
      message: t('sharing.invitation.message', { inventory: input.inventoryName, link: input.link }),
      title: t('sharing.invitation.title')
    });
  }
}
