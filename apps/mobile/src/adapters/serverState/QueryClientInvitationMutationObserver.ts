import type { QueryClient, InfiniteData } from '@tanstack/react-query';
import type { InventoryInvitationPage, InventoryInvitationMutationObserver, InventorySharingScope } from '../../application/sharing/InventorySharing';
import { mobileQueryKeys } from './MobileQueryClient';

export class QueryClientInvitationMutationObserver implements InventoryInvitationMutationObserver {
  constructor(private readonly client: QueryClient, private readonly scopeId: string) {}
  onInvitationsChanged(scope: InventorySharingScope, cancelledInvitationId?: string): void {
    const queryKey = mobileQueryKeys.invitations(this.scopeId, scope.tenantId, scope.inventoryId);
    if (cancelledInvitationId) {
      this.client.setQueryData<InfiniteData<InventoryInvitationPage>>(queryKey, current => current ? {
        ...current,
        pages: current.pages.map(page => ({ ...page, items: page.items.map(item => item.id === cancelledInvitationId ? { ...item, status: 'cancelled' as const } : item) }))
      } : undefined);
    }
    void this.client.invalidateQueries({ queryKey });
  }
}
