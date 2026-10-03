import { expect, it } from 'vitest';
import { redirectSystemPath } from './+native-intent';
it('captures cold labels without exposing their params and leaves warm draft navigation untouched', () => {
  for (const path of ['stuffstash://labels/v1/instance/label', 'https://old.example/prefix/l/v1/instance/label']) {
    expect(redirectSystemPath({ path, initial: true })).toBe('/scan-label');
    expect(redirectSystemPath({ path, initial: false })).toBe('');
  }
  expect(redirectSystemPath({ path: 'stuffstash://auth/callback?code=secret', initial: false })).toBe('');
  expect(redirectSystemPath({ path: 'stuffstash://invitations/accept', initial: false })).toBe('stuffstash://invitations/accept');
});
