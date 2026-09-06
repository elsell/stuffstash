import type { ConnectionProfile, ConnectionProfileStore } from './ConnectionProfile';
import type { OnboardingApiPort, OnboardingClientFactory, OnboardingStartState, OnboardingTenant } from './OnboardingCommand';

export class OnboardingRecoveryRequiredError extends Error {
  constructor() {
    super('The creation result is not yet available. Try again to check its status.');
  }
}

export class OnboardingPartialSetupError extends Error {
  constructor(readonly state: OnboardingStartState, readonly failure: unknown) {
    super('Your household is ready. Finish creating its first inventory.');
  }
}

// Keep ambiguous writes read-only on retry until discovery confirms their outcome.
// This state belongs to one onboarding command/session, never to the transport UI.
export class HouseholdSetup {
  private tenantAttempt?: { server: string; name: string; previousIds: readonly string[] };
  private inventoryAttempt?: { server: string; tenantId: string };
  private createdProfile?: ConnectionProfile;

  constructor(private readonly profiles: ConnectionProfileStore, private readonly clients: OnboardingClientFactory) {}

  clear() {
    this.tenantAttempt = undefined;
    this.inventoryAttempt = undefined;
    this.createdProfile = undefined;
  }

  async create(input: { profile: ConnectionProfile; householdName: string; inventoryName: string }, active: () => void): Promise<OnboardingStartState> {
    const householdName = requiredOnboardingName(input.householdName, 'Enter a household name.');
    const inventoryName = requiredOnboardingName(input.inventoryName, 'Enter an inventory name.');
    const profile = this.createdProfile?.apiBaseUrl === input.profile.apiBaseUrl ? this.createdProfile : input.profile;
    const client = this.clients(profile);
    const tenants = await client.listTenants();
    active();
    let tenant = profile.tenantId ? tenants.find(value => value.id === profile.tenantId) : undefined;
    if (profile.tenantId && !tenant) throw new Error('No usable tenant is available for mobile onboarding.');
    if (!tenant && this.tenantAttempt) tenant = this.recoveredTenant(tenants, profile.apiBaseUrl);
    if (!tenant && tenants.length) {
      throw new Error('Your available households changed. Sign out and sign in again to select one.');
    }
    if (!tenant) {
      this.tenantAttempt = { server: profile.apiBaseUrl, name: householdName, previousIds: tenants.map(value => value.id) };
      try {
        tenant = await client.createTenant(householdName);
      } catch (error) {
        active();
        if (isDefinitiveRejection(error)) { this.tenantAttempt = undefined; throw error; }
        const discovered = await client.listTenants();
        active();
        tenant = this.recoveredTenant(discovered, profile.apiBaseUrl);
      }
      active();
    }
    const nextProfile = { apiBaseUrl: profile.apiBaseUrl, tenantId: tenant.id };
    this.createdProfile = nextProfile;
    this.tenantAttempt = undefined;
    await this.profiles.save(nextProfile);
    active();
    try { return await this.createInventory({ profile: nextProfile, inventoryName }, active); }
    catch (failure) {
      active();
      throw new OnboardingPartialSetupError({ step: 'inventory', profile: nextProfile, tenantName: tenant.name }, failure);
    }
  }

  async createInventory(input: { profile: ConnectionProfile; inventoryName: string }, active: () => void): Promise<OnboardingStartState> {
    const name = requiredOnboardingName(input.inventoryName, 'Enter an inventory name.');
    const client = this.clients(input.profile);
    const tenants = await client.listTenants();
    active();
    const tenant = tenants.find(value => value.id === input.profile.tenantId);
    if (!tenant) throw new Error('No usable tenant is available for mobile onboarding.');
    const inventories = await client.listInventories(tenant.id);
    active();
    if (inventories.length) {
      this.inventoryAttempt = undefined;
      return { step: 'complete', profile: input.profile, tenantName: tenant.name };
    }
    if (!tenant.canCreateInventory) throw new Error('No usable inventory is available for mobile onboarding.');
    if (this.inventoryAttempt) throw new OnboardingRecoveryRequiredError();
    this.inventoryAttempt = { server: input.profile.apiBaseUrl, tenantId: tenant.id };
    try {
      await client.createInventory(tenant.id, name);
    } catch (error) {
      active();
      if (isDefinitiveRejection(error)) { this.inventoryAttempt = undefined; throw error; }
      await this.reconcileInventory(client, tenant.id, active);
    }
    active();
    this.inventoryAttempt = undefined;
    await this.profiles.save(input.profile);
    active();
    return { step: 'complete', profile: input.profile, tenantName: tenant.name };
  }

  private recoveredTenant(tenants: readonly OnboardingTenant[], server: string): OnboardingTenant {
    const attempt = this.tenantAttempt;
    if (!attempt || attempt.server !== server) throw new OnboardingRecoveryRequiredError();
    const matches = tenants.filter(value => value.name === attempt.name && !attempt.previousIds.includes(value.id));
    if (matches.length !== 1) throw new OnboardingRecoveryRequiredError();
    return matches[0];
  }

  private async reconcileInventory(client: OnboardingApiPort, tenantId: string, active: () => void) {
    const inventories = await client.listInventories(tenantId);
    active();
    if (!inventories.length) throw new OnboardingRecoveryRequiredError();
  }
}

export function requiredOnboardingName(value: string, message: string): string {
  if (!value.trim()) throw new Error(message);
  return value.trim();
}

function isDefinitiveRejection(error: unknown): boolean {
  const status = error && typeof error === 'object' && 'status' in error ? error.status : undefined;
  return typeof status === 'number' && status >= 400 && status < 500 && status !== 408;
}
