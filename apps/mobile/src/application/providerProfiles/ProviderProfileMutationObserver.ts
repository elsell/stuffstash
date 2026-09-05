export interface ProviderProfileMutationObserver {
  onProviderProfilesChanged(tenantId: string): void;
}
export const ignoreProviderProfileMutations: ProviderProfileMutationObserver = { onProviderProfilesChanged: () => undefined };
