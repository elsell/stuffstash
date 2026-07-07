import { describe, expect, it } from 'vitest';
import { assetTagKeyFromDisplayName } from './AssetSummary';

describe('asset tag keys', () => {
  it('normalizes display names like the backend tag key rule', () => {
    expect(assetTagKeyFromDisplayName(' Camp / Kitchen ')).toBe('camp-kitchen');
    expect(assetTagKeyFromDisplayName('Kids & Toys')).toBe('kids-toys');
    expect(assetTagKeyFromDisplayName('###')).toBe('');
  });

  it('truncates long keys and trims trailing separators', () => {
    expect(assetTagKeyFromDisplayName(`${'a'.repeat(79)} / camping`)).toBe('a'.repeat(79));
    expect(assetTagKeyFromDisplayName(`${'a'.repeat(80)}b`)).toBe('a'.repeat(80));
  });
});
