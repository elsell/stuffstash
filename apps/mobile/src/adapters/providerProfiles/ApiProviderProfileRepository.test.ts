import { describe, expect, it } from 'vitest';
import type { ProviderProfile, ProviderProfileTestResult } from '@stuff-stash/api-client';
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

  it('runs safe tests with the configured tenant scope', async () => {
    const client = new FakeProviderProfileClient();
    const repository = new ApiProviderProfileRepository(client, 'tenant-home');

    await expect(repository.testProviderProfile('profile-language')).resolves.toEqual(client.testResult);
    expect(client.testedTenantId).toBe('tenant-home');
    expect(client.testedProfileId).toBe('profile-language');
  });
});
