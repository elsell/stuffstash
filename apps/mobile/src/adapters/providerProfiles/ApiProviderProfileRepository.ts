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
  VoiceProviderConfiguration,
  VoiceProviderRecommendedAction
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

export interface ProviderProfileTenantScope {
  getCurrentTenantId(): Promise<string>;
}

export class ApiProviderProfileRepository implements ProviderProfileRepository {
  constructor(
    private readonly client: ProviderProfileApiClient,
    private readonly tenantScope: string | ProviderProfileTenantScope
  ) {}

  async listProviderProfiles(): Promise<readonly ProviderProfileSummary[]> {
    const tenantId = await this.requireTenant();
    const profiles = await this.client.listProviderProfiles(tenantId);
    return profiles.map(mapProviderProfile);
  }

  async getVoiceProviderConfiguration(): Promise<VoiceProviderConfiguration> {
    const tenantId = await this.requireTenant();
    return mapVoiceProviderConfiguration(await this.client.getVoiceProviderConfiguration(tenantId));
  }

  async updateVoiceProviderConfiguration(
    input: UpdateVoiceProviderConfigurationInput
  ): Promise<VoiceProviderConfiguration> {
    const tenantId = await this.requireTenant();
    return mapVoiceProviderConfiguration(
      await this.client.updateVoiceProviderConfiguration(tenantId, input)
    );
  }

  async createProviderProfile(input: CreateProviderProfileInput): Promise<ProviderProfileSummary> {
    const tenantId = await this.requireTenant();
    return mapProviderProfile(await this.client.createProviderProfile(tenantId, input));
  }

  async updateProviderProfile(input: UpdateProviderProfileInput): Promise<ProviderProfileSummary> {
    const tenantId = await this.requireTenant();
    return mapProviderProfile(
      await this.client.updateProviderProfile(tenantId, input.providerProfileId, {
        promptTemplate: input.promptTemplate
      })
    );
  }

  async replaceProviderProfileCredential(
    input: ReplaceProviderProfileCredentialInput
  ): Promise<ProviderProfileSummary> {
    const tenantId = await this.requireTenant();
    const body = input.purpose === 'server_adc'
      ? { purpose: input.purpose }
      : { purpose: input.purpose, credential: input.credential ?? '' };

    return mapProviderProfile(
      await this.client.replaceProviderProfileCredential(tenantId, input.providerProfileId, body)
    );
  }

  async changeProviderProfileLifecycle(
    providerProfileId: string,
    action: ProviderProfileLifecycleAction
  ): Promise<ProviderProfileSummary> {
    const tenantId = await this.requireTenant();
    switch (action) {
      case 'enable':
        return mapProviderProfile(await this.client.enableProviderProfile(tenantId, providerProfileId));
      case 'disable':
        return mapProviderProfile(await this.client.disableProviderProfile(tenantId, providerProfileId));
      case 'archive':
        return mapProviderProfile(await this.client.archiveProviderProfile(tenantId, providerProfileId));
    }
  }

  async testProviderProfile(providerProfileId: string): Promise<ProviderProfileTestResult> {
    const tenantId = await this.requireTenant();
    return this.client.testProviderProfile(tenantId, providerProfileId);
  }

  private async requireTenant(): Promise<string> {
    const tenantId = typeof this.tenantScope === 'string'
      ? this.tenantScope
      : await this.tenantScope.getCurrentTenantId();
    if (tenantId.trim().length === 0) {
      throw new Error('Complete mobile onboarding before managing voice provider profiles.');
    }
    return tenantId;
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
      recommendedAction: parseRecommendedAction(slot.recommendedAction),
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

function parseRecommendedAction(value: string): VoiceProviderRecommendedAction {
  switch (value) {
    case 'none':
    case 'add_profile':
    case 'choose_profile':
    case 'replace_credential':
    case 'enable_profile':
    case 'test_profile':
      return value;
    default:
      return 'none';
  }
}
