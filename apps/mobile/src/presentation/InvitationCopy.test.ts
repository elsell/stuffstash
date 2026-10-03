import { expect, it } from 'vitest';
import { createTranslator, en } from '@stuff-stash/localization';
import { invitationCopy } from './InvitationCopy';
it('formats invitation expiry with the selected locale and reorders the success message', () => {
  const translator = createTranslator(en, { locale: 'de', translations: { de: { 'invitation.acceptedInventory': '{inventoryName}: Zugriff gewährt.' } } });
  const copy = invitationCopy(translator);
  const expiry = '2026-10-01T12:00:00Z';
  expect(copy.expiration(expiry)).toBe(new Intl.DateTimeFormat('de', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(expiry)));
  expect(copy.accepted('Home {box}')).toBe('Home {box}: Zugriff gewährt.');
});
