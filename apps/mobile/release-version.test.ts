import { createRequire } from 'node:module';
import { describe, expect, it } from 'vitest';

const require = createRequire(import.meta.url);
const { resolveMobileReleaseVersion } = require('./scripts/release-version.js');

describe('mobile release version', () => {
  it('derives the marketing version from an exact repository release tag', () => {
    expect(resolveMobileReleaseVersion({ tag: 'v0.15.0', production: true })).toBe('0.15.0');
  });

  it.each(['1.2.3', 'v1.2', 'v1.2.3-beta.1', 'v01.2.3', 'v1.02.3', 'v1.2.03', ''])
  ('rejects a malformed production release tag: %s', (tag) => {
    expect(() => resolveMobileReleaseVersion({ tag, production: true })).toThrow(
      'STUFF_STASH_MOBILE_RELEASE_TAG'
    );
  });

  it('does not trust prepared metadata without the release tag', () => {
    expect(() => resolveMobileReleaseVersion({ preparedVersion: '2.4.6', production: true })).toThrow(
      'STUFF_STASH_MOBILE_RELEASE_TAG'
    );
  });

  it('uses a non-release placeholder for local development', () => {
    expect(resolveMobileReleaseVersion({ production: false })).toBe('0.0.0');
  });
});
