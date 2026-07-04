export interface WorkspaceAddAvailabilityInput {
  hasInventory: boolean;
  canCreateAsset: boolean;
}

export interface WorkspaceAddAvailability {
  canOpen: boolean;
  disabledReason: string;
}

export function workspaceAddAvailability(input: WorkspaceAddAvailabilityInput): WorkspaceAddAvailability {
  if (!input.hasInventory) {
    return {
      canOpen: false,
      disabledReason: 'Select an inventory before adding assets.'
    };
  }
  if (!input.canCreateAsset) {
    return {
      canOpen: false,
      disabledReason: 'Adding assets is unavailable for this inventory.'
    };
  }
  return {
    canOpen: true,
    disabledReason: ''
  };
}
