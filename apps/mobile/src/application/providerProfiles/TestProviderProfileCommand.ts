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

    const result = await this.profiles.testProviderProfile(trimmed);
    if (result.status !== 'succeeded') {
      throw new Error('Connection test failed. Check the profile configuration and credential, then try again.');
    }
    return result;
  }
}
