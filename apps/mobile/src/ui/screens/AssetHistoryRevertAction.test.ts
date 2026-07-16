import { describe, expect, it } from 'vitest';
import type { AssetActivityEntry } from '../../application/assets/AssetActivityQuery';
import type { HistoryRevertEffects, HistoryRevertExecutor, HistoryRevertInput } from './AssetHistoryRevertAction';
import { applyHistoryRevert, requestHistoryRevertConfirmation } from './AssetHistoryRevertAction';

class FakeExecutor implements HistoryRevertExecutor {
  readonly inputs: HistoryRevertInput[] = [];
  result = true;
  failure: unknown;
  constructor(private readonly trace: string[] = []) {}

  async execute(input: HistoryRevertInput): Promise<boolean> {
    this.trace.push('execute');
    this.inputs.push(input);
    if (this.failure) throw this.failure;
    return this.result;
  }
}

class FakeEffects implements HistoryRevertEffects {
  readonly calls: string[] = [];
  constructor(private readonly trace: string[] = []) {}
  invalidateActivity(): void { this.calls.push('invalidate'); this.trace.push('invalidate'); }
  showSuccess(): void { this.calls.push('success'); this.trace.push('success'); }
  navigateBack(): void { this.calls.push('back'); this.trace.push('back'); }
}

describe('applyHistoryRevert', () => {
  const input = { tenantId: 'tenant-home', inventoryId: 'inventory-garage', operationId: 'operation-one' };

  it('applies the selected scoped operation before invalidating History and navigating back', async () => {
    const trace: string[] = [];
    const executor = new FakeExecutor(trace);
    const effects = new FakeEffects(trace);

    await expect(applyHistoryRevert(executor, input, effects)).resolves.toEqual({ status: 'applied' });

    expect(executor.inputs).toEqual([input]);
    expect(effects.calls).toEqual(['invalidate', 'success', 'back']);
    expect(trace).toEqual(['execute', 'invalidate', 'success', 'back']);
  });

  it('does not execute until the native confirmation presenter invokes Confirm', () => {
    const calls: string[] = [];
    let confirm: (() => void) | undefined;

    requestHistoryRevertConfirmation(activityEntry(), (confirmation, onConfirm) => {
      calls.push(confirmation.confirmLabel);
      confirm = onConfirm;
    }, () => calls.push('execute'));

    expect(calls).toEqual(['Revert Change']);
    confirm?.();
    expect(calls).toEqual(['Revert Change', 'execute']);
  });

  it('does not report success or navigate when repeated activation is suppressed', async () => {
    const executor = new FakeExecutor();
    executor.result = false;
    const effects = new FakeEffects();

    await expect(applyHistoryRevert(executor, input, effects)).resolves.toEqual({ status: 'suppressed' });
    expect(effects.calls).toEqual([]);
  });

  it('returns a terminal safe failure without applying success effects when access was revoked', async () => {
    const executor = new FakeExecutor();
    executor.failure = { status: 403 };
    const effects = new FakeEffects();

    await expect(applyHistoryRevert(executor, input, effects)).resolves.toEqual({
      status: 'failed',
      failure: {
        title: 'Revert unavailable',
        message: 'You no longer have permission to revert this change.',
        isTerminal: true
      }
    });
    expect(effects.calls).toEqual([]);
  });
});

function activityEntry(): AssetActivityEntry {
  return {
    id: 'audit-one', principalId: 'principal-one', action: 'asset.updated', category: 'change', source: 'api',
    occurredAt: '2026-07-14T15:00:00Z', changes: [{ field: 'title', previousValue: 'Drill', currentValue: 'Cordless drill' }],
    undo: { operationId: 'operation-one', status: 'available' }, technical: {}
  };
}
