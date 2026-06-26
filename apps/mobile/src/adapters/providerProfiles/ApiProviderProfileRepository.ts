import type { ProviderProfile, ProviderProfileTestResult, StuffStashClient } from '@stuff-stash/api-client';
import {
  ProviderProfileRepository,
  ProviderProfileSummary
} from '../../application/providerProfiles/ProviderProfileRepository';

type ProviderProfileApiClient = Pick<
  StuffStashClient,
  'listProviderProfiles' | 'testProviderProfile'
>;

export class ApiProviderProfileRepository implements ProviderProfileRepository {
  constructor(
    private readonly client: ProviderProfileApiClient,
    private readonly tenantId: string
  ) {}

  async listProviderProfiles(): Promise<readonly ProviderProfileSummary[]> {
    this.requireTenant();
    const profiles = await this.client.listProviderProfiles(this.tenantId);
    return profiles.map(mapProviderProfile);
  }

  async testProviderProfile(providerProfileId: string): Promise<ProviderProfileTestResult> {
    this.requireTenant();
    return this.client.testProviderProfile(this.tenantId, providerProfileId);
  }

  private requireTenant(): void {
    if (this.tenantId.trim().length === 0) {
      throw new Error('Complete mobile onboarding before managing voice provider profiles.');
    }
  }
}

function mapProviderProfile(profile: ProviderProfile): ProviderProfileSummary {
  return {
    id: profile.id,
    capability: profile.capability,
    providerKind: profile.providerKind,
    displayName: profile.displayName,
    modelName: profile.modelName,
    credentialStatus: profile.credentialStatus,
    lifecycleState: profile.lifecycleState,
    lastTestedAt: profile.lastTestedAt,
    hasPromptTemplate: (profile.promptTemplate?.trim().length ?? 0) > 0
  };
}
