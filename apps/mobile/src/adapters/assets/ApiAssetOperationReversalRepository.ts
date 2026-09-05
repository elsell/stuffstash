import { ignoreInventoryMutations, type InventoryMutationObserver } from '../../application/home/InventoryMutationObserver';
import type { StuffStashClient } from '@stuff-stash/api-client';
import type { AssetOperationReversalRepository } from '../../application/assets/AssetOperationReversalRepository';

type UndoClient = Pick<StuffStashClient, 'applyUndoableOperation'>;

export class ApiAssetOperationReversalRepository implements AssetOperationReversalRepository {
  constructor(private readonly client: UndoClient, private readonly mutations: InventoryMutationObserver = ignoreInventoryMutations) {}

  async reverseAssetOperation(input: Parameters<AssetOperationReversalRepository['reverseAssetOperation']>[0]): Promise<void> {
    const asset = await this.client.applyUndoableOperation(input.tenantId, input.inventoryId, input.operationId, 'undo');
    this.mutations.onInventoryMutation({ kind: 'operation_reversed', tenantId: input.tenantId, inventoryId: input.inventoryId, assetId: asset.id, relatedAssetIds: asset.parentAssetId ? [asset.parentAssetId] : [] });
  }
}
