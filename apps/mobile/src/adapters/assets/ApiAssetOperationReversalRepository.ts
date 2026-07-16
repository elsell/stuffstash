import type { StuffStashClient } from '@stuff-stash/api-client';
import type { AssetOperationReversalRepository } from '../../application/assets/AssetOperationReversalRepository';

type UndoClient = Pick<StuffStashClient, 'applyUndoableOperation'>;

export class ApiAssetOperationReversalRepository implements AssetOperationReversalRepository {
  constructor(private readonly client: UndoClient) {}

  async reverseAssetOperation(input: Parameters<AssetOperationReversalRepository['reverseAssetOperation']>[0]): Promise<void> {
    await this.client.applyUndoableOperation(input.tenantId, input.inventoryId, input.operationId, 'undo');
  }
}
