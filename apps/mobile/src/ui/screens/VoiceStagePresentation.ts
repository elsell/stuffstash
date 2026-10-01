import { t } from '../../presentation/localization';
import type { ProviderProfileCapability } from '../../application/providerProfiles/ProviderProfileRepository';

export type VoiceStagePresentation = {
  readonly title: string;
  readonly description: string;
  readonly longDescription: string;
};

export function stagePresentation(
  capability: ProviderProfileCapability
): VoiceStagePresentation {
  switch (capability) {
    case 'speech_to_text':
      return {
        title: t('mobile.VoiceStagePresentation.listen'),
        description: t('mobile.VoiceStagePresentation.speechToText'),
        longDescription: 'Choose the service that turns your spoken words into text.'
      };
    case 'language_inference':
      return {
        title: t('mobile.VoiceStagePresentation.understand'),
        description: t('mobile.VoiceStagePresentation.languageModel'),
        longDescription: 'Choose the service that interprets inventory requests and plans actions.'
      };
    case 'text_to_speech':
      return {
        title: t('mobile.VoiceStagePresentation.speak'),
        description: t('mobile.VoiceStagePresentation.spokenResponses'),
        longDescription: 'Choose the service that reads Stuff Stash responses aloud.'
      };
    default:
      return {
        title: t('mobile.VoiceStagePresentation.voiceService'),
        description: t('mobile.VoiceStagePresentation.unknownCapability'),
        longDescription: 'Review this voice service.'
      };
  }
}
