import { CatalogRecoveryError } from '../shared/CatalogRecoveryError';
import {
  ProviderProfileRepository,
  ProviderProfileTestResult
} from './ProviderProfileRepository';

export class TestProviderProfileCommand {
  constructor(private readonly profiles: ProviderProfileRepository) {}

  async execute(providerProfileId: string): Promise<ProviderProfileTestResult> {
    const trimmed = providerProfileId.trim();
    if (trimmed.length === 0) {
      throw new CatalogRecoveryError('providerTest.chooseProfile');
    }

    const result = await this.profiles.testProviderProfile(trimmed);
    if (result.status !== 'succeeded') {
      throw new CatalogRecoveryError('providerTest.failed');
    }
    return result;
  }
}
