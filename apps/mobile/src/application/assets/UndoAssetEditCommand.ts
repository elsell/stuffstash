import { t } from '../../presentation/localization';
import type { AssetOperationReversalRepository } from './AssetOperationReversalRepository';

export class UndoAssetEditCommand {
  constructor(private readonly repository: AssetOperationReversalRepository) {}

  async execute(input: { readonly tenantId: string; readonly inventoryId: string; readonly operationId: string }): Promise<void> {
    const tenantId = input.tenantId.trim();
    const inventoryId = input.inventoryId.trim();
    const operationId = input.operationId.trim();
    if (!tenantId || !inventoryId || !operationId) {
      throw new Error(t('recovery.undoUnavailable'));
    }
    await this.repository.reverseAssetOperation({ tenantId, inventoryId, operationId });
  }
}
