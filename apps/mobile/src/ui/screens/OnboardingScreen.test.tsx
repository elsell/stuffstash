import { useState } from 'react';
import { describe, expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { OnboardingCommand, type OnboardingStartState } from '../../application/onboarding/OnboardingCommand';
import { onboardingFakes, onboardingServer } from '../../application/onboarding/OnboardingTestSupport';
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
    expect(harness.byLabel('Server address')?.props.value).toBe('');
    await harness.changeText(harness.byLabel('Server address'), onboardingServer);
    await harness.press(harness.byLabel('Connect and sign in'));
    expect(harness.byLabel('Household name')?.props.value).toBe('');
    expect(harness.byLabel('First inventory')?.props.value).toBe('Home Inventory');
    await harness.unmount();
  });
  it('shows field-specific errors and preserves the form for retry', async () => {
    const { harness, api } = await fixture();
    await harness.changeText(harness.byLabel('Server address'), onboardingServer);
    await harness.press(harness.byLabel('Connect and sign in'));
    await harness.press(harness.byLabel('Create household'));
    expect(harness.byText('Enter a household name.')).toBeDefined();
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
    expect(harness.byLabel('Inventory name')?.props.value).toBe('Workshop');
    api.failInventoryBeforeWrite = false;
    await harness.press(harness.byLabel('Create inventory'));
    expect(api.tenantWrites).toBe(1);
    expect(api.inventoryWrites).toBe(1);
    await harness.unmount();
  });

});
