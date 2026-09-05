import { describe, expect, it } from 'vitest';
import { ConversationFailure } from '$lib/domain/conversation';
import { createConversationSession } from './conversationSession';
const scope = { apiIdentity: 'api', principalId: 'owner', tenantId: 'home' };
function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>(done => { resolve = done; });
  return { promise, resolve };
}
describe('conversation session mutation ownership', () => {
  it('reconciles an active mutation result', async () => {
    const session = createConversationSession(scope, () => {});
    const results: string[] = [];
    expect(await session.mutate(async () => 'saved', value => { results.push(value); })).toBe('saved');
    expect(results).toEqual(['saved']);
    await session.dispose();
  });
  it('does not publish a late success into a disposed session', async () => {
    const work = deferred<string>(); const started = deferred<void>();
    const session = createConversationSession(scope, () => {});
    const results: string[] = [];
    const pending = session.mutate(() => { started.resolve(); return work.promise; }, value => { results.push(value); });
    await started.promise; await session.dispose(); work.resolve('saved');
    expect(await pending).toBe('saved'); expect(results).toEqual([]);
    let dispatched = false;
    await expect(session.mutate(async () => { dispatched = true; }, () => {})).rejects.toMatchObject({ name: 'AbortError' });
    expect(dispatched).toBe(false);
  });
  it('revocation fences concurrent late writes and clears cached fixtures', async () => {
    const work = deferred<string>(); const started = deferred<void>();
    let denied = 0;
    const session = createConversationSession(scope, () => { denied++; });
    session.client.setQueryData(['private'], 'fixture text');
    let reconciled = false;
    const pending = session.mutate(() => { started.resolve(); return work.promise; }, () => { reconciled = true; });
    await started.promise;
    await expect(session.mutate(async () => { throw new ConversationFailure('forbidden'); }, () => {})).rejects.toMatchObject({ kind: 'forbidden' });
    work.resolve('saved'); await pending;
    expect(denied).toBe(1); expect(reconciled).toBe(false); expect(session.client.getQueryData(['private'])).toBeUndefined();
    await session.dispose();
  });
});
