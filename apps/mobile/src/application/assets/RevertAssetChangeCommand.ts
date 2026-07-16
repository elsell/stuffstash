import type { AssetOperationReversalRepository } from './AssetOperationReversalRepository';

export class RevertAssetChangeCommand {
  private readonly activeOperations = new Set<string>();

  constructor(private readonly repository: AssetOperationReversalRepository) {}

  async execute(input: { readonly tenantId: string; readonly inventoryId: string; readonly operationId: string }): Promise<boolean> {
    const tenantId = input.tenantId.trim();
    const inventoryId = input.inventoryId.trim();
    const operationId = input.operationId.trim();
    if (!tenantId || !inventoryId || !operationId) {
      throw new Error('This change can’t be reverted.');
    }
    const operationKey = `${tenantId}\u0000${inventoryId}\u0000${operationId}`;
    if (this.activeOperations.has(operationKey)) return false;
    this.activeOperations.add(operationKey);
    try {
      await this.repository.reverseAssetOperation({ tenantId, inventoryId, operationId });
      return true;
    } finally {
      this.activeOperations.delete(operationKey);
    }
  }
}
