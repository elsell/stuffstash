import type { ConnectionProfileStore, SavedConnectionProfile } from './ConnectionProfile';
import type { OnboardingApiPort, OnboardingAuthPort, OnboardingInventory, OnboardingTenant } from './OnboardingCommand';

export const onboardingServer = 'https://stash.example.test';
export class OnboardingProfileFake implements ConnectionProfileStore {
  profile?: SavedConnectionProfile;
  async load() { return this.profile; }
  async save(profile: SavedConnectionProfile) { this.profile = profile; }
  async clear() { this.profile = undefined; }
}
export class OnboardingAuthFake implements OnboardingAuthPort {
  signedIn = false;
  signIns: string[] = [];
  signOuts = 0;
  signInError?: Error;
  beforeSignIn?: () => Promise<void>;
  async prepareSignIn() {}
  async signIn(server: string) {
    this.signIns.push(server);
    await this.beforeSignIn?.();
    if (this.signInError) throw this.signInError;
    this.signedIn = true;
  }
  async signOut() { this.signedIn = false; this.signOuts++; }
  async status() {
    return this.signedIn ? { status: 'signed_in' as const, session: {
      apiBaseUrl: onboardingServer, issuer: 'https://login.example.test', clientId: 'mobile',
      idToken: 'test-token', expiresAt: 9999999999999
    } } : { status: 'signed_out' as const };
  }
}
export class OnboardingApiFake implements OnboardingApiPort {
  tenants: OnboardingTenant[] = [];
  inventories: OnboardingInventory[] = [];
  listTenantsError?: Error;
  tenantWrites = 0;
  inventoryWrites = 0;
  failTenantAfterWrite = false;
  failInventoryAfterWrite = false;
  failInventoryBeforeWrite = false;
  constructor(readonly auth: OnboardingAuthFake) {}
  private authorize() { if (!this.auth.signedIn) throw Object.assign(new Error('Unauthorized'), { status: 401 }); }
  async listTenants() { this.authorize(); if (this.listTenantsError) throw this.listTenantsError; return [...this.tenants]; }
  async listInventories(tenantId: string) { this.authorize(); return this.inventories.filter(i => i.tenantId === tenantId); }
  async createTenant(name: string) {
    this.authorize(); this.tenantWrites++;
    const tenant = { id: `tenant-${this.tenantWrites}`, name, canCreateInventory: true };
    this.tenants.push(tenant);
    if (this.failTenantAfterWrite) throw new Error('Connection lost');
    return tenant;
  }
  async createInventory(tenantId: string, _name: string) {
    this.authorize();
    if (!this.tenants.find(t => t.id === tenantId)?.canCreateInventory) throw Object.assign(new Error('Forbidden'), { status: 403 });
    if (this.failInventoryBeforeWrite) throw Object.assign(new Error('Unavailable'), { status: 400 });
    this.inventoryWrites++;
    const inventory = { id: `inventory-${this.inventoryWrites}`, tenantId };
    this.inventories.push(inventory);
    if (this.failInventoryAfterWrite) throw new Error('Connection lost');
    return inventory;
  }
}
export function onboardingFakes() {
  const profiles = new OnboardingProfileFake();
  const auth = new OnboardingAuthFake();
  const api = new OnboardingApiFake(auth);
  return { profiles, auth, api };
}
