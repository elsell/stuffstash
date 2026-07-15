import { describe, expect, it } from 'vitest';
import { collectCursorPages } from './cursorPagination';

describe('collectCursorPages', () => {
  it('collects every page in order', async () => {
    const cursors: Array<string | undefined> = [];
    const items = await collectCursorPages(async (cursor) => {
      cursors.push(cursor);
      return cursor
        ? { items: ['third'], pagination: { limit: 2, nextCursor: null, hasMore: false } }
        : { items: ['first', 'second'], pagination: { limit: 2, nextCursor: 'page-two', hasMore: true } };
    });

    expect(items).toEqual(['first', 'second', 'third']);
    expect(cursors).toEqual([undefined, 'page-two']);
  });

  it('rejects missing and repeated continuation cursors', async () => {
    await expect(
      collectCursorPages(async () => ({ items: ['first'], pagination: { limit: 1, nextCursor: null, hasMore: true } }))
    ).rejects.toThrow('missing continuation cursor');

    await expect(
      collectCursorPages(async () => ({ items: ['first'], pagination: { limit: 1, nextCursor: 'same', hasMore: true } }))
    ).rejects.toThrow('repeated continuation cursor');
  });
});
