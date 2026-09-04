import type { QueryClient } from '@tanstack/react-query';
import type { InventorySelectionObserver } from '../../application/home/SelectInventoryCommand';
import { resetMobileInventorySelection } from './MobileQueryClient';

export class QueryClientInventorySelectionObserver implements InventorySelectionObserver {
  constructor(
    private readonly client: QueryClient,
    private readonly compositionScopeId: string
  ) {}

  onInventorySelected(): Promise<void> {
    return resetMobileInventorySelection(this.client, this.compositionScopeId);
  }
}
