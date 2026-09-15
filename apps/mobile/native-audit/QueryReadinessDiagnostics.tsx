import { useCallback, useSyncExternalStore } from 'react';
import { focusManager, onlineManager, type QueryClient } from '@tanstack/react-query';
import { Text } from 'react-native';

/** Runner-only synthetic state; never render resource data or error contents. */
export function QueryReadinessDiagnostics({ client }: { readonly client: QueryClient }) {
  const subscribe = useCallback((notify: () => void) => {
    const removeCache = client.getQueryCache().subscribe(notify);
    const removeFocus = focusManager.subscribe(notify);
    const removeOnline = onlineManager.subscribe(notify);
    return () => { removeCache(); removeFocus(); removeOnline(); };
  }, [client]);
  const snapshot = useCallback(() => JSON.stringify({
    online: onlineManager.isOnline(), focused: focusManager.isFocused(),
    queries: client.getQueryCache().getAll().map(query => ({
      key: query.queryKey, status: query.state.status, fetch: query.state.fetchStatus,
      data: query.state.data !== undefined, observers: query.getObserversCount()
    }))
  }), [client]);
  const state = useSyncExternalStore(subscribe, snapshot, snapshot);
  return <Text testID="audit-query-readiness" accessibilityLabel={state} accessibilityValue={{ text: state }}
    pointerEvents="none" numberOfLines={1}
    style={{ position: 'absolute', left: 8, bottom: 4, width: 100, height: 16, fontSize: 10 }}>Audit query state</Text>;
}
