import type { InventoryAccessRelationship } from '$lib/domain/inventory';
import { inventoryAccessRelationships } from '$lib/domain/inventory';

export interface InventoryAccessRelationshipOption {
  value: InventoryAccessRelationship;
  label: string;
}

export type InventoryAccessListKind = 'grants' | 'invitations';
export type InventoryAccessListStatusKind = 'none' | 'loading' | 'empty';

export interface InventoryAccessListStatusPresentation {
  kind: InventoryAccessListStatusKind;
  message: string;
  role?: 'status';
}

const relationshipLabels: Record<InventoryAccessRelationship, string> = {
  viewer: 'Viewer',
  editor: 'Editor'
};

const listCopy: Record<InventoryAccessListKind, { loading: string; empty: string }> = {
  grants: {
    loading: 'Loading grants...',
    empty: 'No direct grants.'
  },
  invitations: {
    loading: 'Loading invitations...',
    empty: 'No invitations.'
  }
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

export function inventoryAccessListStatus(input: {
  kind: InventoryAccessListKind;
  busy: boolean;
  loaded: boolean;
  count: number;
}): InventoryAccessListStatusPresentation {
  if (input.busy && !input.loaded) {
    return { kind: 'loading', message: listCopy[input.kind].loading, role: 'status' };
  }
  if (input.loaded && input.count === 0) {
    return { kind: 'empty', message: listCopy[input.kind].empty };
  }
  return { kind: 'none', message: '' };
}
