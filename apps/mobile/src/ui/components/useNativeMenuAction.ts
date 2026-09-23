import { useCallback, useLayoutEffect, useRef } from 'react';
import type { NativeActionMenuGroup } from './NativeActionMenu.types';

/** Native popups can retain callbacks across React commits and removal. */
export function useNativeMenuAction(groups: readonly NativeActionMenuGroup[], disabled: boolean) {
  const committed = useRef<{ groups: readonly NativeActionMenuGroup[]; disabled: boolean } | null>(null);
  useLayoutEffect(() => {
    committed.current = { groups, disabled };
    return () => { committed.current = null; };
  }, [groups, disabled]);

  const pressItem = useCallback((groupId: string, itemId: string, beforePress?: () => void) => {
    const owner = committed.current;
    if (!owner || owner.disabled) return;
    const item = owner.groups.find(group => group.id === groupId)?.items.find(candidate => candidate.id === itemId);
    if (!item || item.disabled) return;
    beforePress?.();
    item.onPress();
  }, []);
  const trigger = useCallback((open: () => void) => {
    if (committed.current && !committed.current.disabled) open();
  }, []);
  return { pressItem, trigger };
}
