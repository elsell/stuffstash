export type ConversationRunState = 'queued' | 'running' | 'succeeded' | 'failed' | 'cancelled';
export type ConversationFailureKind = 'unauthenticated' | 'forbidden' | 'conflict' | 'precondition' | 'invalid' | 'unavailable';
export interface ConversationScope {
  apiIdentity: string;
  principalId: string;
  tenantId: string;
}
export class ConversationFailure extends Error {
  constructor(readonly kind: ConversationFailureKind) { super(kind); this.name = 'ConversationFailure'; }
}
