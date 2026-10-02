import { expect, it } from 'vitest';
import { createTranslator, en, formatArchiveReview } from '@stuff-stash/localization';

const counts = { assets: 1, tags: 1, photos: 1, otherFiles: 1, customAssetTypes: 1, customFields: 1, omittedAttachments: 1, keyRemappings: [{}] };
it('renders singular restore content, schema and notices on both clients', () => {
  const copy = formatArchiveReview(createTranslator(en, { locale: 'en' }).message, counts);
  expect(copy).toEqual({ content: '1 item · 1 tag · 1 photo · 1 file', schema: '1 custom type · 1 custom field', omitted: '1 attachment omitted', remappings: '1 conflicting definition key will be renamed.' });
});
it('renders zero and mixed counts through plural messages', () => {
  const copy = formatArchiveReview(createTranslator(en, { locale: 'en' }).message, { ...counts, assets: 0, tags: 2, photos: 1, otherFiles: 12, customAssetTypes: 0, customFields: 2, omittedAttachments: 0, keyRemappings: [{}, {}] });
  expect(copy).toEqual({ content: '0 items · 2 tags · 1 photo · 12 files', schema: '0 custom types · 2 custom fields', omitted: '0 attachments omitted', remappings: '2 conflicting definition keys will be renamed.' });
});
it('expands every archive summary without changing counts', () => {
  const copy = formatArchiveReview(createTranslator(en, { locale: 'en-XA' }).message, counts);
  for (const value of Object.values(copy)) { expect(value).toMatch(/^\[/); expect(value).toContain('1'); expect(value).not.toMatch(/\{\w+\}/); }
});
