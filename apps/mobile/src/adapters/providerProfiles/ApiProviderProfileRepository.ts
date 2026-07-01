import type {
  ProviderProfile,
  ProviderProfileSummary as ApiProviderProfileSummary,
  ProviderProfileTestResult,
  StuffStashClient,
  VoiceProviderConfiguration as ApiVoiceProviderConfiguration
} from '@stuff-stash/api-client';
import {
  CreateProviderProfileInput,
  ProviderCredentialPurpose,
  ProviderProfileLifecycleAction,
  ProviderProfileRepository,
  ProviderProfileSummary,
  ReplaceProviderProfileCredentialInput,
  UpdateProviderProfileInput,
  UpdateVoiceProviderConfigurationInput,
  VoiceProviderConfiguration
} from '../../application/providerProfiles/ProviderProfileRepository';

type ProviderProfileApiClient = Pick<
  StuffStashClient,
  | 'listProviderProfiles'
  | 'getVoiceProviderConfiguration'
  | 'updateVoiceProviderConfiguration'
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

  async getVoiceProviderConfiguration(): Promise<VoiceProviderConfiguration> {
    this.requireTenant();
    return mapVoiceProviderConfiguration(await this.client.getVoiceProviderConfiguration(this.tenantId));
  }

  async updateVoiceProviderConfiguration(
    input: UpdateVoiceProviderConfigurationInput
  ): Promise<VoiceProviderConfiguration> {
    this.requireTenant();
    return mapVoiceProviderConfiguration(
      await this.client.updateVoiceProviderConfiguration(this.tenantId, input)
    );
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
    const body = input.purpose === 'server_adc'
      ? { purpose: input.purpose }
      : { purpose: input.purpose, credential: input.credential ?? '' };

    return mapProviderProfile(
      await this.client.replaceProviderProfileCredential(this.tenantId, input.providerProfileId, body)
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

function mapApiProviderProfileSummary(profile: ApiProviderProfileSummary): ProviderProfileSummary {
  return {
    id: profile.id,
    capability: profile.capability,
    providerKind: profile.providerKind,
    displayName: profile.displayName,
    modelName: profile.modelName,
    credentialStatus: profile.credentialStatus,
    credentialPurpose: parseCredentialPurpose(profile.credentialPurpose) ?? credentialPurposeForSummary(profile),
    lifecycleState: profile.lifecycleState,
    lastTestedAt: profile.lastTestedAt,
    hasPromptTemplate: false
  };
}

function mapVoiceProviderConfiguration(
  configuration: ApiVoiceProviderConfiguration
): VoiceProviderConfiguration {
  return {
    tenantId: configuration.tenantId,
    readiness: configuration.readiness,
    updatedAt: configuration.updatedAt,
    profileIds: configuration.profileIds,
    slots: configuration.slots.map((slot) => ({
      capability: slot.capability,
      label: slot.label,
      selectedProfileId: slot.selectedProfileId,
      selectedProfile: slot.selectedProfile ? mapApiProviderProfileSummary(slot.selectedProfile) : undefined,
      selectionSource: slot.selectionSource,
      readiness: slot.readiness,
      issues: slot.issues,
      recommendedAction: slot.recommendedAction,
      duplicateProfiles: slot.duplicateProfiles.map(mapApiProviderProfileSummary)
    }))
  };
}

function credentialPurposeForProfile(profile: ProviderProfile): ProviderCredentialPurpose | undefined {
  const credentialType = profile.runtimeOptions.credentialType;
  if (credentialType === 'server_adc') {
    return 'server_adc';
  }
  if (credentialType === 'api_key') {
    return 'api_key';
  }
  if (profile.capability === 'text_to_speech') {
    return 'oauth_bearer';
  }

  return undefined;
}

function credentialPurposeForSummary(profile: ApiProviderProfileSummary): ProviderCredentialPurpose | undefined {
  if (profile.providerKind !== 'gemini') {
    return undefined;
  }
  return profile.capability === 'text_to_speech' ? 'oauth_bearer' : 'api_key';
}

function parseCredentialPurpose(value: string | undefined): ProviderCredentialPurpose | undefined {
  switch (value) {
    case 'api_key':
    case 'oauth_bearer':
    case 'server_adc':
      return value;
    default:
      return undefined;
  }
}
