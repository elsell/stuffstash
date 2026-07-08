export type ProviderProfileCapability =
  | 'speech_to_text'
  | 'language_inference'
  | 'text_to_speech'
  | string;

export type ProviderCredentialPurpose = 'api_key' | 'oauth_bearer' | 'server_adc';

export type ProviderProfileSummary = {
  readonly id: string;
  readonly capability: ProviderProfileCapability;
  readonly providerKind: string;
  readonly displayName: string;
  readonly modelName: string;
  readonly credentialStatus: string;
  readonly credentialPurpose?: ProviderCredentialPurpose;
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

export type VoiceProviderSelectionSource = 'explicit' | 'implicit' | 'missing' | string;
export type VoiceProviderReadiness = 'ready' | 'needs_attention' | string;
export type VoiceProviderSlotReadiness =
  | 'ready'
  | 'missing'
  | 'disabled'
  | 'archived'
  | 'credential_missing'
  | 'untested'
  | 'duplicate_candidates'
  | 'invalid_selection'
  | string;
export type VoiceProviderRecommendedAction =
  | 'none'
  | 'add_profile'
  | 'choose_profile'
  | 'replace_credential'
  | 'enable_profile'
  | 'test_profile';

export type VoiceProviderSlot = {
  readonly capability: ProviderProfileCapability;
  readonly label: string;
  readonly selectedProfileId?: string;
  readonly selectedProfile?: ProviderProfileSummary;
  readonly selectionSource: VoiceProviderSelectionSource;
  readonly readiness: VoiceProviderSlotReadiness;
  readonly issues: readonly string[];
  readonly recommendedAction: VoiceProviderRecommendedAction;
  readonly duplicateProfiles: readonly ProviderProfileSummary[];
};

export type VoiceProviderConfiguration = {
  readonly tenantId: string;
  readonly readiness: VoiceProviderReadiness;
  readonly updatedAt?: string;
  readonly profileIds: {
    readonly speechToText?: string;
    readonly languageInference?: string;
    readonly textToSpeech?: string;
  };
  readonly slots: readonly VoiceProviderSlot[];
};

export type UpdateVoiceProviderConfigurationInput = {
  readonly speechToTextProfileId?: string;
  readonly languageInferenceProfileId?: string;
  readonly textToSpeechProfileId?: string;
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
  readonly purpose: ProviderCredentialPurpose;
  readonly credential?: string;
};

export type ProviderProfileLifecycleAction = 'enable' | 'disable' | 'archive';

export interface ProviderProfileRepository {
  listProviderProfiles(): Promise<readonly ProviderProfileSummary[]>;
  getVoiceProviderConfiguration(): Promise<VoiceProviderConfiguration>;
  updateVoiceProviderConfiguration(input: UpdateVoiceProviderConfigurationInput): Promise<VoiceProviderConfiguration>;
  createProviderProfile(input: CreateProviderProfileInput): Promise<ProviderProfileSummary>;
  updateProviderProfile(input: UpdateProviderProfileInput): Promise<ProviderProfileSummary>;
  replaceProviderProfileCredential(input: ReplaceProviderProfileCredentialInput): Promise<ProviderProfileSummary>;
  changeProviderProfileLifecycle(
    providerProfileId: string,
    action: ProviderProfileLifecycleAction
  ): Promise<ProviderProfileSummary>;
  testProviderProfile(providerProfileId: string): Promise<ProviderProfileTestResult>;
}
