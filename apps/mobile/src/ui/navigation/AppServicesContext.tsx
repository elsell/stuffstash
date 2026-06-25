import { createContext, ReactNode, useContext } from 'react';
import {
  createMobileComposition,
  MobileComposition
} from '../../bootstrap/mobileComposition';
import { VoiceInteractionStateProvider } from './VoiceInteractionStateContext';

const mobileComposition = createMobileComposition();

const AppServicesContext = createContext<MobileComposition | null>(null);

type AppServicesProviderProps = {
  readonly children: ReactNode;
};

export function AppServicesProvider({ children }: AppServicesProviderProps) {
  return (
    <AppServicesContext.Provider value={mobileComposition}>
      <VoiceInteractionStateProvider
        previewQuery={mobileComposition.voiceInteractionPreviewQuery}
        realtimeController={mobileComposition.realtimeVoiceSessionController}
      >
        {children}
      </VoiceInteractionStateProvider>
    </AppServicesContext.Provider>
  );
}

export function useAppServices(): MobileComposition {
  const services = useContext(AppServicesContext);
  if (services === null) {
    throw new Error('App services are not available.');
  }

  return services;
}
