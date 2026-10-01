import { t } from '../../presentation/localization';
import type { InvitationLinkActions } from '../../application/sharing/InventorySharing';

export type ClipboardGateway = {
  setStringAsync(value: string): Promise<unknown>;
};

export type NativeShareGateway = {
  share(content: { readonly message: string; readonly title: string }): Promise<unknown>;
};

export class NativeInvitationLinkActions implements InvitationLinkActions {
  constructor(
    private readonly clipboard: ClipboardGateway,
    private readonly nativeShare: NativeShareGateway
  ) {}

  async copy(link: string): Promise<void> {
    await this.clipboard.setStringAsync(link);
  }

  async share(input: { readonly link: string; readonly inventoryName: string }): Promise<void> {
    await this.nativeShare.share({
      message: t('sharing.invitation.message', { inventory: input.inventoryName, link: input.link }),
      title: t('sharing.invitation.title')
    });
  }
}
