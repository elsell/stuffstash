import { ProviderProfileSummary } from '../../application/providerProfiles/ProviderProfileRepository';

export type CredentialEditorPresentation = {
  readonly profileId: string;
  readonly profileName: string;
  readonly purpose: 'api_key' | 'oauth_bearer';
  readonly value: string;
};

export type PromptEditorPresentation = {
  readonly profileId: string;
  readonly profileName: string;
  readonly value: string;
};

export function buildCredentialEditorPresentation(
  profile: ProviderProfileSummary
): CredentialEditorPresentation | undefined {
  if (!profile.credentialPurpose) {
    return undefined;
  }

  return {
    profileId: profile.id,
    profileName: profile.displayName,
    purpose: profile.credentialPurpose,
    value: ''
  };
}

export function buildPromptEditorPresentation(
  profile: ProviderProfileSummary
): PromptEditorPresentation {
  return {
    profileId: profile.id,
    profileName: profile.displayName,
    value: ''
  };
}
