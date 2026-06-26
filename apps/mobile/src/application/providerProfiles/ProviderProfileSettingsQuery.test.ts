import { describe, expect, it } from 'vitest';
import {
  ProviderProfileRepository,
  ProviderProfileSummary,
  ProviderProfileTestResult
} from './ProviderProfileRepository';
import { ProviderProfileSettingsQuery } from './ProviderProfileSettingsQuery';
import { TestProviderProfileCommand } from './TestProviderProfileCommand';

class FakeProviderProfileRepository implements ProviderProfileRepository {
  profiles: ProviderProfileSummary[] = [];
  testResults = new Map<string, ProviderProfileTestResult>();

  async listProviderProfiles(): Promise<readonly ProviderProfileSummary[]> {
    return this.profiles;
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
    lifecycleState: overrides.lifecycleState ?? 'enabled',
    lastTestedAt: 'lastTestedAt' in overrides
      ? overrides.lastTestedAt
      : '2026-06-26T12:00:00Z',
    hasPromptTemplate: overrides.hasPromptTemplate ?? false
  };
}
