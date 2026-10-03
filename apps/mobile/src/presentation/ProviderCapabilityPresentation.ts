import { t } from './localization';

export function formatVoiceProviderCapabilityLabel(capability: string): string {
  switch (capability) {
    case 'speech_to_text':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.speechInput');
    case 'language_inference':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.agentBrain');
    case 'text_to_speech':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.spokenOutput');
    default:
      return t('mobile.ProviderProfilesVoiceSetupPresentation.unknownCapability');
  }
}

