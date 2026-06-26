import { ProviderProfileSettingsQuery } from './ProviderProfileSettingsQuery';

export class ProviderProfileVoiceReadinessCheck {
  constructor(private readonly query: ProviderProfileSettingsQuery) {}

  async assertReady(): Promise<void> {
    const viewModel = await this.query.execute();
    if (viewModel.missingCapabilities.length === 0) {
      return;
    }

    throw new Error(
      `Voice provider profiles are not ready: ${viewModel.missingCapabilities.join(', ')}.`
    );
  }
}
