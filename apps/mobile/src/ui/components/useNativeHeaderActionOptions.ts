import { useLayoutEffect, useMemo, useRef } from 'react';
import { nativeHeaderActionOptions } from './NativeHeaderActions';
import type { HeaderOptions, NativeHeaderAction } from './NativeHeaderActions.types';

/** Keep navigation presentation stable; action kinds are unique within each side. */
export function useNativeHeaderActionOptions(actions: readonly NativeHeaderAction[], position: 'left' | 'right' = 'right'): HeaderOptions {
  const current = useRef(actions);
  useLayoutEffect(() => {
    current.current = actions;
    return () => { current.current = []; };
  }, [actions]);
  const presentation = actions.map(({ onPress: _onPress, ...attributes }) => attributes);
  // Function identity is intentionally excluded: updating a closure must not
  // recursively update the navigation context that caused its render.
  const signature = JSON.stringify(presentation);
  return useMemo(() => nativeHeaderActionOptions(presentation.map(attributes => ({
    ...attributes,
    onPress: () => {
      const action = current.current.find(candidate => candidate.kind === attributes.kind);
      if (action && !action.disabled) action.onPress();
    }
  })), position), [signature, position]);
}
