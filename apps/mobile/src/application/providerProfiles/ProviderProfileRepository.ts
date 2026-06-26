export type ProviderProfileCapability =
  | 'speech_to_text'
  | 'language_inference'
  | 'text_to_speech'
  | string;

export type ProviderProfileSummary = {
  readonly id: string;
  readonly capability: ProviderProfileCapability;
  readonly providerKind: string;
  readonly displayName: string;
  readonly modelName: string;
  readonly credentialStatus: string;
  readonly credentialPurpose?: 'api_key' | 'oauth_bearer';
  readonly lifecycleState: string;
  readonly lastTestedAt?: string;
  readonly hasPromptTemplate: boolean;
};

export type ProviderProfileTestResult = {
  readonly providerProfileId: string;
  readonly capability: ProviderProfileCapability;
  readonly providerKind: string;
  readonly status: string;
  readonly message: string;
  readonly testedAt: string;
};

export type CreateProviderProfileInput = {
  readonly capability: ProviderProfileCapability;
  readonly providerKind: string;
  readonly displayName: string;
  readonly endpointUrl?: string;
  readonly modelName?: string;
  readonly runtimeOptions?: Record<string, unknown>;
  readonly capabilityMetadata?: Record<string, unknown>;
  readonly promptTemplate?: string;
};

export type UpdateProviderProfileInput = {
  readonly providerProfileId: string;
  readonly promptTemplate?: string;
};

export type ReplaceProviderProfileCredentialInput = {
  readonly providerProfileId: string;
  readonly purpose: 'api_key' | 'oauth_bearer';
  readonly credential: string;
};

export type ProviderProfileLifecycleAction = 'enable' | 'disable' | 'archive';

export interface ProviderProfileRepository {
  listProviderProfiles(): Promise<readonly ProviderProfileSummary[]>;
  createProviderProfile(input: CreateProviderProfileInput): Promise<ProviderProfileSummary>;
  updateProviderProfile(input: UpdateProviderProfileInput): Promise<ProviderProfileSummary>;
  replaceProviderProfileCredential(input: ReplaceProviderProfileCredentialInput): Promise<ProviderProfileSummary>;
  changeProviderProfileLifecycle(
    providerProfileId: string,
    action: ProviderProfileLifecycleAction
  ): Promise<ProviderProfileSummary>;
  testProviderProfile(providerProfileId: string): Promise<ProviderProfileTestResult>;
}
