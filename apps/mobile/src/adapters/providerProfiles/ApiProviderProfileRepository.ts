import type { ProviderProfile, ProviderProfileTestResult, StuffStashClient } from '@stuff-stash/api-client';
import {
  CreateProviderProfileInput,
  ProviderProfileLifecycleAction,
  ProviderProfileRepository,
  ProviderProfileSummary,
  ReplaceProviderProfileCredentialInput,
  UpdateProviderProfileInput
} from '../../application/providerProfiles/ProviderProfileRepository';

type ProviderProfileApiClient = Pick<
  StuffStashClient,
  | 'listProviderProfiles'
  | 'createProviderProfile'
  | 'updateProviderProfile'
  | 'replaceProviderProfileCredential'
  | 'enableProviderProfile'
  | 'disableProviderProfile'
  | 'archiveProviderProfile'
  | 'testProviderProfile'
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

  async createProviderProfile(input: CreateProviderProfileInput): Promise<ProviderProfileSummary> {
    this.requireTenant();
    return mapProviderProfile(await this.client.createProviderProfile(this.tenantId, input));
  }

  async updateProviderProfile(input: UpdateProviderProfileInput): Promise<ProviderProfileSummary> {
    this.requireTenant();
    return mapProviderProfile(
      await this.client.updateProviderProfile(this.tenantId, input.providerProfileId, {
        promptTemplate: input.promptTemplate
      })
    );
  }

  async replaceProviderProfileCredential(
    input: ReplaceProviderProfileCredentialInput
  ): Promise<ProviderProfileSummary> {
    this.requireTenant();
    return mapProviderProfile(
      await this.client.replaceProviderProfileCredential(this.tenantId, input.providerProfileId, {
        purpose: input.purpose,
        credential: input.credential
      })
    );
  }

  async changeProviderProfileLifecycle(
    providerProfileId: string,
    action: ProviderProfileLifecycleAction
  ): Promise<ProviderProfileSummary> {
    this.requireTenant();
    switch (action) {
      case 'enable':
        return mapProviderProfile(await this.client.enableProviderProfile(this.tenantId, providerProfileId));
      case 'disable':
        return mapProviderProfile(await this.client.disableProviderProfile(this.tenantId, providerProfileId));
      case 'archive':
        return mapProviderProfile(await this.client.archiveProviderProfile(this.tenantId, providerProfileId));
    }
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
    credentialPurpose: credentialPurposeForProfile(profile),
    lifecycleState: profile.lifecycleState,
    lastTestedAt: profile.lastTestedAt,
    hasPromptTemplate: (profile.promptTemplate?.trim().length ?? 0) > 0
  };
}

function credentialPurposeForProfile(profile: ProviderProfile): 'api_key' | 'oauth_bearer' | undefined {
  if (profile.capability === 'text_to_speech') {
    return 'oauth_bearer';
  }

  return profile.runtimeOptions.credentialType === 'api_key' ? 'api_key' : undefined;
}
