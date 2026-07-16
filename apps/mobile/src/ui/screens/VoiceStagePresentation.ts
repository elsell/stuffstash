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
        title: 'Listen',
        description: 'Speech to text',
        longDescription: 'Choose the service that turns your spoken words into text.'
      };
    case 'language_inference':
      return {
        title: 'Understand',
        description: 'Language model',
        longDescription: 'Choose the service that interprets inventory requests and plans actions.'
      };
    case 'text_to_speech':
      return {
        title: 'Speak',
        description: 'Spoken responses',
        longDescription: 'Choose the service that reads Stuff Stash responses aloud.'
      };
    default:
      return {
        title: 'Voice Service',
        description: 'Unknown capability',
        longDescription: 'Review this voice service.'
      };
  }
}
