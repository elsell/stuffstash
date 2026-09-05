import { expect, it } from 'vitest';
import { createMobileQueryClient, mobileQueryKeys } from './MobileQueryClient';
import { QueryClientProviderProfileMutationObserver } from './QueryClientProviderProfileMutationObserver';

it('invalidates only the changed tenant provider and voice configuration', () => {
  const client = createMobileQueryClient();
  const keys = [mobileQueryKeys.providerProfiles('scope', 'home'), mobileQueryKeys.voiceConfiguration('scope', 'home'), mobileQueryKeys.providerProfiles('scope', 'other'), mobileQueryKeys.providerProfiles('other-session', 'home')];
  for (const key of keys) client.setQueryData(key, 'cached');
  new QueryClientProviderProfileMutationObserver(client, 'scope').onProviderProfilesChanged('home');
  expect(keys.map((key) => client.getQueryState(key)?.isInvalidated)).toEqual([true, true, false, false]);
  client.clear();
});
