import { useCallback, useRef, useState } from 'react';
import { useFocusEffect } from 'expo-router';
import { useTaskPresentation } from '../navigation/useTaskPresentation';

/** Keep asynchronous scope verification within the lifetime of its filter sheet. */
export function useBrowseFilterNavigation(scopeKey: string, verify: (signal: AbortSignal) => Promise<void>) {
  const request = useRef<AbortController | undefined>(undefined);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const presentation = useTaskPresentation(undefined, scopeKey);
  const cancel = useCallback(() => { request.current?.abort(); request.current = undefined; setBusy(false); }, []);
  useFocusEffect(useCallback(() => {
    setBusy(false); setError('');
    return cancel;
  }, [scopeKey, cancel]));
  async function navigate(action: () => void) {
    const ownsPresentation = presentation();
    if (request.current || !ownsPresentation()) return;
    const controller = new AbortController();
    request.current = controller;
    setBusy(true); setError('');
    try {
      await verify(controller.signal);
      if (!controller.signal.aborted && ownsPresentation()) action();
    } catch {
      if (!controller.signal.aborted && ownsPresentation()) setError('Inventory changed or could not be verified. Reopen Browse filters.');
    } finally {
      if (request.current === controller) {
        request.current = undefined;
        setBusy(false);
      }
    }
  }
  return { busy, error, navigate, cancel };
}
