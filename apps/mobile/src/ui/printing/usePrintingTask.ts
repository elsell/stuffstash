import { useCallback, useEffect, useRef, useState } from 'react';
import { AppState } from 'react-native';
import { useFocusEffect } from 'expo-router';
/** Own cancellation and suppress stale results whenever this native task loses focus. */
export function usePrintingTask<T>(load: (signal: AbortSignal) => Promise<T>, identity: string, pollMilliseconds?: number) {
  const [data, setData] = useState<T>(); const [error, setError] = useState(false); const [loading, setLoading] = useState(true);
  const [revision, setRevision] = useState(0); const lifetime = useRef<AbortController | undefined>(undefined);
  const [foreground, setForeground] = useState(AppState.currentState === 'active');
  useEffect(() => { const listener = AppState.addEventListener('change', state => setForeground(state === 'active')); return () => listener.remove(); }, []);
  useFocusEffect(useCallback(() => {
    if (!foreground) { setData(undefined); return; }
    const controller = new AbortController(); lifetime.current = controller;
    let timer: ReturnType<typeof setTimeout> | undefined;
    setData(undefined); setError(false); setLoading(true);
    const refresh = async () => {
      try { const value = await load(controller.signal); if (!controller.signal.aborted) { setData(value); setError(false); } }
      catch { if (!controller.signal.aborted) { setData(undefined); setError(true); } }
      finally {
        if (!controller.signal.aborted) { setLoading(false); if (pollMilliseconds) timer = setTimeout(() => void refresh(), pollMilliseconds); }
      }
    };
    void refresh();
    return () => { controller.abort(); if (timer) clearTimeout(timer); setData(undefined); };
  }, [load, identity, revision, pollMilliseconds, foreground]));
  return { data, setData, error, loading, lifetime, reload: () => setRevision(value => value + 1) };
}
