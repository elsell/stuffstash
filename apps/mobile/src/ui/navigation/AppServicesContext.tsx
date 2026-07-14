import { createContext, ReactNode, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react';
import { ActivityIndicator, StyleSheet, Text, View } from 'react-native';
import {
  createMobileComposition,
  getConnectionProfileStore,
  createOnboardingCommand,
  createSeedConnectionProfile,
  MobileComposition
} from '../../bootstrap/mobileComposition';
import type { ConnectionProfile } from '../../application/onboarding/ConnectionProfile';
import { AppFeedbackProvider, useAppFeedback } from '../feedback/AppFeedback';
import { OnboardingScreen } from '../screens/OnboardingScreen';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { spacing, type MobileColorPalette } from '../theme/tokens';
import {
  appServicesStateAfterAuthenticationRequired,
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
  return (
    <AppFeedbackProvider>
      <AppServicesProviderInner>{children}</AppServicesProviderInner>
    </AppFeedbackProvider>
  );
}

function AppServicesProviderInner({ children }: AppServicesProviderProps) {
  const onboardingCommand = useMemo(() => createOnboardingCommand(), []);
  const [state, setState] = useState<AppServicesState>({ status: 'loading' });
  const feedback = useAppFeedback();
  const authPromptVisibleRef = useRef(false);

  const buildComposition = useCallback((profile: ConnectionProfile) => createMobileComposition(profile, {
    onAuthenticationRequired: () => {
      if (authPromptVisibleRef.current) {
        return;
      }

      authPromptVisibleRef.current = true;
      onboardingCommand
        .expireSession({ profile })
        .then((onboardingState) => {
          setState(appServicesStateAfterAuthenticationRequired(onboardingState.profile ?? profile));
          feedback.showDialog({
            title: 'Session expired',
          message: 'Please sign in again to continue using Stuff Stash.',
          primaryAction: {
              label: 'Continue',
              onPress: () => {
                authPromptVisibleRef.current = false;
              }
            }
          });
        })
        .catch(() => {
          authPromptVisibleRef.current = false;
          setState(appServicesStateAfterAuthenticationRequired(profile));
        });
    }
  }), [feedback, onboardingCommand]);

  useEffect(() => {
    let isCurrent = true;

    onboardingCommand
      .getStartState()
      .then((startState) => {
        if (!isCurrent) {
          return;
        }
        setState(appServicesStateFromOnboardingStart(startState, buildComposition));
      })
      .catch(() => {
        if (isCurrent) {
          setState(appServicesStateAfterStartupError());
        }
      });

    return () => {
      isCurrent = false;
    };
  }, [buildComposition, onboardingCommand]);

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
          authPromptVisibleRef.current = false;
          setState({ status: 'ready', composition: buildComposition(profile) });
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
            await onboardingCommand.reset();
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
  const palette = useAppearancePalette();
  const styles = createStyles(palette);
  return (
    <View style={styles.loading}>
      <ActivityIndicator color={palette.accent} />
      <Text style={styles.loadingText}>Loading Stuff Stash</Text>
    </View>
  );
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
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
}
