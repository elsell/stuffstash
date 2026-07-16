import { useEffect } from 'react';
import { installPreventRemove } from './navigation';

export function usePreventRemove(active: boolean, callback: (event: { data: { action: Record<string, unknown> } }) => void) {
  useEffect(() => installPreventRemove(active, callback), [active, callback]);
}
