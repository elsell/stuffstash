import { useState } from 'react';
import { describe, expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { OnboardingCommand, type OnboardingStartState } from '../../application/onboarding/OnboardingCommand';
import { OnboardingAuthFake, onboardingFakes, onboardingServer } from '../../application/onboarding/OnboardingTestSupport';
import { OnboardingScreen } from './OnboardingScreen';

describe('onboarding screen', () => {
  async function fixture() {
    const f = onboardingFakes();
    const command = new OnboardingCommand(f.profiles, () => f.api, f.auth);
    const completed: unknown[] = [];
    const harness = new MobileRenderHarness();
    function Screen() {
      const [state, setState] = useState<OnboardingStartState>({ step: 'instance' });
      return <OnboardingScreen command={command} initialState={state} onStateChange={setState}
        onComplete={profile => completed.push(profile)} />;
    }
    await harness.render(<Screen />);
    return { ...f, harness, completed };
  }
  it('goes directly from connect/sign-in to household fields and clears drafts on start-over', async () => {
    const { harness, auth, profiles } = await fixture();
    expect(harness.byText('Sign in with SSO')).toBeUndefined();
    await harness.changeText(harness.byLabel('Server address'), onboardingServer);
    await harness.press(harness.byLabel('Connect and sign in'));
    expect(auth.signIns).toEqual([onboardingServer]);
    expect(harness.byText('Set up your household')).toBeDefined();
    await harness.changeText(harness.byLabel('Household name'), 'Maple Street');
    await harness.press(harness.byLabel('Sign out and start over'));
    expect(profiles.profile).toBeUndefined();
    expect(harness.byLabel('Server address')?.props.defaultValue).toBe('');
    await harness.changeText(harness.byLabel('Server address'), onboardingServer);
    await harness.press(harness.byLabel('Connect and sign in'));
    expect(harness.byLabel('Household name')?.props.defaultValue).toBe('');
    expect(harness.byLabel('First inventory')?.props.defaultValue).toBe('Home Inventory');
    await harness.unmount();
  });
  it('shows field-specific errors and preserves the form for retry', async () => {
    const { harness, api } = await fixture();
    await harness.changeText(harness.byLabel('Server address'), onboardingServer);
    await harness.press(harness.byLabel('Connect and sign in'));
    await harness.press(harness.byLabel('Create household'));
    expect(harness.byLabel('Create household')?.props.accessibilityState.disabled).toBe(true);
    expect(harness.byText('Enter a household name to continue.')).toBeDefined();
    expect(api.tenantWrites).toBe(0);
    await harness.changeText(harness.byLabel('Household name'), 'Maple Street');
    await harness.press(harness.byLabel('Create household'));
    expect(api.inventoryWrites).toBe(1);
    await harness.unmount();
  });
  it('shows inventory-only recovery after household creation and keeps the entered inventory name', async () => {
    const { harness, api } = await fixture();
    await harness.changeText(harness.byLabel('Server address'), onboardingServer);
    await harness.press(harness.byLabel('Connect and sign in'));
    await harness.changeText(harness.byLabel('Household name'), 'Maple Street');
    await harness.changeText(harness.byLabel('First inventory'), 'Workshop');
    api.failInventoryBeforeWrite = true;
    await harness.press(harness.byLabel('Create household'));
    expect(harness.byText('Create your first inventory')).toBeDefined();
    expect(harness.byLabel('Household name')).toBeUndefined();
    expect(harness.byLabel('Inventory name')?.props.defaultValue).toBe('Workshop');
    api.failInventoryBeforeWrite = false;
    await harness.press(harness.byLabel('Create inventory'));
    expect(api.tenantWrites).toBe(1);
    expect(api.inventoryWrites).toBe(1);
    await harness.unmount();
  });

  it('does not navigate after a pending start-over finishes on an unmounted screen', async () => {
    let finish!: () => void;
    const gate = new Promise<void>(resolve => { finish = resolve; });
    class DelayedSignOut extends OnboardingAuthFake {
      override async signOut() { await gate; await super.signOut(); }
    }
    const f = onboardingFakes();
    const auth = new DelayedSignOut();
    auth.signedIn = true;
    const command = new OnboardingCommand(f.profiles, () => f.api, auth);
    const callbacks: string[] = [];
    const harness = new MobileRenderHarness();
    await harness.render(<OnboardingScreen command={command}
      initialState={{ step: 'tenant' }} onStateChange={() => callbacks.push('state')}
      onStartOver={() => callbacks.push('start-over')} onComplete={() => callbacks.push('complete')} />);
    let pending!: Promise<void>;
    await harness.run(() => { pending = harness.byLabel('Sign out and start over')!.props.onPress(); });
    await harness.unmount();
    finish();
    await pending;
    expect(auth.signOuts).toBe(1);
    expect(callbacks).toEqual([]);
  });

  it('keeps connection unavailable for blank text and enables nonempty input for validation', async () => {
    const { harness, auth } = await fixture();
    const action = () => harness.byLabel('Connect and sign in');
    expect(action()?.props.accessibilityState.disabled).toBe(true);
    expect(harness.byText('Enter a server address to continue.')).toBeDefined();
    await harness.changeText(harness.byLabel('Server address'), '   ');
    await harness.press(action());
    expect(auth.signIns).toEqual([]);
    expect(action()?.props.accessibilityState.disabled).toBe(true);
    await harness.changeText(harness.byLabel('Server address'), 'ftp://example.invalid');
    expect(action()?.props.accessibilityState.disabled).toBe(false);
    await harness.press(action());
    expect(harness.byText('Enter a valid server address using https:// or http://.')).toBeDefined();
    await harness.unmount();
  });

  it.each(['tenant', 'inventory'] as const)('guards blank inventory names from keyboard submission in %s setup', async step => {
    const f = onboardingFakes();
    const command = new OnboardingCommand(f.profiles, () => f.api, f.auth);
    const harness = new MobileRenderHarness();
    await harness.render(<OnboardingScreen command={command} initialState={{ step }}
      onStateChange={() => {}} onComplete={() => {}} />);
    if (step === 'tenant') await harness.changeText(harness.byLabel('Household name'), 'Maple');
    const fieldLabel = step === 'tenant' ? 'First inventory' : 'Inventory name';
    await harness.changeText(harness.byLabel(fieldLabel), '  ');
    const action = harness.byLabel(step === 'tenant' ? 'Create household' : 'Create inventory');
    expect(action?.props.accessibilityState.disabled).toBe(true);
    expect(harness.byText('Enter an inventory name to continue.')).toBeDefined();
    await harness.run(() => harness.byLabel(fieldLabel)!.props.onSubmitEditing());
    expect(harness.byText('Sign in again to continue setup.')).toBeUndefined();
    expect(f.api.tenantWrites).toBe(0);
    expect(f.api.inventoryWrites).toBe(0);
    expect(harness.byLabel('Sign out and start over')?.props.disabled).toBe(false);
    await harness.changeText(harness.byLabel(fieldLabel), 'Garage');
    expect(harness.byLabel(step === 'tenant' ? 'Create household' : 'Create inventory')?.props.accessibilityState.disabled).toBe(false);
    await harness.unmount();
  });

});
