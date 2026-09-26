import { useCallback, useLayoutEffect, useRef } from 'react';

/** Native views may deliver retained callbacks after a new render or teardown. */
export function useCommittedCommand(onPress: () => void, disabled = false) {
  const current = useRef<(() => void) | undefined>(undefined);
  useLayoutEffect(() => {
    current.current = disabled ? undefined : onPress;
    return () => { current.current = undefined; };
  }, [onPress, disabled]);
  return useCallback(() => current.current?.(), []);
}
