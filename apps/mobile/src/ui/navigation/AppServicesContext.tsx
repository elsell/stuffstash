import { createContext, ReactNode, useContext, useEffect, useMemo, useState } from 'react';
import { ActivityIndicator, StyleSheet, Text, View } from 'react-native';
import {
  createMobileComposition,
  getConnectionProfileStore,
  createOnboardingCommand,
  createSeedConnectionProfile,
  MobileComposition
} from '../../bootstrap/mobileComposition';
import { OnboardingScreen } from '../screens/OnboardingScreen';
import { colors, spacing } from '../theme/tokens';
import {
  appServicesStateAfterReset,
  appServicesStateAfterStartupError,
  appServicesStateFromOnboardingStart,
  AppServicesGateState
} from './AppServicesGate';
import { VoiceInteractionStateProvider } from './VoiceInteractionStateContext';

const AppServicesContext = createContext<MobileComposition | null>(null);
const AppConnectionActionsContext = createContext<{ readonly resetConnectionProfile: () => Promise<void> } | null>(null);

type AppServicesProviderProps = {
  readonly children: ReactNode;
};

export function AppServicesProvider({ children }: AppServicesProviderProps) {
  const onboardingCommand = useMemo(() => createOnboardingCommand(), []);
  const [state, setState] = useState<AppServicesState>({ status: 'loading' });

  useEffect(() => {
    let isCurrent = true;

    onboardingCommand
      .getStartState()
      .then((startState) => {
        if (!isCurrent) {
          return;
        }
        setState(appServicesStateFromOnboardingStart(startState, createMobileComposition));
      })
      .catch(() => {
        if (isCurrent) {
          setState(appServicesStateAfterStartupError());
        }
      });

    return () => {
      isCurrent = false;
    };
  }, [onboardingCommand]);

  if (state.status === 'loading') {
    return <LoadingAppState />;
  }

  if (state.status === 'onboarding') {
    return (
      <OnboardingScreen
        command={onboardingCommand}
        initialApiBaseUrl={createSeedConnectionProfile()?.apiBaseUrl}
        initialState={state.onboardingState}
        onComplete={(profile) => {
          setState({ status: 'ready', composition: createMobileComposition(profile) });
        }}
        onStateChange={(onboardingState) => {
          setState({ status: 'onboarding', onboardingState });
        }}
      />
    );
  }

  const mobileComposition = state.composition;
  return (
    <AppServicesContext.Provider value={mobileComposition}>
      <AppConnectionActionsContext.Provider
        value={{
          resetConnectionProfile: async () => {
            await getConnectionProfileStore().clear();
            setState(appServicesStateAfterReset());
          }
        }}
      >
        <VoiceInteractionStateProvider
          diagnosticsEnabled={mobileComposition.voiceDeveloperDiagnosticsEnabled}
          previewQuery={mobileComposition.voiceInteractionPreviewQuery}
          realtimeController={mobileComposition.realtimeVoiceSessionController}
        >
          {children}
        </VoiceInteractionStateProvider>
      </AppConnectionActionsContext.Provider>
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

export function useAppConnectionActions(): { readonly resetConnectionProfile: () => Promise<void> } {
  const actions = useContext(AppConnectionActionsContext);
  if (actions === null) {
    throw new Error('App connection actions are not available.');
  }

  return actions;
}

type AppServicesState = AppServicesGateState;

function LoadingAppState() {
  return (
    <View style={styles.loading}>
      <ActivityIndicator color={colors.accent} />
      <Text style={styles.loadingText}>Loading Stuff Stash</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  loading: {
    alignItems: 'center',
    backgroundColor: colors.background,
    flex: 1,
    justifyContent: 'center',
    padding: spacing.lg
  },
  loadingText: {
    color: colors.textMuted,
    fontSize: 16,
    fontWeight: '700',
    marginTop: spacing.md
  }
});
