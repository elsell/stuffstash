import { expect, it } from 'vitest';
import { createTranslator, en } from '@stuff-stash/localization';
import { sharingCopy } from './SharingCopy';

const invitation = { email: 'Pat@example.com', relationship: 'viewer' as const, status: 'pending' as const, isExpired: false, expiresAt: '2026-10-01T12:00:00Z' };
it('localizes roles and every lifecycle status, including expiration precedence', () => {
  const copy = sharingCopy(createTranslator(en, { locale: 'en-XA' }));
  for (const relationship of ['viewer', 'editor'] as const) {
    for (const status of ['pending', 'accepted', 'revoked', 'cancelled', 'expired'] as const) {
      const metadata = copy.metadata({ ...invitation, relationship, status });
      expect(metadata).not.toContain(status[0].toUpperCase() + status.slice(1));
      expect(metadata).not.toContain(relationship[0].toUpperCase() + relationship.slice(1));
    }
  }
  expect(copy.metadata({ ...invitation, isExpired: true })).toBe(copy.metadata({ ...invitation, status: 'expired' }));
});
it('allows translated messages to reorder fields without translating user content', () => {
  const translator = createTranslator(en, { locale: 'de', translations: { de: {
    'sharing.heading': '{inventoryName}: Freigabe',
    'sharing.createdMetadata': '{date} / {access} / {email}',
    'sharing.invitationMetadata': '{status} / {date} / {access}',
    'mobile.InventorySharingScreen.viewer': 'Leser',
    'sharing.status.pending': 'Ausstehend',
  } } });
  const copy = sharingCopy(translator);
  const date = new Intl.DateTimeFormat('de', { dateStyle: 'medium' }).format(new Date(invitation.expiresAt));
  expect(copy.heading('Home {name}')).toBe('Home {name}: Freigabe');
  expect(copy.createdMetadata(invitation)).toBe(`${date} / Leser / Pat@example.com`);
  expect(copy.metadata(invitation)).toBe(`Ausstehend / ${date} / Leser`);
  expect(copy.metadata({ ...invitation, expiresAt: 'invalid-date' })).toContain('invalid-date');
});
