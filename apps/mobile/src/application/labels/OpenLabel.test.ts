import { describe, expect, it } from 'vitest';
import { OpenLabel } from './OpenLabel';
import type { LabelRepository } from './LabelWorkspace';

describe('authorized label navigation', () => {
  function fixture() {
    const calls: string[] = [];
    const repository: Pick<LabelRepository, 'instance' | 'resolve'> = {
      instance: async () => { calls.push('instance'); return 'this-instance'; },
      resolve: async () => { calls.push('resolve'); return { tenantId: 'tenant', inventoryId: 'inventory', assetId: 'asset', archived: true }; }
    };
    const command = new OpenLabel(repository, async (_target, signal) => { calls.push('select'); if (signal.aborted) throw new Error('aborted'); });
    return { calls, repository, command };
  }
  it('resolves through the configured API and selects the authorized scope including archived assets', async () => {
    const f = fixture();
    expect(await f.command.execute({ instanceId: 'this-instance', labelId: 'label' }, new AbortController().signal)).toEqual({ tenantId: 'tenant', inventoryId: 'inventory', assetId: 'asset', archived: true });
    expect(f.calls).toEqual(['instance', 'resolve', 'select']);
  });
  it('never resolves a foreign instance or selects a denied label', async () => {
    const f = fixture();
    await expect(f.command.execute({ instanceId: 'foreign', labelId: 'label' }, new AbortController().signal)).rejects.toMatchObject({ code: 'wrong_instance' });
    expect(f.calls).toEqual(['instance']);
    f.repository.resolve = async () => { throw new Error('denied'); };
    await expect(f.command.execute({ instanceId: 'this-instance', labelId: 'label' }, new AbortController().signal)).rejects.toThrow('denied');
    expect(f.calls).not.toContain('select');
  });
  it('does not navigate/select after an auth or server change aborts a resolution', async () => {
    const f = fixture(); const controller = new AbortController();
    f.repository.resolve = async () => { controller.abort(); return { tenantId: 'tenant', inventoryId: 'inventory', assetId: 'asset', archived: false }; };
    await expect(f.command.execute({ instanceId: 'this-instance', labelId: 'label' }, controller.signal)).rejects.toBeDefined();
    expect(f.calls).not.toContain('select');
  });
});
