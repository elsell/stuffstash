import { describe, expect, it } from 'vitest';
import {
  CreateProviderProfileInput,
  ProviderProfileLifecycleAction,
  ProviderProfileRepository,
  ProviderProfileSummary,
  ProviderProfileTestResult,
  ReplaceProviderProfileCredentialInput,
  UpdateProviderProfileInput
} from './ProviderProfileRepository';
import { ManageProviderProfileCommand } from './ManageProviderProfileCommand';
import { ProviderProfileSettingsQuery } from './ProviderProfileSettingsQuery';
import { recommendedProviderProfiles } from './RecommendedProviderProfiles';
import { TestProviderProfileCommand } from './TestProviderProfileCommand';

class FakeProviderProfileRepository implements ProviderProfileRepository {
  profiles: ProviderProfileSummary[] = [];
  testResults = new Map<string, ProviderProfileTestResult>();
  createdProfile: CreateProviderProfileInput | undefined;
  updatedProfile: UpdateProviderProfileInput | undefined;
  replacedCredential: ReplaceProviderProfileCredentialInput | undefined;
  lifecycleChange: { providerProfileId: string; action: ProviderProfileLifecycleAction } | undefined;

  async listProviderProfiles(): Promise<readonly ProviderProfileSummary[]> {
    return this.profiles;
  }

  async createProviderProfile(input: CreateProviderProfileInput): Promise<ProviderProfileSummary> {
    this.createdProfile = input;
    return profile({
      id: 'profile-created',
      capability: input.capability,
      providerKind: input.providerKind,
      displayName: input.displayName,
      modelName: input.modelName ?? ''
    });
  }

  async updateProviderProfile(input: UpdateProviderProfileInput): Promise<ProviderProfileSummary> {
    this.updatedProfile = input;
    return profile({
      id: input.providerProfileId,
      displayName: 'Updated profile',
      hasPromptTemplate: Boolean(input.promptTemplate)
    });
  }

  async replaceProviderProfileCredential(
    input: ReplaceProviderProfileCredentialInput
  ): Promise<ProviderProfileSummary> {
    this.replacedCredential = input;
    return profile({ id: input.providerProfileId, credentialStatus: 'configured' });
  }

  async changeProviderProfileLifecycle(
    providerProfileId: string,
    action: ProviderProfileLifecycleAction
  ): Promise<ProviderProfileSummary> {
    this.lifecycleChange = { providerProfileId, action };
    return profile({ id: providerProfileId, lifecycleState: action === 'enable' ? 'enabled' : action === 'disable' ? 'disabled' : 'archived' });
  }

  async testProviderProfile(providerProfileId: string): Promise<ProviderProfileTestResult> {
    const result = this.testResults.get(providerProfileId);
    if (!result) {
      throw new Error('Provider profile test failed safely.');
    }

    return result;
  }
}

describe('ProviderProfileSettingsQuery', () => {
  it('sorts safe profile metadata and reports missing ready capabilities', async () => {
    const repository = new FakeProviderProfileRepository();
    repository.profiles = [
      profile({ id: 'tts', capability: 'text_to_speech', lifecycleState: 'disabled' }),
      profile({ id: 'lm', capability: 'language_inference', hasPromptTemplate: true }),
      profile({ id: 'stt', capability: 'speech_to_text', lastTestedAt: undefined })
    ];
    const query = new ProviderProfileSettingsQuery(repository);

    await expect(query.execute()).resolves.toEqual({
      profiles: [repository.profiles[1], repository.profiles[2], repository.profiles[0]],
      missingCapabilities: ['speech_to_text', 'text_to_speech']
    });
  });
});

describe('TestProviderProfileCommand', () => {
  it('runs a safe profile diagnostic through the repository', async () => {
    const repository = new FakeProviderProfileRepository();
    repository.testResults.set('profile-language', {
      providerProfileId: 'profile-language',
      capability: 'language_inference',
      providerKind: 'gemini',
      status: 'success',
      message: 'Provider profile test succeeded.',
      testedAt: '2026-06-26T12:01:00Z'
    });
    const command = new TestProviderProfileCommand(repository);

    await expect(command.execute(' profile-language ')).resolves.toEqual({
      providerProfileId: 'profile-language',
      capability: 'language_inference',
      providerKind: 'gemini',
      status: 'success',
      message: 'Provider profile test succeeded.',
      testedAt: '2026-06-26T12:01:00Z'
    });
  });
});

describe('ManageProviderProfileCommand', () => {
  it('creates recommended cheap Gemini profiles through the repository', async () => {
    const repository = new FakeProviderProfileRepository();
    const command = new ManageProviderProfileCommand(repository);

    await command.createRecommended(recommendedProviderProfiles[1]);

    expect(repository.createdProfile).toMatchObject({
      capability: 'language_inference',
      providerKind: 'gemini',
      displayName: 'Gemini Flash-Lite language',
      modelName: 'gemini-2.5-flash-lite',
      runtimeOptions: { credentialType: 'api_key' }
    });
    expect(recommendedProviderProfiles[1].credentialPurpose).toBe('api_key');
  });

  it('routes prompt updates, credential replacement, and lifecycle actions', async () => {
    const repository = new FakeProviderProfileRepository();
    const command = new ManageProviderProfileCommand(repository);

    await command.replacePromptTemplate({
      providerProfileId: 'profile-language',
      promptTemplate: 'Answer briefly.'
    });
    await command.replaceCredential({
      providerProfileId: 'profile-language',
      purpose: 'api_key',
      credential: ' secret-api-key '
    });
    await command.changeLifecycle('profile-language', 'enable');

    expect(repository.updatedProfile).toEqual({
      providerProfileId: 'profile-language',
      promptTemplate: 'Answer briefly.'
    });
    expect(repository.replacedCredential).toEqual({
      providerProfileId: 'profile-language',
      purpose: 'api_key',
      credential: 'secret-api-key'
    });
    expect(repository.lifecycleChange).toEqual({
      providerProfileId: 'profile-language',
      action: 'enable'
    });
  });

  it('rejects blank prompt replacement so hidden existing prompts are not erased by accident', async () => {
    const repository = new FakeProviderProfileRepository();
    const command = new ManageProviderProfileCommand(repository);

    await expect(
      command.replacePromptTemplate({
        providerProfileId: 'profile-language',
        promptTemplate: ' '
      })
    ).rejects.toThrow('Enter a replacement prompt template.');
    expect(repository.updatedProfile).toBeUndefined();
  });
});

function profile(
  overrides: Partial<ProviderProfileSummary>
): ProviderProfileSummary {
  return {
    id: overrides.id ?? 'profile',
    capability: overrides.capability ?? 'speech_to_text',
    providerKind: overrides.providerKind ?? 'gemini',
    displayName: overrides.displayName ?? overrides.id ?? 'Profile',
    modelName: overrides.modelName ?? 'model',
    credentialStatus: overrides.credentialStatus ?? 'configured',
    credentialPurpose: overrides.credentialPurpose,
    lifecycleState: overrides.lifecycleState ?? 'enabled',
    lastTestedAt: 'lastTestedAt' in overrides
      ? overrides.lastTestedAt
      : '2026-06-26T12:00:00Z',
    hasPromptTemplate: overrides.hasPromptTemplate ?? false
  };
}
