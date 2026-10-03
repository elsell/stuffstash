import { expect, it } from 'vitest';
import { createTranslator, en } from '@stuff-stash/localization';
import { timestampLabel } from './timestamp';

it('honors the configured locale without discarding time precision', () => {
  const value = '2026-10-03T12:34:56Z';
  const de = createTranslator(en, { locale: 'de' });
  const enUS = createTranslator(en, { locale: 'en-US' });
  expect(timestampLabel(value, de)).toBe(new Date(value).toLocaleString('de'));
  expect(timestampLabel(value, enUS)).toBe(new Date(value).toLocaleString('en-US'));
  expect(timestampLabel(value, de)).not.toBe(timestampLabel(value, enUS));
});
it('preserves the prior browser fallback for invalid values', () => {
  expect(timestampLabel('invalid', createTranslator(en, { locale: 'de' }))).toBe(new Date('invalid').toLocaleString('de'));
});
