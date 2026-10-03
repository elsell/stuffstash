import type { createTranslator, en } from '@stuff-stash/localization';
export function expirationMonth(date: string, expired: boolean, translator: ReturnType<typeof createTranslator<typeof en>>): string {
  const month = translator.date(new Date(`${date.slice(0, 7)}-01T12:00:00Z`), { month: 'long', year: 'numeric', timeZone: 'UTC' });
  return expired ? translator.message('expiration.expiredMonth', { month }) : month;
}
