import type { AssetExpiration } from '$lib/domain/inventory';
export interface InventoryConversationScope { tenantId: string; inventoryId: string }
export interface ConversationAssetReference { id: string; title: string }
export interface InventoryConversationPlan {
  id: string;
  summary: string;
  commands: { summary: string; title?: string; destination?: string; changes?: string[]; expiration?: AssetExpiration; expirationCleared?: boolean }[];
  risks: string[];
}
export type InventoryConversationEvent =
  | { type: 'answer'; text: string; assets: ConversationAssetReference[] }
  | { type: 'review'; plan: InventoryConversationPlan }
  | { type: 'changed' }
  | { type: 'cancelled' }
  | { type: 'ready' }
  | { type: 'ended' };
export interface InventoryConversationConnection {
  send(text: string): void;
  decide(planId: string, approve: boolean): void;
  close(): void;
}
export interface InventoryConversationTransport {
  connect(scope: InventoryConversationScope, signal: AbortSignal,
    onEvent: (event: InventoryConversationEvent) => void,
    onFailure: (kind: 'authentication' | 'unavailable') => void): Promise<InventoryConversationConnection>;
}
