import type { QueryClient } from '@tanstack/react-query';
import type { ProviderProfileMutationObserver } from '../../application/providerProfiles/ProviderProfileMutationObserver';
import { mobileQueryKeys } from './MobileQueryClient';

export class QueryClientProviderProfileMutationObserver implements ProviderProfileMutationObserver {
  constructor(private readonly client: QueryClient, private readonly scopeId: string) {}
  onProviderProfilesChanged(tenantId: string): void {
    void this.client.invalidateQueries({ queryKey: mobileQueryKeys.providerProfiles(this.scopeId, tenantId) });
    void this.client.invalidateQueries({ queryKey: mobileQueryKeys.voiceConfiguration(this.scopeId, tenantId) });
  }
}
