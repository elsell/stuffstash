import { describe, expect, it } from 'vitest';
import { ConversationProviderAPIRepository } from './providerRepository';
const profile = { id: 'model', tenantId: 'home', capability: 'language_inference', lifecycleState: 'active',
  providerKind: 'gemini', displayName: 'Household model', modelName: 'configured-model',
  credentialStatus: 'configured', runtimeOptions: { private: 'settings' }, promptTemplate: 'private prompt' };
function repository(profiles: unknown[]) {
  return new ConversationProviderAPIRepository('https://api.example.test', () => 'session', async () => Response.json({ data: profiles, meta: { tenantId: 'home' } }));
}
describe('conversation model choices', () => {
  it('keeps active language profiles and exposes only choice fields', async () => {
    const choices = await repository([profile, { ...profile, id: 'speech', capability: 'speech_to_text' },
      { ...profile, id: 'retired', lifecycleState: 'retired' }]).list('home');
    expect(choices).toEqual([{ id: 'model', name: 'Household model', providerKind: 'gemini', modelName: 'configured-model' }]);
  });
  it('rejects foreign profiles before capability filtering', async () => {
    await expect(repository([{ ...profile, tenantId: 'other', capability: 'speech_to_text' }]).list('home'))
      .rejects.toMatchObject({ kind: 'invalid' });
  });
});
