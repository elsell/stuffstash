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
  readonly readyNoticePlacement?: 'root' | 'screen';
  readonly runtime: AppServicesGateRuntime<C>;
  readonly children: (controller: AppServicesGateController<C>) => ReactNode;
};

export function AppServicesFeedbackGate<C extends AppServicesGateComposition>({ runtime, children, readyNoticePlacement = 'root' }: GateProps<C>) {
  const [state, setState] = useState<AppServicesGateState<C>>({ status: 'loading' });
  const feedbackScope = state.status === 'ready' ? state.composition.serviceScopeId : 'disconnected';
  return <AppFeedbackProvider scopeKey={feedbackScope} noticePlacement={state.status === 'ready' ? readyNoticePlacement : 'root'}>
    <ServicesController runtime={runtime} state={state} setState={setState}>{children}</ServicesController>
  </AppFeedbackProvider>;
}

function ServicesController<C extends AppServicesGateComposition>({ runtime, state, setState, children }: GateProps<C> & {
  readonly state: AppServicesGateState<C>;
  readonly setState: (state: AppServicesGateState<C>) => void;
}) {
  const { showDialog } = useAppFeedback();
  const authPromptVisibleRef = useRef(false);

  const compositionVisit = useRef<object | undefined>(undefined);
  const retireComposition = useCallback((visit: object | undefined) => {
    if (!visit || compositionVisit.current !== visit) return false;
    compositionVisit.current = undefined;
    authPromptVisibleRef.current = false;
    return true;
  }, []);
  const buildComposition = useCallback((profile: ConnectionProfile) => {
    const visit = {};
    compositionVisit.current = visit;
    const isCurrent = () => compositionVisit.current === visit;
    return runtime.createComposition(profile, () => {
      if (!isCurrent() || authPromptVisibleRef.current) return;
      authPromptVisibleRef.current = true;
      runtime.onboarding.expireSession({ profile }).then(onboardingState => {
        if (!retireComposition(visit)) return;
        setState(appServicesStateAfterAuthenticationRequired(onboardingState.profile ?? profile));
        showDialog({
          title: 'Session expired',
          message: 'Please sign in again to continue using Stuff Stash.',
          primaryAction: { label: 'Continue' }
        });
      }).catch(() => {
        if (!retireComposition(visit)) return;
        setState(appServicesStateAfterAuthenticationRequired(profile));
      });
    });
  }, [showDialog, runtime, retireComposition]);

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
      compositionVisit.current = undefined;
    };
  }, [buildComposition, runtime.onboarding]);

  const signOut = async (): Promise<void> => {
    if (state.status !== 'ready') return;
    const composition = state.composition;
    const visit = compositionVisit.current;
    await composition.pushSession.disconnect(async () => {
      composition.disposePerformance();
      const profile = await runtime.profiles.load();
      if (!profile) {
        await runtime.onboarding.reset();
        if (retireComposition(visit)) setState(appServicesStateAfterServerChange());
        return;
      }

      await runtime.onboarding.expireSession({ profile });
      if (retireComposition(visit)) setState(appServicesStateAfterSignOut(profile));
    });
  };
  const changeServer = async (): Promise<void> => {
    if (state.status !== 'ready') return;
    const composition = state.composition;
    const visit = compositionVisit.current;
    await composition.pushSession.disconnect(async () => {
      composition.disposePerformance();
      await runtime.onboarding.reset();
      if (retireComposition(visit)) setState(appServicesStateAfterServerChange());
    });
  };

  return children({
    state, signOut, changeServer,
    complete: profile => {
      authPromptVisibleRef.current = false;
      setState({ status: 'ready', composition: buildComposition(profile) });
    },
    setOnboardingState: onboardingState => {
      retireComposition(compositionVisit.current);
      setState({ status: 'onboarding', onboardingState });
    }
  });
}
