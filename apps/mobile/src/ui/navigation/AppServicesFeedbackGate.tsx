import { type ReactNode, useCallback, useEffect, useRef, useState } from 'react';
import type { ConnectionProfile, ConnectionProfileStore } from '../../application/onboarding/ConnectionProfile';
import type { OnboardingCommand, OnboardingStartState } from '../../application/onboarding/OnboardingCommand';
import { AppFeedbackProvider, useAppFeedback } from '../feedback/AppFeedback';
import {
  appServicesStateAfterAuthenticationRequired, appServicesStateAfterServerChange,
  appServicesStateAfterSignOut, appServicesStateAfterStartupError,
  appServicesStateFromOnboardingStart, type AppServicesGateState
} from './AppServicesGate';

export type AppServicesGateComposition = {
  readonly serviceScopeId: string;
  readonly disposePerformance: () => void;
  readonly pushSession: { disconnect(action: () => Promise<void>): Promise<void> };
};

export type AppServicesGateRuntime<C extends AppServicesGateComposition> = {
  readonly onboarding: Pick<OnboardingCommand, 'getStartState' | 'expireSession' | 'reset'>;
  readonly profiles: Pick<ConnectionProfileStore, 'load'>;
  readonly createComposition: (profile: ConnectionProfile, onAuthenticationRequired: () => void) => C;
};

export type AppServicesGateController<C extends AppServicesGateComposition> = {
  readonly state: AppServicesGateState<C>;
  readonly complete: (profile: ConnectionProfile) => void;
  readonly setOnboardingState: (state: OnboardingStartState) => void;
  readonly signOut: () => Promise<void>;
  readonly changeServer: () => Promise<void>;
};

type GateProps<C extends AppServicesGateComposition> = {
  readonly runtime: AppServicesGateRuntime<C>;
  readonly children: (controller: AppServicesGateController<C>) => ReactNode;
};

export function AppServicesFeedbackGate<C extends AppServicesGateComposition>({ runtime, children }: GateProps<C>) {
  const [state, setState] = useState<AppServicesGateState<C>>({ status: 'loading' });
  const feedbackScope = state.status === 'ready' ? state.composition.serviceScopeId : 'disconnected';
  return <AppFeedbackProvider scopeKey={feedbackScope}>
    <ServicesController runtime={runtime} state={state} setState={setState}>{children}</ServicesController>
  </AppFeedbackProvider>;
}

function ServicesController<C extends AppServicesGateComposition>({ runtime, state, setState, children }: GateProps<C> & {
  readonly state: AppServicesGateState<C>;
  readonly setState: (state: AppServicesGateState<C>) => void;
}) {
  const { showDialog } = useAppFeedback();
  const authPromptVisibleRef = useRef(false);

  const buildComposition = useCallback((profile: ConnectionProfile) => runtime.createComposition(profile, () => {
    if (authPromptVisibleRef.current) {
      return;
    }

    authPromptVisibleRef.current = true;
    runtime.onboarding
      .expireSession({ profile })
      .then((onboardingState) => {
        setState(appServicesStateAfterAuthenticationRequired(onboardingState.profile ?? profile));
        showDialog({
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
  }), [showDialog, runtime]);

  useEffect(() => {
    let isCurrent = true;

    runtime.onboarding
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
  }, [buildComposition, runtime.onboarding]);

  const signOut = async (): Promise<void> => {
    if (state.status !== 'ready') return;
    const composition = state.composition;
    await composition.pushSession.disconnect(async () => {
      composition.disposePerformance();
      const profile = await runtime.profiles.load();
      if (!profile) {
        await runtime.onboarding.reset();
        setState(appServicesStateAfterServerChange());
        return;
      }

      await runtime.onboarding.expireSession({ profile });
      setState(appServicesStateAfterSignOut(profile));
    });
  };
  const changeServer = async (): Promise<void> => {
    if (state.status !== 'ready') return;
    const composition = state.composition;
    await composition.pushSession.disconnect(async () => {
      composition.disposePerformance();
      await runtime.onboarding.reset();
      setState(appServicesStateAfterServerChange());
    });
  };

  return children({
    state, signOut, changeServer,
    complete: profile => {
      authPromptVisibleRef.current = false;
      setState({ status: 'ready', composition: buildComposition(profile) });
    },
    setOnboardingState: onboardingState => setState({ status: 'onboarding', onboardingState })
  });
}
