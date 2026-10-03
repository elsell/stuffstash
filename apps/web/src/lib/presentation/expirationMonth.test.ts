import { expect, it } from 'vitest';
import { createTranslator, en } from '@stuff-stash/localization';
import { expirationMonth } from './expirationMonth';
it('keeps expiration months anchored to UTC with localized expired headings', () => {
  const translator = createTranslator(en, { locale: 'de', translations: { de: { 'expiration.expiredMonth': '{month} abgelaufen' } } });
  expect(expirationMonth('2026-10-31', false, translator)).toBe('Oktober 2026');
  expect(expirationMonth('2026-10', true, translator)).toBe('Oktober 2026 abgelaufen');
  expect(expirationMonth('2026-11-01', false, translator)).toBe('November 2026');
});
