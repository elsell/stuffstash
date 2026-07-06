import { describe, expect, it } from 'vitest';
import {
  ConnectionProfileStore,
  normalizeInstanceUrl,
  SavedConnectionProfile
} from './ConnectionProfile';
import {
  OnboardingCommand,
  OnboardingAuthPort,
  OnboardingInventory,
  OnboardingTenant
} from './OnboardingCommand';
import { MobileAuthenticationRequiredError } from '../auth/MobileAuthSession';

class FakeProfileStore implements ConnectionProfileStore {
  profile: SavedConnectionProfile | undefined;

  async load(): Promise<SavedConnectionProfile | undefined> {
    return this.profile;
  }

  async save(profile: SavedConnectionProfile): Promise<void> {
    this.profile = profile;
  }

  async clear(): Promise<void> {
    this.profile = undefined;
  }
}

class FakeOnboardingClient {
  tenants: OnboardingTenant[] = [];
  inventoriesByTenant = new Map<string, OnboardingInventory[]>();
  listTenantsError: Error | undefined;

  async listTenants(): Promise<readonly OnboardingTenant[]> {
    if (this.listTenantsError) {
      throw this.listTenantsError;
    }
    return this.tenants;
  }

  async listInventories(tenantId: string): Promise<readonly OnboardingInventory[]> {
    return this.inventoriesByTenant.get(tenantId) ?? [];
  }

  async createTenant(name: string): Promise<OnboardingTenant> {
    const tenant = {
      id: `tenant-${this.tenants.length + 1}`,
      name,
      canCreateInventory: true
    } satisfies OnboardingTenant;
    this.tenants.push(tenant);
    return tenant;
  }

  async createInventory(tenantId: string, name: string): Promise<OnboardingInventory> {
    const inventory = {
      id: `inventory-${name.toLowerCase().replace(/\s+/g, '-')}`,
      tenantId
    } satisfies OnboardingInventory;
    this.inventoriesByTenant.set(tenantId, [
      ...(this.inventoriesByTenant.get(tenantId) ?? []),
      inventory
    ]);
    return inventory;
  }
}

describe('normalizeInstanceUrl', () => {
  it('normalizes a typed instance URL', () => {
    expect(normalizeInstanceUrl(' stuffstash.example.test/ ')).toBe(
      'https://stuffstash.example.test'
    );
    expect(normalizeInstanceUrl('http://192.168.1.117:8080/')).toBe(
      'http://192.168.1.117:8080'
    );
  });
});

describe('OnboardingCommand', () => {
  it('saves an instance URL and asks the user to sign in before tenant discovery', async () => {
    const store = new FakeProfileStore();
    const client = new FakeOnboardingClient();
    const auth = newFakeAuth();
    const command = new OnboardingCommand(store, () => client, auth);

    await expect(command.saveInstanceUrl({ apiBaseUrl: 'http://localhost:8080' })).resolves.toEqual({
      step: 'signIn',
      profile: { apiBaseUrl: 'http://localhost:8080' }
    });
    expect(auth.preparedApiBaseUrls).toEqual(['http://localhost:8080']);
    expect(store.profile).toEqual({ apiBaseUrl: 'http://localhost:8080' });
  });

  it('does not save an instance URL when mobile SSO metadata is unavailable', async () => {
    const store = new FakeProfileStore();
    const client = new FakeOnboardingClient();
    const auth = newFakeAuth();
    auth.prepareError = new Error('Stuff Stash mobile sign-in is not configured for this instance.');
    const command = new OnboardingCommand(store, () => client, auth);

    await expect(command.saveInstanceUrl({ apiBaseUrl: 'http://localhost:8080' })).rejects.toThrow(
      'Stuff Stash mobile sign-in is not configured for this instance.'
    );
    expect(store.profile).toBeUndefined();
  });

  it('signs in before requesting tenant creation when no tenant exists', async () => {
    const store = new FakeProfileStore();
    const client = new FakeOnboardingClient();
    const auth = newFakeAuth(false);
    const command = new OnboardingCommand(store, () => client, auth);

    const saved = await command.saveInstanceUrl({ apiBaseUrl: 'http://localhost:8080' });
    await expect(command.signIn({ profile: saved.profile! })).resolves.toEqual({
      step: 'tenant',
      profile: { apiBaseUrl: 'http://localhost:8080' }
    });
    expect(auth.signedInApiBaseUrls).toEqual(['http://localhost:8080']);
  });

  it('creates a tenant, creates an inventory, and keeps the profile ready for app services', async () => {
    const store = new FakeProfileStore();
    const client = new FakeOnboardingClient();
    const command = new OnboardingCommand(store, () => client, newFakeAuth(true));
    const instanceState = await command.saveInstanceUrl({ apiBaseUrl: 'http://localhost:8080' });
    if (!instanceState.profile) {
      throw new Error('expected saved profile');
    }
    await command.signIn({ profile: instanceState.profile });

    const tenantState = await command.createTenant({
      profile: instanceState.profile,
      name: 'Home'
    });

    expect(tenantState.step).toBe('inventory');
    expect(tenantState.profile?.tenantId).toBe('tenant-1');

    const readyProfile = await command.createInventory({
      profile: tenantState.profile!,
      name: 'Household'
    });

    expect(readyProfile).toEqual({
      apiBaseUrl: 'http://localhost:8080',
      tenantId: 'tenant-1'
    });
    expect(store.profile).toEqual({
      apiBaseUrl: 'http://localhost:8080',
      tenantId: 'tenant-1'
    });
    await expect(command.getStartState()).resolves.toEqual({
      step: 'complete',
      profile: readyProfile,
      tenantName: 'Home'
    });
  });

  it('returns to sign-in for the saved instance when startup auth is lost', async () => {
    const store = new FakeProfileStore();
    store.profile = { apiBaseUrl: 'http://localhost:8080', tenantId: 'tenant-1' };
    const client = new FakeOnboardingClient();
    client.listTenantsError = new MobileAuthenticationRequiredError();
    const auth = newFakeAuth(true);
    const command = new OnboardingCommand(store, () => client, auth);

    await expect(command.getStartState()).resolves.toEqual({
      step: 'signIn',
      profile: { apiBaseUrl: 'http://localhost:8080', tenantId: 'tenant-1' }
    });
    expect(auth.signedOutCount).toBe(1);
  });

  it('returns to sign-in when tenant discovery reports an API authentication error', async () => {
    const store = new FakeProfileStore();
    store.profile = { apiBaseUrl: 'http://localhost:8080', tenantId: 'tenant-1' };
    const client = new FakeOnboardingClient();
    client.listTenantsError = Object.assign(new Error('Authentication required.'), {
      status: 401,
      code: 'authentication_required'
    });
    const auth = newFakeAuth(true);
    const command = new OnboardingCommand(store, () => client, auth);

    await expect(command.getStartState()).resolves.toEqual({
      step: 'signIn',
      profile: { apiBaseUrl: 'http://localhost:8080', tenantId: 'tenant-1' }
    });
    expect(auth.signedOutCount).toBe(1);
  });

  it('expires an active session without clearing the saved connection profile', async () => {
    const store = new FakeProfileStore();
    store.profile = { apiBaseUrl: 'http://localhost:8080', tenantId: 'tenant-1' };
    const client = new FakeOnboardingClient();
    const auth = newFakeAuth(true);
    const command = new OnboardingCommand(store, () => client, auth);

    await expect(command.expireSession({ profile: store.profile })).resolves.toEqual({
      step: 'signIn',
      profile: { apiBaseUrl: 'http://localhost:8080', tenantId: 'tenant-1' }
    });

    expect(auth.signedOutCount).toBe(1);
    expect(store.profile).toEqual({ apiBaseUrl: 'http://localhost:8080', tenantId: 'tenant-1' });
  });

  it('does not route users into inventory creation for a tenant they cannot configure', async () => {
    const store = new FakeProfileStore();
    const client = new FakeOnboardingClient();
    client.tenants = [
      { id: 'tenant-viewer', name: 'Shared View', canCreateInventory: false }
    ];
    const command = new OnboardingCommand(store, () => client, newFakeAuth(true));

    const saved = await command.saveInstanceUrl({ apiBaseUrl: 'http://localhost:8080' });
    await expect(command.signIn({ profile: saved.profile! })).rejects.toThrow(
      'No usable tenant is available for mobile onboarding.'
    );
  });
});

function newFakeAuth(signedIn = true): OnboardingAuthPort & {
  signedIn: boolean;
  prepareError?: Error;
  preparedApiBaseUrls: string[];
  signedInApiBaseUrls: string[];
  signedOutCount: number;
} {
  return {
    signedIn,
    prepareError: undefined as Error | undefined,
    preparedApiBaseUrls: [] as string[],
    signedInApiBaseUrls: [] as string[],
    signedOutCount: 0,
    async status() {
      return this.signedIn
        ? {
            status: 'signed_in',
            session: {
              apiBaseUrl: 'http://localhost:8080',
              issuer: 'https://accounts.example.test',
              clientId: 'stuff-stash-mobile',
              idToken: 'id-token',
              expiresAt: 4_000
            }
          }
        : { status: 'signed_out' };
    },
    async prepareSignIn(apiBaseUrl: string) {
      if (this.prepareError) {
        throw this.prepareError;
      }
      this.preparedApiBaseUrls.push(apiBaseUrl);
    },
    async signIn(apiBaseUrl: string) {
      this.signedIn = true;
      this.signedInApiBaseUrls.push(apiBaseUrl);
    },
    async signOut() {
      this.signedIn = false;
      this.signedOutCount += 1;
    }
  };
}
