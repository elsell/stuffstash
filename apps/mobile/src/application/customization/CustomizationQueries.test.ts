import { describe, expect, it } from 'vitest';
import type { CustomizationRepository } from './CustomizationRepository';
import { BufferedCustomizationObservability } from './CustomizationObservability';
import { CustomizationCollectionQuery } from './CustomizationQueries';

describe('CustomizationCollectionQuery', () => {
  it('loads every page before presenting local search as complete', async () => {
    const calls: Array<string | undefined> = [];
    const repository = {
      listTags: async (_context: never, cursor?: string) => {
        calls.push(cursor);
        return cursor ? { items: [{ kind: 'tag', id: '1', key: 'a', displayName: 'Alpha' }] }
          : { items: [{ kind: 'tag', id: '2', key: 'z', displayName: 'Zulu' }], nextCursor: 'next' };
      }
    } as unknown as CustomizationRepository;
    const result = await new CustomizationCollectionQuery(repository).tags({} as never);
    expect(calls).toEqual([undefined, 'next']);
    expect(result.items.map((item) => item.displayName)).toEqual(['Alpha', 'Zulu']);
    expect(result.complete).toBe(true);
  });

  it('stops a repeated cursor, removes duplicate records, and marks the result incomplete', async () => {
    const repository = {
      listTags: async (_context: never, cursor?: string) => ({
        items: [{ kind: 'tag', id: 'same', key: 'same', displayName: cursor ? 'Duplicate' : 'Original' }],
        nextCursor: 'stuck'
      })
    } as unknown as CustomizationRepository;

    const result = await new CustomizationCollectionQuery(repository).tags({} as never);

    expect(result).toMatchObject({ complete: false });
    expect(result.items.map((item) => item.displayName)).toEqual(['Original']);
  });

  it('records a safe domain event when a collection adapter fails', async () => {
    const observability = new BufferedCustomizationObservability();
    const repository = {
      listFields: async () => { throw new Error('private transport detail'); }
    } as unknown as CustomizationRepository;

    await expect(new CustomizationCollectionQuery(repository, observability).fields({} as never, 'tenant', 'active')).rejects.toThrow();
    expect(observability.events()).toEqual([{
      name: 'customization.collection_load_failed', resource: 'field', scope: 'tenant'
    }]);
    expect(JSON.stringify(observability.events())).not.toContain('private transport detail');
  });
});
