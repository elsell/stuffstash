import type { AssetActivityEntry } from '../../application/assets/AssetActivityQuery';
import { historyRevertConfirmation, historyRevertFailure } from './AssetHistoryPresentation';

export type HistoryRevertConfirmation = ReturnType<typeof historyRevertConfirmation>;
export type HistoryRevertConfirmationPresenter = (
  confirmation: HistoryRevertConfirmation,
  confirm: () => void
) => void;

export function requestHistoryRevertConfirmation(
  entry: AssetActivityEntry,
  presenter: HistoryRevertConfirmationPresenter,
  confirm: () => void
): void {
  presenter(historyRevertConfirmation(entry), confirm);
}

export type HistoryRevertInput = {
  readonly tenantId: string;
  readonly inventoryId: string;
  readonly operationId: string;
};

export interface HistoryRevertExecutor {
  execute(input: HistoryRevertInput): Promise<boolean>;
}

export interface HistoryRevertEffects {
  invalidateActivity(): void;
  showSuccess(): void;
  navigateBack(): void;
}

export type HistoryRevertResult =
  | { readonly status: 'applied' }
  | { readonly status: 'suppressed' }
  | { readonly status: 'failed'; readonly failure: ReturnType<typeof historyRevertFailure> };

export async function applyHistoryRevert(
  executor: HistoryRevertExecutor,
  input: HistoryRevertInput,
  effects: HistoryRevertEffects
): Promise<HistoryRevertResult> {
  try {
    const applied = await executor.execute(input);
    if (!applied) return { status: 'suppressed' };
    effects.invalidateActivity();
    effects.showSuccess();
    effects.navigateBack();
    return { status: 'applied' };
  } catch (error) {
    return { status: 'failed', failure: historyRevertFailure(error) };
  }
}
