import { readFileSync } from 'node:fs';
import { expect, it } from 'vitest';

it('uses the production Tags title for native settings header acceptance', () => {
  const production = readFileSync(new URL('../src/app/_layout.tsx', import.meta.url), 'utf8');
  const fixture = readFileSync(new URL('./FixtureApplication.tsx', import.meta.url), 'utf8');
  const title = production.match(/name="settings\/inventory\/tags\/index" options=\{\{ title: '([^']+)'/);
  expect(title).not.toBeNull();
  const fixtureTitle = fixture.match(/name="audit-customization" options=\{\{ title: '([^']+)'/);
  expect(fixtureTitle?.[1]).toBe(title?.[1]);
});
