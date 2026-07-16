export interface CursorPage<T> {
  items: T[];
  pagination: {
    hasMore: boolean;
    nextCursor: string | null;
  };
}

export async function collectCursorPages<T>(loadPage: (cursor?: string) => Promise<CursorPage<T>>): Promise<T[]> {
  const items: T[] = [];
  const seenCursors = new Set<string>();
  let cursor: string | undefined;

  while (true) {
    const page = await loadPage(cursor);
    items.push(...page.items);
    if (!page.pagination.hasMore) {
      return items;
    }
    const nextCursor = page.pagination.nextCursor;
    if (!nextCursor) {
      throw new Error('Paginated response is missing continuation cursor.');
    }
    if (seenCursors.has(nextCursor)) {
      throw new Error('Paginated response repeated continuation cursor.');
    }
    seenCursors.add(nextCursor);
    cursor = nextCursor;
  }
}
