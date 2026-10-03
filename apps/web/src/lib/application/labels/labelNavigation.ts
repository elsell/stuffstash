import type {LabelDestination} from '$lib/domain/label';
import {workspaceRouteHref} from '../workspaceRoute';
export function labelDestinationHref(destination:LabelDestination):string {
  return workspaceRouteHref({mode:'asset',tenantId:destination.tenantId,inventoryId:destination.inventoryId,assetId:destination.assetId,lifecycleState:destination.archived?'archived':'active'},destination.tenantId,destination.inventoryId);
}
