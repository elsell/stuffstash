import { Text } from 'react-native';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { OnboardingCommand } from '../../application/onboarding/OnboardingCommand';
import { onboardingFakes, onboardingServer } from '../../application/onboarding/OnboardingTestSupport';
import { useAppFeedback, type AppFeedbackContextValue } from '../feedback/AppFeedback';
import { AppServicesFeedbackGate, type AppServicesGateController, type AppServicesGateComposition, type AppServicesGateRuntime } from './AppServicesFeedbackGate';

it.each(['sign-out', 'server-change', 'expiry'] as const)('invalidates notices through real %s transitions without restarting initialization', async transition => {
  const h = new MobileRenderHarness(); const fakes = onboardingFakes();
  const profile = { apiBaseUrl: onboardingServer, tenantId: 'tenant' };
  fakes.profiles.profile = profile; fakes.auth.signedIn = true;
  fakes.api.tenants = [{ id: 'tenant', name: 'Home', canCreateInventory: true }];
  fakes.api.inventories = [{ id: 'inventory', tenantId: 'tenant' }];
  const command = new OnboardingCommand(fakes.profiles, () => fakes.api, fakes.auth);
  let starts = 0; let builds = 0; let disconnects = 0; let expired: (() => void) | undefined;
  let gate!: AppServicesGateController<AppServicesGateComposition>; let feedback!: AppFeedbackContextValue; let actions = 0;
  const runtime: AppServicesGateRuntime<AppServicesGateComposition> = {
    onboarding: {
      getStartState: async () => { starts++; return command.getStartState(); },
      expireSession: input => command.expireSession(input), reset: () => command.reset()
    },
    profiles: fakes.profiles,
    createComposition: (_profile, onAuthenticationRequired) => {
      expired = onAuthenticationRequired;
      return { serviceScopeId: `services-${++builds}`, disposePerformance: () => undefined,
        pushSession: { disconnect: async action => { disconnects++; await action(); } } };
    }
  };
  function Probe({ controller }: { controller: AppServicesGateController<AppServicesGateComposition> }) {
    gate = controller; feedback = useAppFeedback();
    return <Text>{controller.state.status}</Text>;
  }
  try {
    await h.render(<AppServicesFeedbackGate runtime={runtime}>{controller => <Probe controller={controller} />}</AppServicesFeedbackGate>);
    await h.settle(); expect(gate.state.status).toBe('ready');
    expect(starts).toBe(1); expect(builds).toBe(1);
    const oldPublisher = feedback;
    await h.run(() => feedback.showNotice({ tone: 'success', title: 'Private item saved', action: { label: 'Undo old edit', onPress: () => { actions++; } } }));
    const oldAction = h.byLabel('Undo old edit');
    expect(oldAction).toBeDefined();
    if (transition === 'sign-out') await h.run(() => gate.signOut());
    else if (transition === 'server-change') await h.run(() => gate.changeServer());
    else { await h.run(() => expired?.()); await h.settle(); }
    expect(gate.state.status).toBe('onboarding');
    expect(h.byText('Private item saved')).toBeUndefined();
    await h.press(oldAction); expect(actions).toBe(0);
    await h.run(() => oldPublisher.showNotice({ tone: 'error', title: 'Late previous failure' }));
    expect(h.byText('Late previous failure')).toBeUndefined();
    expect(starts).toBe(1); expect(builds).toBe(1);
    expect(disconnects).toBe(transition === 'expiry' ? 0 : 1);
    if (gate.state.status === 'onboarding') expect(gate.state.onboardingState.step).toBe(transition === 'server-change' ? 'instance' : 'signIn');
    await h.run(() => gate.complete(profile));
    expect(gate.state.status).toBe('ready'); expect(starts).toBe(1); expect(builds).toBe(2);
    await h.run(() => feedback.showNotice({ tone: 'success', title: 'New session item', action: { label: 'View new item', onPress: () => { actions++; } } }));
    await h.press(h.byLabel('View new item')); expect(actions).toBe(1);
    await h.settle(); expect(starts).toBe(1); expect(builds).toBe(2);
  } finally { await h.unmount(); }
});


it.each(['sign-out', 'server-change'] as const)('keeps the current session and notice when push cleanup prevents %s', async transition => {
  const h = new MobileRenderHarness();
  const fakes = onboardingFakes();
  const profile = { apiBaseUrl: onboardingServer, tenantId: 'tenant' };
  fakes.profiles.profile = profile;
  fakes.auth.signedIn = true;
  fakes.api.tenants = [{ id: 'tenant', name: 'Home', canCreateInventory: true }];
  fakes.api.inventories = [{ id: 'inventory', tenantId: 'tenant' }];
  const command = new OnboardingCommand(fakes.profiles, () => fakes.api, fakes.auth);
  let disposed = false;
  let gate!: AppServicesGateController<AppServicesGateComposition>;
  let feedback!: AppFeedbackContextValue;
  let actions = 0;
  const runtime: AppServicesGateRuntime<AppServicesGateComposition> = {
    onboarding: command, profiles: fakes.profiles,
    createComposition: () => ({
      serviceScopeId: 'current-session',
      disposePerformance: () => { disposed = true; },
      pushSession: { disconnect: async () => { throw new Error('Push cleanup unavailable'); } }
    })
  };
  function Probe({ controller }: { controller: AppServicesGateController<AppServicesGateComposition> }) {
    gate = controller;
    feedback = useAppFeedback();
    return <Text>{controller.state.status}</Text>;
  }
  try {
    await h.render(<AppServicesFeedbackGate runtime={runtime}>{controller => <Probe controller={controller} />}</AppServicesFeedbackGate>);
    await h.settle();
    expect(gate.state.status).toBe('ready');
    await h.run(() => feedback.showNotice({ tone: 'success', title: 'Current session item', action: { label: 'View current item', onPress: () => { actions++; } } }));
    await h.run(async () => {
      await expect(transition === 'sign-out' ? gate.signOut() : gate.changeServer()).rejects.toThrow('Push cleanup unavailable');
    });
    expect(gate.state.status).toBe('ready');
    expect(disposed).toBe(false);
    expect(fakes.profiles.profile).toEqual(profile);
    expect(fakes.auth.signedIn).toBe(true);
    expect(h.byText('Current session item')).toBeDefined();
    await h.press(h.byLabel('View current item'));
    expect(actions).toBe(1);
  } finally { await h.unmount(); }
});
