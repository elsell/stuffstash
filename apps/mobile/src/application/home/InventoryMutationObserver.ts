export type InventoryMutationKind =
  | 'asset_created'
  | 'asset_updated'
  | 'asset_lifecycle_changed'
  | 'asset_photo_changed'
  | 'asset_checkout_changed'
  | 'asset_tag_created'
  | 'operation_reversed';

export type InventoryMutation = {
  readonly kind: InventoryMutationKind;
  readonly tenantId: string;
  readonly inventoryId: string;
  readonly assetId?: string;
};

export interface InventoryMutationObserver {
  onInventoryMutation(mutation: InventoryMutation): void;
}

export const ignoreInventoryMutations: InventoryMutationObserver = {
  onInventoryMutation: () => undefined
};
