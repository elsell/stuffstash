import {
  ProviderProfileLifecycleAction,
  ProviderProfileRepository,
  ProviderProfileSummary,
  ReplaceProviderProfileCredentialInput,
  UpdateProviderProfileInput,
  UpdateVoiceProviderConfigurationInput,
  VoiceProviderConfiguration
} from './ProviderProfileRepository';
import { RecommendedProviderProfileTemplate } from './RecommendedProviderProfiles';

export class ManageProviderProfileCommand {
  constructor(private readonly profiles: ProviderProfileRepository) {}

  async createRecommended(template: RecommendedProviderProfileTemplate): Promise<ProviderProfileSummary> {
    const input = template.input;
    return this.profiles.createProviderProfile({
      ...input,
      displayName: requireText(input.displayName, 'Name the provider profile.'),
      capability: requireText(input.capability, 'Choose a provider capability.'),
      providerKind: requireText(input.providerKind, 'Choose a provider kind.')
    });
  }

  async replacePromptTemplate(input: UpdateProviderProfileInput): Promise<ProviderProfileSummary> {
    const promptTemplate = requireText(input.promptTemplate ?? '', 'Enter a replacement prompt template.');
    return this.profiles.updateProviderProfile({
      providerProfileId: requireText(input.providerProfileId, 'Choose a provider profile.'),
      promptTemplate
    });
  }

  async replaceCredential(
    input: ReplaceProviderProfileCredentialInput
  ): Promise<ProviderProfileSummary> {
    return this.profiles.replaceProviderProfileCredential({
      providerProfileId: requireText(input.providerProfileId, 'Choose a provider profile.'),
      purpose: input.purpose,
      credential: requireText(input.credential, 'Enter the provider credential.')
    });
  }

  async changeLifecycle(
    providerProfileId: string,
    action: ProviderProfileLifecycleAction
  ): Promise<ProviderProfileSummary> {
    return this.profiles.changeProviderProfileLifecycle(
      requireText(providerProfileId, 'Choose a provider profile.'),
      action
    );
  }

  async updateVoiceProviderConfiguration(
    input: UpdateVoiceProviderConfigurationInput
  ): Promise<VoiceProviderConfiguration> {
    return this.profiles.updateVoiceProviderConfiguration(input);
  }
}

function requireText(value: string, message: string): string {
  const trimmed = value.trim();
  if (trimmed.length === 0) {
    throw new Error(message);
  }

  return trimmed;
}
