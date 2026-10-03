import { describe, expect, it } from 'vitest';
import { redirectSystemPath } from '../../app/+native-intent';

describe('native auth callback navigation ownership', () => {
  it('leaves warm callback handling to AuthSession and removes cold callback parameters', () => {
    const path = 'stuffstash://auth/callback?code=private-code&state=private-state';
    expect(redirectSystemPath({ path, initial: false })).toBe('');
    expect(redirectSystemPath({ path, initial: true })).toBe('/');
    expect(redirectSystemPath({ path: 'stuffstash://auth/callback?error=access_denied', initial: false })).toBe('');
  });
  it.each([
    'stuffstash://invitations/accept?token=invitation',
    'https://stash.example.test/invitations/accept?token=invitation',
    'https://auth/callback?code=private-code',
    'stuffstash://auth.evil/callback?code=private-code',
    'stuffstash://auth/callback/extra?code=private-code',
    'stuffstash://user@auth/callback?code=private-code',
    'stuffstash://auth:80/callback?code=private-code',
    '/assets/item', 'not a URL'
  ])('preserves unrelated or lookalike links: %s', path => {
    expect(redirectSystemPath({ path, initial: false })).toBe(path);
    expect(redirectSystemPath({ path, initial: true })).toBe(path);
  });
});
