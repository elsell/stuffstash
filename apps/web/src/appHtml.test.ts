import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { describe, expect, it } from 'vitest';

describe('document shell security policy', () => {
  it('declares no-referrer before application-managed head content', () => {
    const appHtml = readFileSync(resolve(process.cwd(), 'src/app.html'), 'utf8');
    const policy = '<meta name="referrer" content="no-referrer" />';

    expect(appHtml.match(new RegExp(policy, 'g'))).toHaveLength(1);
    expect(appHtml.indexOf(policy)).toBeLessThan(appHtml.indexOf('%sveltekit.head%'));
  });
});
