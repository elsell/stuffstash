import {
  ProviderProfileRepository,
  ProviderProfileSummary,
  VoiceProviderConfiguration
} from './ProviderProfileRepository';

export type ProviderProfileSettingsViewModel = {
  readonly profiles: readonly ProviderProfileSummary[];
  readonly configuration: VoiceProviderConfiguration;
  readonly missingCapabilities: readonly string[];
};

const requiredCapabilities = [
  'speech_to_text',
  'language_inference',
  'text_to_speech'
] as const;

export class ProviderProfileSettingsQuery {
  constructor(private readonly profiles: ProviderProfileRepository) {}

  async execute(): Promise<ProviderProfileSettingsViewModel> {
    const [profiles, configuration] = await Promise.all([
      this.profiles.listProviderProfiles(),
      this.profiles.getVoiceProviderConfiguration()
    ]);

    return {
      profiles: [...profiles].sort(compareProfiles),
      configuration,
      missingCapabilities: requiredCapabilities.filter(
        (capability) =>
          configuration.slots.some(
            (slot) => slot.capability === capability && slot.readiness !== 'ready'
          )
      )
    };
  }
}

function compareProfiles(a: ProviderProfileSummary, b: ProviderProfileSummary): number {
  const capability = a.capability.localeCompare(b.capability);
  if (capability !== 0) {
    return capability;
  }

  return a.displayName.localeCompare(b.displayName);
}
