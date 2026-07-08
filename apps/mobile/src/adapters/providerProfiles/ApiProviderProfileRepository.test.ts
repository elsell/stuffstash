import { describe, expect, it } from 'vitest';
import type {
  ProviderProfile,
  ProviderProfileTestResult,
  UpdateVoiceProviderConfigurationInput,
  VoiceProviderConfiguration
} from '@stuff-stash/api-client';
import { ApiProviderProfileRepository } from './ApiProviderProfileRepository';

class FakeProviderProfileClient {
  profiles: ProviderProfile[] = [];
  testResult: ProviderProfileTestResult = {
    providerProfileId: 'profile-language',
    capability: 'language_inference',
    providerKind: 'gemini',
    status: 'success',
    message: 'Provider profile test succeeded.',
    testedAt: '2026-06-26T12:01:00Z'
  };
  listedTenantId: string | undefined;
  testedTenantId: string | undefined;
  testedProfileId: string | undefined;
  voiceConfiguration: VoiceProviderConfiguration = voiceProviderConfiguration();
  updatedVoiceInput: UpdateVoiceProviderConfigurationInput | undefined;
  replacedCredentialInput: unknown;

  async listProviderProfiles(tenantId: string): Promise<ProviderProfile[]> {
    this.listedTenantId = tenantId;
    return this.profiles;
  }

  async testProviderProfile(
    tenantId: string,
    providerProfileId: string
  ): Promise<ProviderProfileTestResult> {
    this.testedTenantId = tenantId;
    this.testedProfileId = providerProfileId;
    return this.testResult;
  }

  async getVoiceProviderConfiguration(): Promise<VoiceProviderConfiguration> {
    return this.voiceConfiguration;
  }

  async updateVoiceProviderConfiguration(
    _tenantId: string,
    input: UpdateVoiceProviderConfigurationInput
  ): Promise<VoiceProviderConfiguration> {
    this.updatedVoiceInput = input;
    this.voiceConfiguration = voiceProviderConfiguration({
      profileIds: {
        speechToText: input.speechToTextProfileId,
        languageInference: input.languageInferenceProfileId,
        textToSpeech: input.textToSpeechProfileId
      }
    });
    return this.voiceConfiguration;
  }

  async createProviderProfile(
    _tenantId: string,
    input: Parameters<ApiProviderProfileRepository['createProviderProfile']>[0]
  ): Promise<ProviderProfile> {
    return {
      ...providerProfile(),
      capability: input.capability,
      providerKind: input.providerKind,
      displayName: input.displayName,
      modelName: input.modelName ?? ''
    };
  }

  async updateProviderProfile(
    _tenantId: string,
    providerProfileId: string,
    input: {
      promptTemplate?: string;
    }
  ): Promise<ProviderProfile> {
    return {
      ...providerProfile(),
      id: providerProfileId,
      promptTemplate: input.promptTemplate
    };
  }

  async replaceProviderProfileCredential(
    _tenantId: string,
    _providerProfileId: string,
    input: unknown
  ): Promise<ProviderProfile> {
    this.replacedCredentialInput = input;
    return {
      ...providerProfile(),
      credentialStatus: 'configured'
    };
  }

  async enableProviderProfile(): Promise<ProviderProfile> {
    return { ...providerProfile(), lifecycleState: 'enabled' };
  }

  async disableProviderProfile(): Promise<ProviderProfile> {
    return { ...providerProfile(), lifecycleState: 'disabled' };
  }

  async archiveProviderProfile(): Promise<ProviderProfile> {
    return { ...providerProfile(), lifecycleState: 'archived' };
  }
}

describe('ApiProviderProfileRepository', () => {
  it('maps provider profiles to safe mobile summaries', async () => {
    const client = new FakeProviderProfileClient();
    client.profiles = [
      {
        id: 'profile-language',
        tenantId: 'tenant-home',
        capability: 'language_inference',
        providerKind: 'gemini',
        displayName: 'Gemini language',
        endpointUrl: 'https://generativelanguage.googleapis.com',
        modelName: 'gemini-2.5-flash-lite',
        runtimeOptions: { credentialType: 'api_key', projectId: 'hidden-project' },
        capabilityMetadata: { structuredOutput: true },
        promptTemplate: 'This full prompt must not render in mobile summaries.',
        credentialStatus: 'configured',
        lifecycleState: 'enabled',
        lastTestedAt: '2026-06-26T12:00:00Z',
        createdAt: '2026-06-26T11:00:00Z',
        updatedAt: '2026-06-26T12:00:00Z'
      }
    ];
    const repository = new ApiProviderProfileRepository(client, 'tenant-home');

    const summaries = await repository.listProviderProfiles();

    expect(summaries).toEqual([
      {
        id: 'profile-language',
        capability: 'language_inference',
        providerKind: 'gemini',
        displayName: 'Gemini language',
        modelName: 'gemini-2.5-flash-lite',
        credentialStatus: 'configured',
        credentialPurpose: 'api_key',
        lifecycleState: 'enabled',
        lastTestedAt: '2026-06-26T12:00:00Z',
        hasPromptTemplate: true
      }
    ]);
    expect(summaries[0]).not.toHaveProperty('endpointUrl');
    expect(summaries[0]).not.toHaveProperty('promptTemplate');
    expect(summaries[0]).not.toHaveProperty('runtimeOptions');
    expect(summaries[0]).not.toHaveProperty('capabilityMetadata');
    expect(client.listedTenantId).toBe('tenant-home');
  });

  it('maps server ADC profile metadata and omits credential material on replacement', async () => {
    const client = new FakeProviderProfileClient();
    client.profiles = [
      {
        ...providerProfile(),
        id: 'profile-tts',
        capability: 'text_to_speech',
        displayName: 'Google Cloud voice',
        runtimeOptions: { credentialType: 'server_adc', languageCode: 'en-US', voiceName: 'en-US-Standard-C' }
      }
    ];
    const repository = new ApiProviderProfileRepository(client, 'tenant-home');

    await expect(repository.listProviderProfiles()).resolves.toMatchObject([
      {
        id: 'profile-tts',
        credentialPurpose: 'server_adc'
      }
    ]);
    await repository.replaceProviderProfileCredential({
      providerProfileId: 'profile-tts',
      purpose: 'server_adc'
    });

    expect(client.replacedCredentialInput).toEqual({ purpose: 'server_adc' });
  });

  it('runs safe tests with the configured tenant scope', async () => {
    const client = new FakeProviderProfileClient();
    const repository = new ApiProviderProfileRepository(client, 'tenant-home');

    await expect(repository.testProviderProfile('profile-language')).resolves.toEqual(client.testResult);
    expect(client.testedTenantId).toBe('tenant-home');
    expect(client.testedProfileId).toBe('profile-language');
  });

  it('maps voice provider configuration slots without exposing provider internals', async () => {
    const client = new FakeProviderProfileClient();
    client.voiceConfiguration = voiceProviderConfiguration({
      slots: [
        {
          capability: 'speech_to_text',
          label: 'Speech input',
          selectedProfileId: 'profile-stt',
          selectedProfile: {
            id: 'profile-stt',
            capability: 'speech_to_text',
            providerKind: 'gemini',
            displayName: 'Gemini speech',
            modelName: 'gemini-2.5-flash-lite',
            credentialStatus: 'configured',
            credentialPurpose: 'server_adc',
            lifecycleState: 'enabled',
            lastTestedAt: '2026-06-26T12:00:00Z'
          },
          selectionSource: 'explicit',
          readiness: 'ready',
          issues: ['providerSessionId:abc123'],
          recommendedAction: 'providerSessionId:action',
          duplicateProfiles: []
        }
      ]
    });
    const repository = new ApiProviderProfileRepository(client, 'tenant-home');

    await expect(repository.getVoiceProviderConfiguration()).resolves.toMatchObject({
      readiness: 'ready',
      slots: [
        {
          label: 'Speech input',
          issues: ['providerSessionId:abc123'],
          recommendedAction: 'none',
          selectedProfile: {
            id: 'profile-stt',
            credentialPurpose: 'server_adc',
            hasPromptTemplate: false
          }
        }
      ]
    });

    await repository.updateVoiceProviderConfiguration({
      speechToTextProfileId: 'profile-stt',
      languageInferenceProfileId: 'profile-lm',
      textToSpeechProfileId: 'profile-tts'
    });
    expect(client.updatedVoiceInput).toEqual({
      speechToTextProfileId: 'profile-stt',
      languageInferenceProfileId: 'profile-lm',
      textToSpeechProfileId: 'profile-tts'
    });
  });


  it('maps management responses back to safe summaries', async () => {
    const client = new FakeProviderProfileClient();
    const repository = new ApiProviderProfileRepository(client, 'tenant-home');

    await expect(
      repository.createProviderProfile({
        capability: 'language_inference',
        providerKind: 'gemini',
        displayName: 'Gemini language',
        modelName: 'gemini-2.5-flash-lite'
      })
    ).resolves.toMatchObject({
      displayName: 'Gemini language',
      modelName: 'gemini-2.5-flash-lite'
    });
    await expect(
      repository.updateProviderProfile({
        providerProfileId: 'profile-language',
        promptTemplate: 'Do not expose this prompt in summary.'
      })
    ).resolves.toMatchObject({
      id: 'profile-language',
      hasPromptTemplate: true
    });
    const credentialResult = await repository.replaceProviderProfileCredential({
      providerProfileId: 'profile-language',
      purpose: 'api_key',
      credential: 'secret-api-key'
    });

    expect(credentialResult.credentialStatus).toBe('configured');
    expect(credentialResult).not.toHaveProperty('credential');
    await expect(repository.changeProviderProfileLifecycle('profile-language', 'archive')).resolves.toMatchObject({
      lifecycleState: 'archived'
    });
  });
});

function providerProfile(): ProviderProfile {
  return {
    id: 'profile-language',
    tenantId: 'tenant-home',
    capability: 'language_inference',
    providerKind: 'gemini',
    displayName: 'Gemini language',
    endpointUrl: 'https://generativelanguage.googleapis.com',
    modelName: 'gemini-2.5-flash-lite',
    runtimeOptions: {},
    capabilityMetadata: {},
    credentialStatus: 'missing',
    lifecycleState: 'disabled',
    createdAt: '2026-06-26T11:00:00Z',
    updatedAt: '2026-06-26T12:00:00Z'
  };
}

function voiceProviderConfiguration(
  overrides: Partial<VoiceProviderConfiguration> = {}
): VoiceProviderConfiguration {
  return {
    tenantId: overrides.tenantId ?? 'tenant-home',
    readiness: overrides.readiness ?? 'ready',
    updatedAt: overrides.updatedAt,
    profileIds: overrides.profileIds ?? {},
    slots: overrides.slots ?? []
  };
}
