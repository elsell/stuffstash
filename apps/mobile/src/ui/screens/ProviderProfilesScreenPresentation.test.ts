import { describe, expect, it } from 'vitest';
import { ProviderProfileSummary } from '../../application/providerProfiles/ProviderProfileRepository';
import {
  buildCredentialEditorPresentation,
  buildPromptEditorPresentation
} from './ProviderProfilesScreenPresentation';

describe('ProviderProfilesScreenPresentation', () => {
  it('does not expose credential replacement when credential purpose is unknown', () => {
    expect(buildCredentialEditorPresentation(profile({ credentialPurpose: undefined }))).toBeUndefined();
  });

  it('opens credential replacement with an empty secret field', () => {
    expect(buildCredentialEditorPresentation(profile({ credentialPurpose: 'api_key' }))).toEqual({
      profileId: 'profile-language',
      profileName: 'Gemini language',
      purpose: 'api_key',
      value: ''
    });
  });

  it('opens server ADC replacement without requiring secret text', () => {
    expect(buildCredentialEditorPresentation(profile({ credentialPurpose: 'server_adc' }))).toEqual({
      profileId: 'profile-language',
      profileName: 'Gemini language',
      purpose: 'server_adc',
      value: ''
    });
  });

  it('opens prompt replacement without rendering hidden existing prompt text', () => {
    expect(buildPromptEditorPresentation(profile({ hasPromptTemplate: true }))).toEqual({
      profileId: 'profile-language',
      profileName: 'Gemini language',
      value: ''
    });
  });
});

function profile(overrides: Partial<ProviderProfileSummary>): ProviderProfileSummary {
  return {
    id: 'profile-language',
    capability: 'language_inference',
    providerKind: 'gemini',
    displayName: 'Gemini language',
    modelName: 'gemini-2.5-flash-lite',
    credentialStatus: 'configured',
    lifecycleState: 'enabled',
    hasPromptTemplate: false,
    ...overrides
  };
}
