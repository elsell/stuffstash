import { useEffect, useSyncExternalStore } from 'react';
import { getScreenFocused, subscribeScreenFocus } from './navigation';
export function useIsFocused() { return useSyncExternalStore(subscribeScreenFocus, getScreenFocused); }
import { installPreventRemove } from './navigation';

export function usePreventRemove(active: boolean, callback: (event: { data: { action: Record<string, unknown> } }) => void) {
  useEffect(() => installPreventRemove(active, callback), [active, callback]);
}
