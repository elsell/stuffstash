import { ProviderProfileSettingsQuery } from './ProviderProfileSettingsQuery';
import type { ProviderProfileCapability } from './ProviderProfileRepository';

type VoiceRequiredProviderCapability = 'speech_to_text' | 'language_inference' | 'text_to_speech';

export class VoiceProviderReadinessError extends Error {
  readonly code = 'provider_readiness';
  readonly missingCapabilities: readonly VoiceRequiredProviderCapability[];

  constructor(missingCapabilities: readonly ProviderProfileCapability[]) {
    const safeMissingCapabilities = missingCapabilities.filter(isVoiceRequiredProviderCapability);
    super(readinessMessage(safeMissingCapabilities));
    this.missingCapabilities = safeMissingCapabilities;
  }
}

export class ProviderProfileVoiceReadinessCheck {
  constructor(private readonly query: ProviderProfileSettingsQuery) {}

  async assertReady(): Promise<void> {
    const viewModel = await this.query.execute();
    if (viewModel.missingCapabilities.length === 0) {
      return;
    }

    throw new VoiceProviderReadinessError(viewModel.missingCapabilities);
  }
}

function readinessMessage(missingCapabilities: readonly VoiceRequiredProviderCapability[]): string {
  return missingCapabilities.length > 0
    ? `Voice provider profiles are not ready: ${missingCapabilities.join(', ')}.`
    : 'Voice provider profiles are not ready.';
}

function isVoiceRequiredProviderCapability(
  value: ProviderProfileCapability
): value is VoiceRequiredProviderCapability {
  return value === 'speech_to_text' || value === 'language_inference' || value === 'text_to_speech';
}
