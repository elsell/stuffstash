import { describe, expect, it } from 'vitest';
import {
  settingsAuditScopeOptions,
  settingsAuditScopeHref,
  settingsInvitationStatusHref,
  settingsSectionHref
} from './workspaceSettingsNavigation';

describe('workspace settings navigation', () => {
  it('builds canonical settings section hrefs with section-specific durable filter state', () => {
    expect(settingsSectionHref('tenant-one', 'inventory-one', 'overview', 'revoked', 'tenant')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings'
    );
    expect(settingsSectionHref('tenant-one', 'inventory-one', 'access', 'revoked', 'tenant')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings/access?invitationStatus=revoked'
    );
    expect(settingsSectionHref('tenant-one', 'inventory-one', 'activity', 'revoked', 'tenant')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings/activity?auditScope=tenant'
    );
    expect(settingsSectionHref('tenant-one', 'inventory-one', 'fields', 'revoked', 'tenant')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings/fields'
    );
  });

  it('builds access invitation and audit scope filter hrefs', () => {
    expect(settingsInvitationStatusHref('tenant-one', 'inventory-one', 'all')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings/access'
    );
    expect(settingsInvitationStatusHref('tenant-one', 'inventory-one', 'pending')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings/access?invitationStatus=pending'
    );
    expect(settingsAuditScopeHref('tenant-one', 'inventory-one', 'inventory')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings/activity'
    );
    expect(settingsAuditScopeHref('tenant-one', 'inventory-one', 'tenant')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings/activity?auditScope=tenant'
    );
  });

  it('builds audit scope segmented options with durable hrefs and availability', () => {
    expect(
      settingsAuditScopeOptions({
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        hasTenant: true,
        hasInventory: false
      })
    ).toEqual([
      {
        value: 'inventory',
        label: 'Inventory',
        href: '/tenants/tenant-one/inventories/inventory-one/settings/activity',
        disabled: true
      },
      {
        value: 'tenant',
        label: 'Tenant',
        href: '/tenants/tenant-one/inventories/inventory-one/settings/activity?auditScope=tenant',
        disabled: false
      }
    ]);
  });

  it('omits audit scope hrefs until an inventory route exists', () => {
    expect(
      settingsAuditScopeOptions({
        tenantId: 'tenant-one',
        inventoryId: null,
        hasTenant: true,
        hasInventory: false
      }).map((option) => option.href)
    ).toEqual([undefined, undefined]);
  });
});
