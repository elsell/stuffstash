/** Bounds complete traversals and rejects a server cursor that would repeat work. */
export class ReadPageGuard {
  private readonly seen: Set<string>;
  private pages = 0;
  constructor(initialCursor?: string, private readonly maxPages = 1000) {
    this.seen = new Set(initialCursor ? [initialCursor] : []);
  }
  accept(value: string | null | undefined, hasMore?: boolean): string | undefined {
    const cursor = value ?? undefined;
    this.pages++;
    if ((hasMore && !cursor) || (cursor && this.seen.has(cursor))) throw new Error('Invalid read continuation cursor.');
    if (cursor && this.pages >= this.maxPages) throw new Error('Read exceeded the page limit.');
    if (cursor) this.seen.add(cursor);
    return cursor;
  }
}
