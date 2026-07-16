import { Share } from 'react-native';
import type { InvitationLinkActions } from '../../application/sharing/InventorySharing';

export class ExpoInvitationLinkActions implements InvitationLinkActions {
  async copy(link: string): Promise<void> {
    const clipboard = await import('expo-clipboard');
    await clipboard.setStringAsync(link);
  }

  async share(input: { readonly link: string; readonly inventoryName: string }): Promise<void> {
    await Share.share({
      message: `You’re invited to ${input.inventoryName} in Stuff Stash.\n\n${input.link}`,
      title: 'Share Stuff Stash invitation'
    });
  }
}
