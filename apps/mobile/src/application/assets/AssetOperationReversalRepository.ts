export interface AssetOperationReversalRepository {
  reverseAssetOperation(input: { readonly tenantId: string; readonly inventoryId: string; readonly operationId: string }): Promise<void>;
}
