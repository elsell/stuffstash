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

export interface ProviderProfileRepository {
  listProviderProfiles(): Promise<readonly ProviderProfileSummary[]>;
  testProviderProfile(providerProfileId: string): Promise<ProviderProfileTestResult>;
}
