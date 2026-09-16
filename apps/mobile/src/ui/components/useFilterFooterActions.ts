import { useCallback, useLayoutEffect, useRef } from 'react';
import { useFocusEffect } from 'expo-router';
import type { NativeFilterSheetProps } from './NativeFilterSheet.types';

/** Native callbacks belong to the focused filter task and its committed draft. */
export function useFilterFooterActions(actions: NativeFilterSheetProps['actions']) {
  const current = useRef<typeof actions | undefined>(undefined);
  const focused = useRef(false);
  useLayoutEffect(() => {
    current.current = actions;
    return () => { current.current = undefined; };
  }, [actions]);
  useFocusEffect(useCallback(() => {
    focused.current = true;
    return () => { focused.current = false; };
  }, []));
  const onApply = useCallback(() => {
    const latest = focused.current ? current.current : undefined;
    if (latest && !latest.disabled) latest.onApply();
  }, []);
  const onBack = useCallback(() => {
    const latest = focused.current ? current.current : undefined;
    if (latest && !latest.secondaryDisabled) latest.onBack();
  }, []);
  return { ...actions, onApply, onBack };
}
