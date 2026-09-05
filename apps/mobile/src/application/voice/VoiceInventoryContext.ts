import type { InventoryId, TenantId } from '../../domain/inventories/InventorySummary';
import type { ReadRequest } from '../shared/ReadRequest';

export type VoiceInventoryContext = {
  readonly tenantId: TenantId;
  readonly inventoryId: InventoryId;
  readonly tenantName: string;
  readonly inventoryName: string;
};
export interface VoiceInventoryContextRepository {
  getVoiceInventoryContext(request?: ReadRequest): Promise<VoiceInventoryContext>;
}
export interface VoiceInventoryMutationObserver {
  onVoicePlanExecuted(impact: { readonly tenantId: string; readonly inventoryId: string; readonly assetIds: readonly string[] }): void;
}
