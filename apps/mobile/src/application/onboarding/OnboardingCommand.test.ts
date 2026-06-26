import { describe, expect, it } from 'vitest';
import {
  ConnectionProfileStore,
  normalizeInstanceUrl,
  SavedConnectionProfile
} from './ConnectionProfile';
import {
  OnboardingCommand,
  OnboardingInventory,
  OnboardingTenant
} from './OnboardingCommand';

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

  async listTenants(): Promise<readonly OnboardingTenant[]> {
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
  it('saves an instance URL and requests tenant creation when no tenant exists', async () => {
    const store = new FakeProfileStore();
    const client = new FakeOnboardingClient();
    const command = new OnboardingCommand(store, () => client, 'dev:user-1');

    await expect(command.saveInstanceUrl({ apiBaseUrl: 'http://localhost:8080' })).resolves.toEqual({
      step: 'tenant',
      profile: { apiBaseUrl: 'http://localhost:8080', devToken: 'dev:user-1' }
    });
    expect(store.profile).toEqual({ apiBaseUrl: 'http://localhost:8080' });
  });

  it('requires configured mobile authentication before saving an instance URL', async () => {
    const store = new FakeProfileStore();
    const client = new FakeOnboardingClient();
    const command = new OnboardingCommand(store, () => client);

    await expect(command.saveInstanceUrl({ apiBaseUrl: 'http://localhost:8080' })).rejects.toThrow(
      'Mobile authentication is not configured for this build.'
    );
    expect(store.profile).toBeUndefined();
  });

  it('creates a tenant, creates an inventory, and keeps the profile ready for app services', async () => {
    const store = new FakeProfileStore();
    const client = new FakeOnboardingClient();
    const command = new OnboardingCommand(store, () => client, 'dev:user-1');
    const instanceState = await command.saveInstanceUrl({ apiBaseUrl: 'http://localhost:8080' });
    if (!instanceState.profile) {
      throw new Error('expected saved profile');
    }

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
      devToken: 'dev:user-1',
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

  it('does not route users into inventory creation for a tenant they cannot configure', async () => {
    const store = new FakeProfileStore();
    const client = new FakeOnboardingClient();
    client.tenants = [
      { id: 'tenant-viewer', name: 'Shared View', canCreateInventory: false }
    ];
    const command = new OnboardingCommand(store, () => client, 'dev:user-1');

    await expect(command.saveInstanceUrl({ apiBaseUrl: 'http://localhost:8080' })).rejects.toThrow(
      'No usable tenant is available for mobile onboarding.'
    );
  });
});
