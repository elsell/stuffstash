import { describe, expect, it } from 'vitest';
import { compareNaturalText } from './textCollation';

describe('natural text collation', () => {
  it('orders numeric suffixes naturally and ignores case for grouping', () => {
    const values = ['Bin 10', 'bin 2', 'Attic', 'bin 1'];

    expect([...values].sort(compareNaturalText)).toEqual(['Attic', 'bin 1', 'bin 2', 'Bin 10']);
  });
});
