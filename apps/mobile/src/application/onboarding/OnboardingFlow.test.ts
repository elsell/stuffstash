import { describe, expect, it } from 'vitest';
import { OnboardingCommand } from './OnboardingCommand';
import { onboardingFakes, onboardingServer } from './OnboardingTestSupport';

function fixture() {
  const f = onboardingFakes();
  return { ...f, command: new OnboardingCommand(f.profiles, () => f.api, f.auth) };
}
const names = { householdName: 'Maple Street', inventoryName: 'Home Inventory' };
describe('combined onboarding flow', () => {
  it('connects and authenticates in one action before discovering setup', async () => {
    const { command, auth } = fixture();
    await expect(command.connectAndSignIn({ apiBaseUrl: ' stash.example.test/ ' })).resolves.toMatchObject({ step: 'tenant' });
    expect(auth.signIns).toEqual([onboardingServer]);
  });
  it('clears old tenant context before signing in to another server', async () => {
    const { command, profiles, auth } = fixture();
    profiles.profile = { apiBaseUrl: 'https://old.example.test', tenantId: 'old-tenant' };
    auth.signedIn = true;
    await command.connectAndSignIn({ apiBaseUrl: onboardingServer });
    expect(auth.signOuts).toBe(1);
    expect(profiles.profile).toEqual({ apiBaseUrl: onboardingServer });
  });
  it('preserves the normalized address when browser sign-in is canceled without creating anything', async () => {
    const { command, auth, profiles, api } = fixture();
    auth.signInError = new Error('Sign-in was cancelled.');
    await expect(command.connectAndSignIn({ apiBaseUrl: onboardingServer })).rejects.toThrow('cancelled');
    expect(profiles.profile?.apiBaseUrl).toBe(onboardingServer);
    expect(api.tenantWrites).toBe(0);
  });
  it('validates both names before creating either resource', async () => {
    const { command, api } = fixture();
    const state = await command.connectAndSignIn({ apiBaseUrl: onboardingServer });
    await expect(command.createHousehold({ profile: state.profile!, ...names, inventoryName: ' ' })).rejects.toThrow('Enter an inventory name.');
    expect(api.tenantWrites).toBe(0);
  });
  it('creates both resources and discovers them after relaunch', async () => {
    const { command, api, profiles, auth } = fixture();
    const state = await command.connectAndSignIn({ apiBaseUrl: onboardingServer });
    await expect(command.createHousehold({ profile: state.profile!, ...names })).resolves.toMatchObject({ step: 'complete' });
    expect(api.tenantWrites).toBe(1);
    expect(api.inventoryWrites).toBe(1);
    await expect(new OnboardingCommand(profiles, () => api, auth).getStartState()).resolves.toMatchObject({ step: 'complete' });
  });
  it('reuses the household after an inventory failure instead of creating a duplicate', async () => {
    const { command, api } = fixture();
    const state = await command.connectAndSignIn({ apiBaseUrl: onboardingServer });
    api.failInventoryBeforeWrite = true;
    await expect(command.createHousehold({ profile: state.profile!, ...names })).rejects.toThrow();
    api.failInventoryBeforeWrite = false;
    await expect(command.createHousehold({ profile: state.profile!, ...names })).resolves.toMatchObject({ step: 'complete' });
    expect(api.tenantWrites).toBe(1);
    expect(api.inventoryWrites).toBe(1);
  });
  it.each(['tenant', 'inventory'])('reconciles a lost %s creation response before retrying', async resource => {
    const { command, api } = fixture();
    const state = await command.connectAndSignIn({ apiBaseUrl: onboardingServer });
    api.failTenantAfterWrite = resource === 'tenant';
    api.failInventoryAfterWrite = resource === 'inventory';
    await expect(command.createHousehold({ profile: state.profile!, ...names })).resolves.toMatchObject({ step: 'complete' });
    expect(api.tenantWrites).toBe(1);
    expect(api.inventoryWrites).toBe(1);
  });
  it('does not create resources for an unauthenticated caller', async () => {
    const { command, api } = fixture();
    await expect(command.createHousehold({ profile: { apiBaseUrl: onboardingServer }, ...names })).rejects.toThrow();
    expect(api.tenantWrites).toBe(0);
  });
  it('does not turn read-only tenant access into inventory creation permission', async () => {
    const { command, auth, api } = fixture();
    auth.signedIn = true;
    api.tenants = [{ id: 'tenant-viewer', name: 'Shared', canCreateInventory: false }];
    await expect(command.createHousehold({ profile: { apiBaseUrl: onboardingServer, tenantId: 'tenant-viewer' }, ...names })).rejects.toThrow();
    expect(api.tenantWrites).toBe(0);
    expect(api.inventoryWrites).toBe(0);
  });
});

it('discards in-memory household recovery when signing in as another account on the same server', async () => {
  const { command, api, auth } = fixture();
  const state = await command.connectAndSignIn({ apiBaseUrl: onboardingServer });
  api.failInventoryBeforeWrite = true;
  await expect(command.createHousehold({ profile: state.profile!, ...names })).rejects.toThrow();
  await command.expireSession({ profile: state.profile! });
  api.tenants = [];
  api.failInventoryBeforeWrite = false;
  auth.signedIn = false;
  const second = await command.connectAndSignIn({ apiBaseUrl: onboardingServer });
  await expect(command.createHousehold({ profile: second.profile!, ...names })).resolves.toMatchObject({ step: 'complete' });
  expect(api.tenantWrites).toBe(2);
});

it('prevents duplicate submission while sign-in is pending', async () => {
  const { command, auth } = fixture();
  let release!: () => void;
  const waiting = new Promise<void>(resolve => { release = resolve; });
  auth.beforeSignIn = () => waiting;
  const first = command.connectAndSignIn({ apiBaseUrl: onboardingServer });
  await expect(command.connectAndSignIn({ apiBaseUrl: onboardingServer })).rejects.toThrow('already in progress');
  release();
  await first;
  expect(auth.signIns).toHaveLength(1);
});

it('serializes start-over after a profile save already in progress', async () => {
  const f = onboardingFakes();
  let release!: () => void;
  let started!: () => void;
  const waiting = new Promise<void>(resolve => { release = resolve; });
  const entered = new Promise<void>(resolve => { started = resolve; });
  const store = {
    load: () => f.profiles.load(),
    async save(profile: Parameters<typeof f.profiles.save>[0]) { started(); await waiting; await f.profiles.save(profile); },
    clear: () => f.profiles.clear()
  };
  const command = new OnboardingCommand(store, () => f.api, f.auth);
  const pending = command.connectAndSignIn({ apiBaseUrl: onboardingServer }).catch(() => undefined);
  await entered;
  const reset = command.reset();
  release();
  await Promise.all([pending, reset]);
  expect(f.profiles.profile).toBeUndefined();
  expect(f.auth.signIns).toHaveLength(0);
});

it('clears an API-rejected session instead of leaving setup authenticated', async () => {
  const { command, auth, api } = fixture();
  const state = await command.connectAndSignIn({ apiBaseUrl: onboardingServer });
  api.listTenantsError = Object.assign(new Error('Revoked'), { status: 401 });
  await expect(command.createHousehold({ profile: state.profile!, ...names })).rejects.toThrow('Sign in to Stuff Stash.');
  expect(auth.signedIn).toBe(false);
  expect(api.tenantWrites).toBe(0);
});
