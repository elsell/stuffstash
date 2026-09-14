import { useEffect, useRef, useState } from 'react';

/** Keep asynchronous scope verification within the lifetime of its filter sheet. */
export function useBrowseFilterNavigation(scopeKey: string, verify: (signal: AbortSignal) => Promise<void>) {
  const request = useRef<AbortController | undefined>(undefined);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const abort = () => { request.current?.abort(); request.current = undefined; };
  const cancel = () => { abort(); setBusy(false); };
  useEffect(() => {
    setBusy(false); setError('');
    return abort;
  }, [scopeKey]);
  async function navigate(action: () => void) {
    if (request.current) return;
    const controller = new AbortController();
    request.current = controller;
    setBusy(true); setError('');
    try {
      await verify(controller.signal);
      if (!controller.signal.aborted) action();
    } catch {
      if (!controller.signal.aborted) setError('Inventory changed or could not be verified. Reopen Browse filters.');
    } finally {
      if (request.current === controller) {
        request.current = undefined;
        setBusy(false);
      }
    }
  }
  return { busy, error, navigate, cancel };
}
