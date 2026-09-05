import type { QueryClient } from '@tanstack/react-query';
import type { VoiceInventoryMutationObserver } from '../../application/voice/VoiceInventoryContext';
import { mobileQueryKeys } from './MobileQueryClient';

export class QueryClientVoiceMutationObserver implements VoiceInventoryMutationObserver {
  constructor(private readonly client: QueryClient, private readonly scopeId: string) {}
  onVoicePlanExecuted(impact: Parameters<VoiceInventoryMutationObserver['onVoicePlanExecuted']>[0]): void {
    const prefix = mobileQueryKeys.inventory(this.scopeId, impact.tenantId, impact.inventoryId);
    const affected = new Set(impact.assetIds);
    void this.client.invalidateQueries({ predicate: ({ queryKey }) => {
      if (!prefix.every((value, index) => queryKey[index] === value)) return false;
      const resource = queryKey[prefix.length];
      if (resource !== 'asset') return ['home', 'map', 'assets', 'locations', 'location', 'browse', 'parent-candidates'].includes(String(resource));
      const region = queryKey[prefix.length + 2];
      if (region === 'photos') return false;
      // Executed results omit automatically promoted parent identities.
      if (region === 'core' || region === 'contents' || region === 'placement') return true;
      return affected.size === 0 || affected.has(String(queryKey[prefix.length + 1]));
    } });
  }
}
