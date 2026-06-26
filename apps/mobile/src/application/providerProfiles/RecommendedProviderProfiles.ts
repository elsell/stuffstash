import { CreateProviderProfileInput, ProviderProfileCapability } from './ProviderProfileRepository';

export type RecommendedProviderProfileTemplate = {
  readonly key: string;
  readonly title: string;
  readonly description: string;
  readonly credentialPurpose: 'api_key' | 'oauth_bearer';
  readonly input: CreateProviderProfileInput;
};

export const recommendedProviderProfiles: readonly RecommendedProviderProfileTemplate[] = [
  {
    key: 'gemini-stt-api-key',
    title: 'Gemini speech-to-text',
    description: 'Cheapest current Google path for transcribing local voice tests.',
    credentialPurpose: 'api_key',
    input: geminiProfile('speech_to_text', 'Gemini Flash-Lite speech-to-text')
  },
  {
    key: 'gemini-language-api-key',
    title: 'Gemini language inference',
    description: 'Cheap model for inventory tool calls and spoken answers.',
    credentialPurpose: 'api_key',
    input: {
      ...geminiProfile('language_inference', 'Gemini Flash-Lite language'),
      promptTemplate: ''
    }
  },
  {
    key: 'google-cloud-tts-oauth',
    title: 'Google Cloud text-to-speech',
    description: 'Standard voice for spoken responses using an OAuth bearer token.',
    credentialPurpose: 'oauth_bearer',
    input: {
      capability: 'text_to_speech',
      providerKind: 'gemini',
      displayName: 'Google Cloud Standard voice',
      runtimeOptions: {
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
