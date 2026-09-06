import type { ConnectivitySource } from '../../application/shared/ConnectivitySource';
import { focusManager, onlineManager, QueryClientProvider, type QueryClient } from '@tanstack/react-query';
import { createContext, type ReactNode, useContext, useEffect } from 'react';
import { AppState } from 'react-native';
import { disposeMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import type { CurrentInventoryScope } from '../../application/home/CurrentInventoryScopeQuery';
import type { ReadRequest } from '../../application/shared/ReadRequest';

type MobileServerStateProviderProps = {
  readonly children: ReactNode;
  readonly connectivitySource?: ConnectivitySource;
  readonly client: QueryClient;
  readonly disposePerformance?: () => void;
  readonly scopeId: string;
  readonly loadInventoryScope: (request?: ReadRequest) => Promise<CurrentInventoryScope>;
};

export type MobileServerStateScope = {
  readonly scopeId: string;
  readonly loadInventoryScope: (request?: ReadRequest) => Promise<CurrentInventoryScope>;
};

const MobileServerStateScopeContext = createContext<MobileServerStateScope | undefined>(undefined);

export function MobileServerStateProvider({ children, client, loadInventoryScope, scopeId, connectivitySource, disposePerformance }: MobileServerStateProviderProps) {
  useEffect(() => () => disposePerformance?.(), [disposePerformance]);

  useEffect(() => {
    focusManager.setFocused(AppState.currentState === 'active');
    const subscription = AppState.addEventListener('change', (state) => {
      focusManager.setFocused(state === 'active');
    });

    return () => {
      subscription.remove();
      focusManager.setFocused(undefined);
      void disposeMobileQueryClient(client);
    };
  }, [client]);

  useEffect(() => {
    const remove = connectivitySource?.subscribe(connected => onlineManager.setOnline(connected));
    return () => { remove?.(); onlineManager.setOnline(true); };
  }, [connectivitySource]);

  return (
    <QueryClientProvider key={scopeId} client={client}>
      <MobileServerStateScopeContext.Provider value={{ scopeId, loadInventoryScope }}>
        {children}
      </MobileServerStateScopeContext.Provider>
    </QueryClientProvider>
  );
}

export function useMobileServerStateScope(): MobileServerStateScope {
  const scope = useContext(MobileServerStateScopeContext);
  if (!scope) {
    throw new Error('Mobile server state is not available.');
  }
  return scope;
}

export function useMobileServerStateScopeId(): string {
  return useMobileServerStateScope().scopeId;
}
