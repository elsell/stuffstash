import { describe, expect, it } from 'vitest';
import { ConversationFailure } from '$lib/domain/conversation';
import { conversationKey, createConversationQueryClient, disposeConversationQueryClient, runPollInterval } from './conversationQueryClient';
const scope = { apiIdentity: 'https://example.test/api', principalId: 'owner', tenantId: 'home' };
describe('conversation server-state ownership', () => {
  it('separates API, account, tenant and revision identities', () => {
    const key = conversationKey(scope, 'workflow', 'one', 'first');
    for (const change of [{ tenantId: 'other' }, { principalId: 'other' }, { apiIdentity: 'https://other.test' }]) expect(key).not.toEqual(conversationKey({ ...scope, ...change }, 'workflow', 'one', 'first'));
    expect(key).not.toEqual(conversationKey(scope, 'workflow', 'one', 'second'));
  });
  it('deduplicates in-flight reads and reuses fresh results', async () => {
    const client = createConversationQueryClient(() => {});
    let calls = 0;
    const options = { queryKey: conversationKey(scope, 'workflows'), queryFn: async () => { calls++; return ['Household']; } };
    await Promise.all([client.fetchQuery(options), client.fetchQuery(options)]);
    expect(await client.fetchQuery(options)).toEqual(['Household']);
    expect(calls).toBe(1);
    await disposeConversationQueryClient(client);
  });
  it('clears private fixtures and never retries denied requests', async () => {
    let denied = 0;
    const client = createConversationQueryClient(() => { denied++; });
    client.setQueryData(conversationKey(scope, 'case', 'private'), { utterance: 'Private household case' });
    let calls = 0;
    await expect(client.fetchQuery({ queryKey: conversationKey(scope, 'runs'), queryFn: async () => { calls++; throw new ConversationFailure('forbidden'); } })).rejects.toThrow();
    expect(calls).toBe(1);
    expect(denied).toBe(1);
    expect(client.getQueryCache().getAll()).toHaveLength(0);
    await disposeConversationQueryClient(client);
  });
  it('aborts transport before clearing on context disposal', async () => {
    const client = createConversationQueryClient(() => {});
    let signal: AbortSignal | undefined;
    const pending = client.fetchQuery({ queryKey: conversationKey(scope, 'runs'), queryFn: ({ signal: current }) => { signal = current; return new Promise<string>(() => {}); } });
    const observed = pending.catch(() => undefined);
    await disposeConversationQueryClient(client);
    await observed;
    expect(signal?.aborted).toBe(true);
    expect(client.getQueryCache().getAll()).toHaveLength(0);
  });
  it('bounds polling and stops hidden or completed runs', () => {
    expect(runPollInterval('queued', 0, true)).toBe(2000);
    expect(runPollInterval('running', 2, true)).toBe(8000);
    expect(runPollInterval('running', 10, true)).toBe(30000);
    expect(runPollInterval('running', 0, false)).toBe(false);
    for (const state of ['succeeded', 'failed', 'cancelled'] as const) expect(runPollInterval(state, 0, true)).toBe(false);
  });
});
