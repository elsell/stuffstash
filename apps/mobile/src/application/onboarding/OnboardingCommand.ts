import {
  ConnectionProfile,
  ConnectionProfileStore,
  normalizeInstanceUrl,
  SavedConnectionProfile
} from './ConnectionProfile';

export type OnboardingTenant = {
  readonly id: string;
  readonly name: string;
  readonly canCreateInventory: boolean;
};

export type OnboardingInventory = {
  readonly id: string;
  readonly tenantId: string;
};

export interface OnboardingApiPort {
  listTenants(): Promise<readonly OnboardingTenant[]>;
  listInventories(tenantId: string): Promise<readonly OnboardingInventory[]>;
  createTenant(name: string): Promise<OnboardingTenant>;
  createInventory(tenantId: string, name: string): Promise<OnboardingInventory>;
}

export type OnboardingStartStep = 'instance' | 'tenant' | 'inventory' | 'complete';

export type OnboardingStartState = {
  readonly step: OnboardingStartStep;
  readonly profile?: ConnectionProfile;
  readonly tenantName?: string;
};

export type OnboardingClientFactory = (profile: ConnectionProfile) => OnboardingApiPort;

export class OnboardingCommand {
  constructor(
    private readonly profiles: ConnectionProfileStore,
    private readonly clientFactory: OnboardingClientFactory,
    private readonly defaultDevToken?: string
  ) {}

  async getStartState(): Promise<OnboardingStartState> {
    const profile = await this.profiles.load();
    if (!profile) {
      return { step: 'instance' };
    }

    return this.resolveProfileState(this.toConnectionProfile(profile));
  }

  async saveInstanceUrl(input: { readonly apiBaseUrl: string }): Promise<OnboardingStartState> {
    if (!this.defaultDevToken) {
      throw new Error('Mobile authentication is not configured for this build.');
    }

    const savedProfile = {
      apiBaseUrl: normalizeInstanceUrl(input.apiBaseUrl)
    };
    await this.profiles.save(savedProfile);

    return this.resolveProfileState(this.toConnectionProfile(savedProfile));
  }

  async createTenant(input: {
    readonly profile: ConnectionProfile;
    readonly name: string;
  }): Promise<OnboardingStartState> {
    const tenantName = requireName(input.name, 'Name your tenant.');
    const tenant = await this.clientFactory(input.profile).createTenant(tenantName);
    const profile = { ...input.profile, tenantId: tenant.id };
    await this.profiles.save(toSavedProfile(profile));

    return this.resolveTenantState(profile, tenant);
  }

  async createInventory(input: {
    readonly profile: ConnectionProfile;
    readonly name: string;
  }): Promise<ConnectionProfile> {
    const tenantId = input.profile.tenantId;
    if (!tenantId) {
      throw new Error('Create a tenant before creating an inventory.');
    }

    const inventoryName = requireName(input.name, 'Name your inventory.');
    await this.clientFactory(input.profile).createInventory(tenantId, inventoryName);
    await this.profiles.save(toSavedProfile(input.profile));

    return input.profile;
  }

  async reset(): Promise<void> {
    await this.profiles.clear();
  }

  private async resolveProfileState(profile: ConnectionProfile): Promise<OnboardingStartState> {
    const client = this.clientFactory(profile);
    const tenants = await client.listTenants();
    const tenantWithInventory = await firstTenantWithInventory(client, tenants, profile.tenantId);
    const configuredTenant = tenants.find((tenant) => tenant.id === profile.tenantId);
    const firstCreatableTenant = tenants.find((tenant) => tenant.canCreateInventory);
    const firstTenant = tenantWithInventory ?? configuredTenant ?? firstCreatableTenant;

    if (!firstTenant) {
      if (tenants.length > 0) {
        throw new Error('No usable tenant is available for mobile onboarding.');
      }

      return { step: 'tenant', profile };
    }

    const nextProfile =
      profile.tenantId === firstTenant.id ? profile : { ...profile, tenantId: firstTenant.id };
    if (nextProfile !== profile) {
      await this.profiles.save(toSavedProfile(nextProfile));
    }

    return this.resolveTenantState(nextProfile, firstTenant);
  }

  private toConnectionProfile(profile: SavedConnectionProfile): ConnectionProfile {
    if (!this.defaultDevToken) {
      throw new Error('Mobile authentication is not configured for this build.');
    }

    return {
      ...profile,
      devToken: this.defaultDevToken
    };
  }

  private async resolveTenantState(
    profile: ConnectionProfile,
    tenant: OnboardingTenant
  ): Promise<OnboardingStartState> {
    const inventories = await this.clientFactory(profile).listInventories(tenant.id);
    if (inventories.length === 0) {
      if (!tenant.canCreateInventory) {
        throw new Error('No usable inventory is available for mobile onboarding.');
      }

      return { step: 'inventory', profile, tenantName: tenant.name };
    }

    return { step: 'complete', profile, tenantName: tenant.name };
  }
}

function toSavedProfile(profile: ConnectionProfile): SavedConnectionProfile {
  return {
    apiBaseUrl: profile.apiBaseUrl,
    tenantId: profile.tenantId
  };
}

function requireName(value: string, message: string): string {
  const trimmed = value.trim();
  if (trimmed.length === 0) {
    throw new Error(message);
  }
  return trimmed;
}

async function firstTenantWithInventory(
  client: OnboardingApiPort,
  tenants: readonly OnboardingTenant[],
  preferredTenantId: string | undefined
): Promise<OnboardingTenant | undefined> {
  const orderedTenants = [
    ...tenants.filter((tenant) => tenant.id === preferredTenantId),
    ...tenants.filter((tenant) => tenant.id !== preferredTenantId)
  ];

  for (const tenant of orderedTenants) {
    const inventories = await client.listInventories(tenant.id);
    if (inventories.length > 0) {
      return tenant;
    }
  }

  return undefined;
}
