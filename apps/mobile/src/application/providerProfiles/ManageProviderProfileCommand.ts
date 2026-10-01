import { t } from '../../presentation/localization';
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
      displayName: requireText(input.displayName, t('mobile.ManageProviderProfileCommand.nameTheProviderProfile')),
      capability: requireText(input.capability, t('mobile.ManageProviderProfileCommand.chooseAProviderCapability')),
      providerKind: requireText(input.providerKind, t('mobile.ManageProviderProfileCommand.chooseAProviderKind'))
    });
  }

  async replacePromptTemplate(input: UpdateProviderProfileInput): Promise<ProviderProfileSummary> {
    const promptTemplate = requireText(input.promptTemplate ?? '', t('mobile.ManageProviderProfileCommand.enterAReplacementPromptTemplate'));
    return this.profiles.updateProviderProfile({
      providerProfileId: requireText(input.providerProfileId, t('mobile.ManageProviderProfileCommand.chooseAProviderProfile')),
      promptTemplate
    });
  }

  async replaceCredential(
    input: ReplaceProviderProfileCredentialInput
  ): Promise<ProviderProfileSummary> {
    const credential = input.purpose === 'server_adc'
      ? undefined
      : requireText(input.credential ?? '', t('mobile.ManageProviderProfileCommand.enterTheProviderCredential'));

    return this.profiles.replaceProviderProfileCredential({
      providerProfileId: requireText(input.providerProfileId, t('mobile.ManageProviderProfileCommand.chooseAProviderProfile')),
      purpose: input.purpose,
      credential
    });
  }

  async changeLifecycle(
    providerProfileId: string,
    action: ProviderProfileLifecycleAction
  ): Promise<ProviderProfileSummary> {
    return this.profiles.changeProviderProfileLifecycle(
      requireText(providerProfileId, t('mobile.ManageProviderProfileCommand.chooseAProviderProfile')),
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
