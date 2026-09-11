import { expect, it } from 'vitest';
import { isAssetExpiration } from './AssetExpiration';
it('validates calendar dates without timezone conversion', () => {
  expect(isAssetExpiration({ date: '2028-02-29', precision: 'day' })).toBe(true);
  expect(isAssetExpiration({ date: '2028-02', precision: 'month' })).toBe(true);
  for (const value of [null, {}, { date: '2027-02-29', precision: 'day' }, { date: '2028-02-30', precision: 'day' }, { date: '0000-01', precision: 'month' }, { date: '2028-13', precision: 'month' }, { date: '2028-02', precision: 'day' }]) expect(isAssetExpiration(value)).toBe(false);
});
