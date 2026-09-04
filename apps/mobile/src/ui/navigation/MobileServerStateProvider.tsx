import { focusManager, QueryClientProvider, type QueryClient } from '@tanstack/react-query';
import { createContext, type ReactNode, useContext, useEffect } from 'react';
import { AppState } from 'react-native';
import { disposeMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import type { CurrentInventoryScope } from '../../application/home/CurrentInventoryScopeQuery';
import type { ReadRequest } from '../../application/shared/ReadRequest';

type MobileServerStateProviderProps = {
  readonly children: ReactNode;
  readonly client: QueryClient;
  readonly scopeId: string;
  readonly loadInventoryScope: (request?: ReadRequest) => Promise<CurrentInventoryScope>;
};

export type MobileServerStateScope = {
  readonly scopeId: string;
  readonly loadInventoryScope: (request?: ReadRequest) => Promise<CurrentInventoryScope>;
};

const MobileServerStateScopeContext = createContext<MobileServerStateScope | undefined>(undefined);

export function MobileServerStateProvider({ children, client, loadInventoryScope, scopeId }: MobileServerStateProviderProps) {
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

  return (
    <QueryClientProvider client={client}>
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
