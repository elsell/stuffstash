import type { AssetKind } from '$lib/domain/inventory';
import type { WorkspaceRouteState } from './workspaceRoute';

export interface WorkspaceAddRouteInput {
  createAllowed: boolean;
  validParentIds: string[];
  selectedTenantId: string | null;
  selectedInventoryId: string | null;
}

export interface WorkspaceAddRouteResolution {
  open: boolean;
  kind: AssetKind;
  parentAssetId: string | null;
  deniedMessage: string;
  replacementRoute: Partial<WorkspaceRouteState> | null;
}

export function resolveWorkspaceAddRoute(
  route: WorkspaceRouteState,
  input: WorkspaceAddRouteInput
): WorkspaceAddRouteResolution {
  const kind = route.addKind ?? 'item';
  if (route.action !== 'add') {
    return closedResolution(kind);
  }
  if (!input.createAllowed) {
    return {
      ...closedResolution(kind),
      deniedMessage: 'You do not have permission to add assets in this inventory.'
    };
  }
  const parentAssetId = validAddParentId(route.addParentAssetId, input.validParentIds);
  return {
    open: true,
    kind,
    parentAssetId,
    deniedMessage: '',
    replacementRoute:
      route.addParentAssetId && !parentAssetId
        ? {
            action: 'add',
            addKind: kind,
            addParentAssetId: null,
            tenantId: input.selectedTenantId,
            inventoryId: input.selectedInventoryId
          }
        : null
  };
}

function closedResolution(kind: AssetKind): WorkspaceAddRouteResolution {
  return {
    open: false,
    kind,
    parentAssetId: null,
    deniedMessage: '',
    replacementRoute: null
  };
}

function validAddParentId(parentAssetId: string | null, validParentIds: string[]): string | null {
  if (!parentAssetId) {
    return null;
  }
  return validParentIds.includes(parentAssetId) ? parentAssetId : null;
}
