import { readFileSync } from 'node:fs';
import { expect, it } from 'vitest';
import { createTranslator, en } from '@stuff-stash/localization';

it('uses the production Tags title for native settings header acceptance', () => {
  const production = readFileSync(new URL('../src/ui/navigation/PrimaryTabStack.tsx', import.meta.url), 'utf8');
  const fixture = readFileSync(new URL('./FixtureApplication.tsx', import.meta.url), 'utf8');
  const title = production.match(/name="settings\/inventory\/tags\/index" options=\{\{ title: t\('([^']+)'\)/);
  expect(title).not.toBeNull();
  const fixtureTitle = fixture.match(/name="audit-customization" options=\{\{ title: '([^']+)'/);
  const messageKey = title?.[1] as keyof typeof en;
  expect(fixtureTitle?.[1]).toBe(createTranslator(en, { locale: 'en' }).message(messageKey));
});
