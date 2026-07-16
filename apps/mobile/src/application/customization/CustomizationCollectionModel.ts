import type { CustomizationLifecycle } from '../../domain/customization/Customization';

export type CustomizationCollectionState<Row> = {
  readonly lifecycle: CustomizationLifecycle;
  readonly pendingLifecycle?: CustomizationLifecycle;
  readonly rows: readonly Row[];
};

export function beginLifecycleTransition<Row>(
  state: CustomizationCollectionState<Row>,
  target: CustomizationLifecycle
): CustomizationCollectionState<Row> {
  if (target === state.lifecycle || state.pendingLifecycle) return state;
  return { ...state, pendingLifecycle: target };
}

export function commitLifecycleTransition<Row>(
  state: CustomizationCollectionState<Row>,
  target: CustomizationLifecycle,
  rows: readonly Row[]
): CustomizationCollectionState<Row> {
  if (state.pendingLifecycle && state.pendingLifecycle !== target) return state;
  return { lifecycle: target, pendingLifecycle: undefined, rows };
}

export function rollbackLifecycleTransition<Row>(
  state: CustomizationCollectionState<Row>
): CustomizationCollectionState<Row> {
  return { ...state, pendingLifecycle: undefined };
}
