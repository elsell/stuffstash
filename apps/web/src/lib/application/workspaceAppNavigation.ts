import type { WorkspaceContext, WorkspaceMode } from '$lib/domain/inventory';
import { workspaceRouteHref, type WorkspaceRouteState } from './workspaceRoute';

type WorkspaceRouteDraft = Partial<WorkspaceRouteState>;

export function workspaceHomeRoute(context: WorkspaceContext): WorkspaceRouteDraft {
  return {
    mode: 'home',
    tenantId: context.selectedTenantId,
    inventoryId: context.selectedInventoryId,
    lifecycleState: context.assetLifecycleState
  };
}

export function workspaceHomeHref(context: WorkspaceContext): string {
  return workspaceRouteHref(workspaceHomeRoute(context), context.selectedTenantId || null, context.selectedInventoryId || null);
}

export interface WorkspaceAddCloseContext {
  mode: WorkspaceMode;
  selectedLocationId: string | null;
  selectedAssetId: string | null;
}

export function workspaceAddCloseRoute(context: WorkspaceContext, closeContext: WorkspaceAddCloseContext): WorkspaceRouteDraft {
  if (closeContext.selectedLocationId) {
    return {
      mode: 'location',
      tenantId: context.selectedTenantId,
      inventoryId: context.selectedInventoryId,
      locationId: closeContext.selectedLocationId
    };
  }

  if (closeContext.selectedAssetId) {
    return {
      mode: 'asset',
      tenantId: context.selectedTenantId,
      inventoryId: context.selectedInventoryId,
      assetId: closeContext.selectedAssetId
    };
  }

  return {
    mode: closeContext.mode,
    tenantId: context.selectedTenantId,
    inventoryId: context.selectedInventoryId,
    lifecycleState: context.assetLifecycleState
  };
}

export function workspaceAddCloseHref(context: WorkspaceContext, closeContext: WorkspaceAddCloseContext): string {
  return workspaceRouteHref(workspaceAddCloseRoute(context, closeContext), context.selectedTenantId || null, context.selectedInventoryId || null);
}

export function settingsOverviewRoute(context: WorkspaceContext): WorkspaceRouteDraft {
  return {
    mode: 'settings',
    tenantId: context.selectedTenantId,
    inventoryId: context.selectedInventoryId,
    settingsSection: 'overview'
  };
}

export function settingsOverviewHref(context: WorkspaceContext): string {
  return workspaceRouteHref(settingsOverviewRoute(context), context.selectedTenantId || null, context.selectedInventoryId || null);
}

export function inventoryHomeNormalizationRoute(context: WorkspaceContext, route: WorkspaceRouteState): WorkspaceRouteDraft {
  return {
    mode: 'home',
    tenantId: context.selectedTenantId,
    inventoryId: context.selectedInventoryId,
    lifecycleState: route.lifecycleState
  };
}

export function inventoryHomeNormalizationHref(context: WorkspaceContext, route: WorkspaceRouteState): string {
  return workspaceRouteHref(inventoryHomeNormalizationRoute(context, route), context.selectedTenantId || null, context.selectedInventoryId || null);
}

export function assetDetailBackRoute(context: WorkspaceContext, selectedLocationId: string | null): WorkspaceRouteDraft {
  if (selectedLocationId) {
    return {
      mode: 'location',
      tenantId: context.selectedTenantId,
      inventoryId: context.selectedInventoryId,
      locationId: selectedLocationId
    };
  }

  return workspaceHomeRoute(context);
}

export function assetDetailBackHref(context: WorkspaceContext, selectedLocationId: string | null): string {
  return workspaceRouteHref(assetDetailBackRoute(context, selectedLocationId), context.selectedTenantId || null, context.selectedInventoryId || null);
}
