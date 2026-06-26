import {
  ProviderProfileRepository,
  ProviderProfileTestResult
} from './ProviderProfileRepository';

export class TestProviderProfileCommand {
  constructor(private readonly profiles: ProviderProfileRepository) {}

  async execute(providerProfileId: string): Promise<ProviderProfileTestResult> {
    const trimmed = providerProfileId.trim();
    if (trimmed.length === 0) {
      throw new Error('Choose a provider profile to test.');
    }

    return this.profiles.testProviderProfile(trimmed);
  }
}
