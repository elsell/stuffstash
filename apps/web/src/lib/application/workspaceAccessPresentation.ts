import type { InventoryAccessRelationship } from '$lib/domain/inventory';
import { inventoryAccessRelationships } from '$lib/domain/inventory';

export interface InventoryAccessRelationshipOption {
  value: InventoryAccessRelationship;
  label: string;
}

const relationshipLabels: Record<InventoryAccessRelationship, string> = {
  viewer: 'Viewer',
  editor: 'Editor'
};

export function inventoryAccessRelationshipOptions(): InventoryAccessRelationshipOption[] {
  return inventoryAccessRelationships.map((relationship) => ({
    value: relationship,
    label: inventoryAccessRelationshipLabel(relationship)
  }));
}

export function inventoryAccessRelationshipLabel(relationship: InventoryAccessRelationship): string {
  return relationshipLabels[relationship];
}
