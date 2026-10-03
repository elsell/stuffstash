import { useCallback, useEffect, useRef, useState } from 'react';
import { AppState } from 'react-native';
import { useFocusEffect } from 'expo-router';
export interface PrintingPolling<T> {
  readonly intervalMilliseconds: number;
  readonly maximumDelayMilliseconds: number;
  readonly shouldContinue: (value: T) => boolean;
}
/** Own cancellation and suppress stale results whenever this native task loses focus. */
export function usePrintingTask<T>(load: (signal: AbortSignal) => Promise<T>, identity: string, polling?: PrintingPolling<T>) {
  const [data, setData] = useState<T>(); const [error, setError] = useState(false); const [loading, setLoading] = useState(true);
  const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  useEffect(() => { if (data !== undefined && polling && !polling.shouldContinue(data) && timer.current) { clearTimeout(timer.current); timer.current = undefined; } }, [data, polling]);
  const [revision, setRevision] = useState(0); const lifetime = useRef<AbortController | undefined>(undefined);
  const [foreground, setForeground] = useState(AppState.currentState === 'active');
  useEffect(() => { const listener = AppState.addEventListener('change', state => setForeground(state === 'active')); return () => listener.remove(); }, []);
  useFocusEffect(useCallback(() => {
    if (!foreground) { setData(undefined); return; }
    const controller = new AbortController(); lifetime.current = controller;
    let delay = polling?.intervalMilliseconds ?? 0;
    setData(undefined); setError(false); setLoading(true);
    const refresh = async () => {
      let continuePolling = Boolean(polling);
      try {
        const value = await load(controller.signal);
        if (!controller.signal.aborted) { setData(value); setError(false); }
        continuePolling = Boolean(polling?.shouldContinue(value));
        delay = polling?.intervalMilliseconds ?? 0;
      }
      catch { delay = Math.min(delay * 2, polling?.maximumDelayMilliseconds ?? 0); if (!controller.signal.aborted) { setData(undefined); setError(true); } }
      finally {
        if (!controller.signal.aborted) { setLoading(false); if (continuePolling) timer.current = setTimeout(() => void refresh(), delay); }
      }
    };
    void refresh();
    return () => { controller.abort(); if (timer.current) clearTimeout(timer.current); timer.current = undefined; setData(undefined); };
  }, [load, identity, revision, polling, foreground]));
  return { data, setData, error, loading, lifetime, reload: () => setRevision(value => value + 1) };
}
