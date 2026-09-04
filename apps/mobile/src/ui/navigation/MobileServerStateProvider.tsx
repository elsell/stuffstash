import { focusManager, QueryClientProvider, type QueryClient } from '@tanstack/react-query';
import { type ReactNode, useEffect } from 'react';
import { AppState } from 'react-native';
import { disposeMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';

type MobileServerStateProviderProps = {
  readonly children: ReactNode;
  readonly client: QueryClient;
};

export function MobileServerStateProvider({ children, client }: MobileServerStateProviderProps) {
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

  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}
