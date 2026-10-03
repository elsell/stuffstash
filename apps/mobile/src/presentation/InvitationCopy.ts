import type { createTranslator, en } from '@stuff-stash/localization';
export function invitationCopy(translator: ReturnType<typeof createTranslator<typeof en>>) {
  return {
    expiration: (value: string) => translator.date(new Date(value), { dateStyle: 'medium', timeStyle: 'short' }),
    accepted: (inventoryName: string) => translator.message('invitation.acceptedInventory', { inventoryName }),
  };
}
