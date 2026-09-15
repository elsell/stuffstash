import { useEffect, useState } from 'react';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { ManageProviderProfileCommand } from '../src/application/providerProfiles/ManageProviderProfileCommand';
import { ProviderProfileSettingsQuery } from '../src/application/providerProfiles/ProviderProfileSettingsQuery';
import type { ProviderProfileRepository, ProviderProfileSummary } from '../src/application/providerProfiles/ProviderProfileRepository';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { ProviderCredentialScreen, ProviderPromptScreen } from '../src/ui/screens/ProviderProfileEditorScreens';

// Runner-only ports. Never sends credentials or requests to a real provider.
export function ProviderEditorFixture() {
  const { kind } = useLocalSearchParams<{ kind: string }>();
  const router = useRouter();
  const [fixture] = useState(() => {
    const client = createMobileQueryClient();
    const profile: ProviderProfileSummary = {
      id: 'audit-provider', capability: 'language_inference', providerKind: 'gemini',
      displayName: 'Audit provider', modelName: 'audit-model', credentialStatus: 'missing',
      credentialPurpose: 'api_key', lifecycleState: 'disabled', hasPromptTemplate: false
    };
    let attempts = 0;
    async function replace() {
      if (++attempts === 1) throw new Error('Audit replacement unavailable. Try again.');
      return profile;
    }
    const repository: ProviderProfileRepository = {
      listProviderProfiles: async () => [profile],
      getVoiceProviderConfiguration: async () => ({ tenantId: 'audit-tenant', readiness: 'needs_attention', profileIds: {}, slots: [] }),
      replaceProviderProfileCredential: replace,
      updateProviderProfile: replace,
      createProviderProfile: async () => { throw new Error('Creation is outside this fixture.'); },
      changeProviderProfileLifecycle: async () => { throw new Error('Lifecycle is outside this fixture.'); },
      testProviderProfile: async () => { throw new Error('Provider testing is outside this fixture.'); },
      updateVoiceProviderConfiguration: async () => { throw new Error('Selection is outside this fixture.'); }
    };
    return { client, command: new ManageProviderProfileCommand(repository), query: new ProviderProfileSettingsQuery(repository) };
  });
  useEffect(() => () => fixture.client.clear(), [fixture]);
  const props = { manageCommand: fixture.command, query: fixture.query, profileId: 'audit-provider', onSaved: () => router.back() };
  return <MobileServerStateProvider client={fixture.client} scopeId="audit-provider-editor" loadInventoryScope={async () => ({ tenantId: 'audit-tenant', inventoryId: 'audit-inventory', inventoryName: 'Audit inventory', permissions: ['configure'] })}>
    {kind === 'prompt' ? <ProviderPromptScreen {...props} /> : <ProviderCredentialScreen {...props} />}
  </MobileServerStateProvider>;
}
