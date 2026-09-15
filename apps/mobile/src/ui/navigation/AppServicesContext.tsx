import { PushRegistrationLifecycle } from './PushRegistrationLifecycle';
import { useRouter } from 'expo-router';
import { createContext, ReactNode, useContext, useMemo } from 'react';
import { ActivityIndicator, StyleSheet, Text, View } from 'react-native';
import {
  createMobileComposition,
  getConnectionProfileStore,
  createOnboardingCommand,
  createSeedConnectionProfile,
  MobileComposition
} from '../../bootstrap/mobileComposition';
import type { ConnectionProfile } from '../../application/onboarding/ConnectionProfile';
import { OnboardingScreen } from '../screens/OnboardingScreen';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { spacing, type MobileColorPalette } from '../theme/tokens';
import { AppServicesFeedbackGate, type AppServicesGateController } from './AppServicesFeedbackGate';
import { VoiceInteractionStateProvider } from './VoiceInteractionStateContext';
import { useInventoryInvitationLink } from './InventoryInvitationLinkContext';
import { MobileServerStateProvider } from './MobileServerStateProvider';

const AppServicesContext = createContext<MobileComposition | null>(null);

export type AppConnectionActions = {
  readonly signOut: () => Promise<void>;
  readonly changeServer: () => Promise<void>;
};

const AppConnectionActionsContext = createContext<AppConnectionActions | null>(null);

type AppServicesProviderProps = {
  readonly children: ReactNode;
};

export function AppServicesProvider({ children }: AppServicesProviderProps) {
  const onboardingCommand = useMemo(() => createOnboardingCommand(), []);
  const runtime = useMemo(() => ({
    onboarding: onboardingCommand,
    profiles: getConnectionProfileStore(),
    createComposition: (profile: ConnectionProfile, onAuthenticationRequired: () => void) =>
      createMobileComposition(profile, { onAuthenticationRequired })
  }), [onboardingCommand]);
  return (
    <AppServicesFeedbackGate runtime={runtime} readyNoticePlacement="screen">
      {controller => <AppServicesContent controller={controller} onboardingCommand={onboardingCommand}>{children}</AppServicesContent>}
    </AppServicesFeedbackGate>
  );
}

function AppServicesContent({ children, controller, onboardingCommand }: AppServicesProviderProps & {
  readonly controller: AppServicesGateController<MobileComposition>;
  readonly onboardingCommand: ReturnType<typeof createOnboardingCommand>;
}) {
  const invitationLink = useInventoryInvitationLink();
  const router = useRouter();
  const { state, signOut, changeServer } = controller;
  if (state.status === 'loading') {
    return <LoadingAppState />;
  }

  if (state.status === 'onboarding') {
    return (
      <OnboardingScreen
        command={onboardingCommand}
        initialApiBaseUrl={createSeedConnectionProfile()?.apiBaseUrl}
        initialState={state.onboardingState}
        invitationPending={Boolean(invitationLink.reference)}
        onStartOver={() => { invitationLink.clear(); router.replace('/'); }}
        onComplete={controller.complete}
        onStateChange={controller.setOnboardingState}
      />
    );
  }

  const mobileComposition = state.composition;

  return (
    <MobileServerStateProvider
      client={mobileComposition.queryClient}
      acquirePerformance={mobileComposition.acquirePerformance}
      connectivitySource={mobileComposition.connectivitySource}
      loadInventoryScope={(request) => mobileComposition.currentInventoryScopeQuery.execute(request)}
      scopeId={mobileComposition.serviceScopeId}
    >
      <AppServicesContext.Provider value={mobileComposition}>
        <AppConnectionActionsContext.Provider
          value={{
            signOut,
            changeServer
          }}
        >
          <VoiceInteractionStateProvider
            diagnosticsEnabled={mobileComposition.voiceDeveloperDiagnosticsEnabled}
            previewQuery={mobileComposition.voiceInteractionPreviewQuery}
            realtimeController={mobileComposition.realtimeVoiceSessionController}
          >
            <PushRegistrationLifecycle />
            {children}
          </VoiceInteractionStateProvider>
        </AppConnectionActionsContext.Provider>
      </AppServicesContext.Provider>
    </MobileServerStateProvider>
  );
}

export function useAppServices(): MobileComposition {
  const services = useContext(AppServicesContext);
  if (services === null) {
    throw new Error('App services are not available.');
  }

  return services;
}

export function useAppConnectionActions(): AppConnectionActions {
  const actions = useContext(AppConnectionActionsContext);
  if (actions === null) {
    throw new Error('App connection actions are not available.');
  }

  return actions;
}

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
