import type { StuffStashClient } from '@stuff-stash/api-client';
import {
  OnboardingApiPort,
  OnboardingInventory,
  OnboardingTenant
} from '../../application/onboarding/OnboardingCommand';

type OnboardingApiClient = Pick<
  StuffStashClient,
  'createTenant' | 'createInventory' | 'listMyTenants' | 'listInventories'
>;

export class ApiOnboardingGateway implements OnboardingApiPort {
  constructor(private readonly client: OnboardingApiClient) {}

  async listTenants(): Promise<readonly OnboardingTenant[]> {
    const page = await this.client.listMyTenants(100);
    return page.items.map((tenant) => ({
      id: tenant.id,
      name: tenant.name,
      canCreateInventory: tenant.access.permissions.includes('create_inventory')
    }));
  }

  async listInventories(tenantId: string): Promise<readonly OnboardingInventory[]> {
    const page = await this.client.listInventories(tenantId, 100);
    return page.items.map((inventory) => ({
      id: inventory.id,
      tenantId: inventory.tenantId
    }));
  }

  async createTenant(name: string): Promise<OnboardingTenant> {
    const tenant = await this.client.createTenant(name);
    return {
      id: tenant.id,
      name: tenant.name,
      canCreateInventory: tenant.access.permissions.includes('create_inventory')
    };
  }

  async createInventory(tenantId: string, name: string): Promise<OnboardingInventory> {
    const inventory = await this.client.createInventory(tenantId, name);
    return {
      id: inventory.id,
      tenantId: inventory.tenantId
    };
  }
}
