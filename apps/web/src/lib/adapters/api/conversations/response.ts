import { ConversationFailure } from '$lib/domain/conversation';
import type { ConversationPage } from '$lib/ports/conversationWorkflowRepository';

interface Meta {
  tenantId?: string;
  pagination?: { limit: number; nextCursor?: string | null; hasMore: boolean };
}
export async function conversationResponse<T extends { meta: Meta }>(
  pending: Promise<{ data?: T; response: Response }>, tenantId: string, signal?: AbortSignal
): Promise<T> {
  let result: { data?: T; response: Response };
  try { result = await pending; } catch (error) {
    if (signal?.aborted) throw signal.reason;
    if (error instanceof Error && error.name === 'AbortError') throw error;
    throw new ConversationFailure('unavailable');
  }
  const { data, response } = result;
  if (!response.ok) {
    const kind = response.status === 401 ? 'unauthenticated'
      : response.status === 403 ? 'forbidden'
      : response.status === 409 ? 'conflict'
      : response.status === 412 ? 'precondition'
      : response.status >= 400 && response.status < 500 ? 'invalid' : 'unavailable';
    throw new ConversationFailure(kind);
  }
  if (!data || !data.meta || (data.meta.tenantId && data.meta.tenantId !== tenantId)) throw new ConversationFailure('invalid');
  return data;
}
export function conversationPage<T, U>(envelope: { data: T[] | null; meta: Meta }, map: (value: T) => U): ConversationPage<U> {
  const pagination = envelope.meta.pagination;
  if (!pagination || !Number.isInteger(pagination.limit) || pagination.limit < 1 || pagination.limit > 100 ||
    (pagination.hasMore && !pagination.nextCursor)) throw new ConversationFailure('invalid');
  return { items: (envelope.data ?? []).map(map), pagination: {
    limit: pagination.limit, nextCursor: pagination.nextCursor || null, hasMore: pagination.hasMore
  } };
}
