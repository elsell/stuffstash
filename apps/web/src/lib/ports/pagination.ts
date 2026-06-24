export interface Pagination {
  limit: number;
  nextCursor: string | null;
  hasMore: boolean;
}
