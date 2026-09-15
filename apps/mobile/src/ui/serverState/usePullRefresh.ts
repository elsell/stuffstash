import { useCallback, useRef, useState } from 'react';
import { useFocusEffect } from 'expo-router';

/** Native refresh presentation belongs to a pull gesture, not query activity. */
export function usePullRefresh(refreshData: () => Promise<void>) {
  const [refreshing, setRefreshing] = useState(false);
  const sequence = useRef(0);
  const active = useRef<number | undefined>(undefined);
  const focused = useRef(false);
  useFocusEffect(useCallback(() => {
    focused.current = true;
    return () => {
      focused.current = false;
      sequence.current++;
      active.current = undefined;
      setRefreshing(false);
    };
  }, []));
  async function refresh() {
    if (!focused.current || active.current !== undefined) return;
    const request = ++sequence.current;
    active.current = request;
    setRefreshing(true);
    try { await refreshData(); }
    finally {
      if (active.current === request) {
        active.current = undefined;
        setRefreshing(false);
      }
    }
  }
  return { refreshing, refresh };
}
