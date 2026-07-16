import { describe, expect, it } from 'vitest';
import { settingsOverviewDestinations, settingsResourceHref } from './settingsManagementNavigation';

describe('settings management navigation', () => {
  it('uses actual tenant and inventory names for compact drill-ins', () => {
    expect(settingsOverviewDestinations({
      tenant: { id: 'tenant-one', name: 'The Sell House' },
      inventory: { id: 'inventory-one', tenantId: 'tenant-one', name: 'Home' }
    })).toEqual([
      expect.objectContaining({ label: 'Account and app', href: '/settings/account/general' }),
      expect.objectContaining({ label: 'The Sell House', eyebrow: 'Tenant settings', href: '/settings/tenants/tenant-one' }),
      expect.objectContaining({ label: 'Home', eyebrow: 'Inventory settings', href: '/settings/tenants/tenant-one/inventories/inventory-one' })
    ]);
  });

  it('builds durable resource, lifecycle, and action links', () => {
    expect(settingsResourceHref({ level: 'inventory', tenantId: 'tenant-one', inventoryId: 'inventory-one', collection: 'fields', lifecycle: 'archived' }))
      .toBe('/settings/tenants/tenant-one/inventories/inventory-one/fields?lifecycle=archived');
    expect(settingsResourceHref({ level: 'tenant', tenantId: 'tenant-one', collection: 'asset-types', resourceId: 'type-one', action: 'edit' }))
      .toBe('/settings/tenants/tenant-one/asset-types/type-one/edit');
  });
});
