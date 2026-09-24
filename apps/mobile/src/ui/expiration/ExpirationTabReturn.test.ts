import { expect, it } from 'vitest';
import { expirationOriginTab, expirationReturnPath } from './ExpirationTabReturn';

it('returns filters to their originating tab and safely defaults cold entry to Home', () => {
  expect(expirationReturnPath(expirationOriginTab(['(tabs)', '(search)', 'expiration'])))
    .toBe('/(tabs)/(search)/expiration');
  expect(expirationReturnPath(expirationOriginTab(['(tabs)', '(home)', 'expiration'])))
    .toBe('/(tabs)/(home)/expiration');
  for (const origin of [undefined, 'other', ['search'], '/settings']) {
    expect(expirationReturnPath(origin)).toBe('/(tabs)/(home)/expiration');
  }
});
