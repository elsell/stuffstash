import { t } from '$lib/presentation/localization';
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
      disabledReason: t('web.workspaceAddAvailability.selectAnInventoryBeforeAddingAssets')
    };
  }
  if (!input.canCreateAsset) {
    return {
      canOpen: false,
      disabledReason: t('web.workspaceAddAvailability.addingAssetsIsUnavailableForThisInventory')
    };
  }
  return {
    canOpen: true,
    disabledReason: ''
  };
}
