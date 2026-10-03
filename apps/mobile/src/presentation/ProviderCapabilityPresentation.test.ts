import { expect, it } from 'vitest';
import { t } from './localization';
import { formatVoiceProviderCapabilityLabel } from './ProviderCapabilityPresentation';
import { VoiceProviderReadinessError } from '../application/providerProfiles/ProviderProfileVoiceReadinessCheck';
import { buildFailedVoiceRealtimeState } from '../ui/navigation/VoiceInteractionStateContext';

it('uses catalog labels across readiness and voice recovery without exposing unknown capabilities', () => {
  const error = new VoiceProviderReadinessError(['text_to_speech', 'private endpoint']);
  const label = t('mobile.ProviderProfilesVoiceSetupPresentation.spokenOutput');
  expect(formatVoiceProviderCapabilityLabel('text_to_speech')).toBe(label);
  expect(error.missingCapabilities).toEqual(['text_to_speech']);
  expect(error.message).toBe(t('provider.readinessMissing', { capabilities: label }));
  expect(buildFailedVoiceRealtimeState(error).errorMessage).toBe(t('mobile.VoiceInteractionStateContext.voiceProviderProfilesAreNotReady', { value: label }));
});
