import { t } from '../../presentation/localization';
import {
  CreateProviderProfileInput,
  ProviderCredentialPurpose,
  ProviderProfileCapability
} from './ProviderProfileRepository';

export type RecommendedProviderProfileTemplate = {
  readonly key: string;
  readonly title: string;
  readonly description: string;
  readonly credentialPurpose: ProviderCredentialPurpose;
  readonly input: CreateProviderProfileInput;
};

export const recommendedProviderProfiles: readonly RecommendedProviderProfileTemplate[] = [
  {
    key: 'gemini-stt-api-key',
    title: t('mobile.RecommendedProviderProfiles.geminiSpeechToText'),
    description: t('mobile.RecommendedProviderProfiles.cheapestCurrentGooglePathForTranscribingLocalVoiceTests'),
    credentialPurpose: 'api_key',
    input: geminiProfile('speech_to_text', t('mobile.RecommendedProviderProfiles.geminiFlashLiteSpeechToText'))
  },
  {
    key: 'gemini-language-api-key',
    title: t('mobile.RecommendedProviderProfiles.geminiLanguageInference'),
    description: t('mobile.RecommendedProviderProfiles.validatedModelForInventoryChangesExpirationDatesAndAnswers'),
    credentialPurpose: 'api_key',
    input: {
      ...geminiProfile('language_inference', t('mobile.RecommendedProviderProfiles.geminiFlashLanguage')),
      modelName: 'gemini-2.5-flash',
      promptTemplate: ''
    }
  },
  {
    key: 'google-cloud-tts-server-adc',
    title: t('mobile.RecommendedProviderProfiles.googleCloudTextToSpeech'),
    description: t('mobile.RecommendedProviderProfiles.standardVoiceForSpokenResponsesUsingServerApplicationDefault'),
    credentialPurpose: 'server_adc',
    input: {
      capability: 'text_to_speech',
      providerKind: 'gemini',
      displayName: t('mobile.RecommendedProviderProfiles.googleCloudStandardVoice'),
      runtimeOptions: {
        credentialType: 'server_adc',
        languageCode: 'en-US',
        voiceName: 'en-US-Standard-C'
      },
      capabilityMetadata: {
        audioFormat: 'mp3'
      }
    }
  }
] as const;

function geminiProfile(
  capability: ProviderProfileCapability,
  displayName: string
): CreateProviderProfileInput {
  return {
    capability,
    providerKind: 'gemini',
    displayName,
    modelName: 'gemini-2.5-flash-lite',
    runtimeOptions: {
      credentialType: 'api_key'
    },
    capabilityMetadata: {
      recommendedForMobileTesting: true
    }
  };
}
