import {
  ProviderProfileRepository,
  ProviderProfileSummary
} from './ProviderProfileRepository';

export type ProviderProfileSettingsViewModel = {
  readonly profiles: readonly ProviderProfileSummary[];
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
    const profiles = await this.profiles.listProviderProfiles();
    const readyCapabilities = new Set(
      profiles
        .filter(
          (profile) =>
            profile.lifecycleState === 'enabled' &&
            profile.credentialStatus === 'configured' &&
            Boolean(profile.lastTestedAt)
        )
        .map((profile) => profile.capability)
    );

    return {
      profiles: [...profiles].sort(compareProfiles),
      missingCapabilities: requiredCapabilities.filter((capability) => !readyCapabilities.has(capability))
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
