import { describe, expect, it } from 'vitest';
import {
  settingsInvitationStatusOptions,
  settingsAuditScopeOptions,
  settingsAuditScopeHref,
  settingsInvitationStatusHref,
  settingsSectionHref,
  settingsSectionOptions
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

  it('builds settings section navigation options with metadata and current state', () => {
    expect(
      settingsSectionOptions({
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        section: 'access',
        invitationStatus: 'pending',
        auditScope: 'tenant'
      })
    ).toEqual([
      {
        value: 'overview',
        label: 'Overview',
        description: 'Inventory context and access summary',
        icon: 'boxes',
        href: '/tenants/tenant-one/inventories/inventory-one/settings',
        current: false
      },
      {
        value: 'access',
        label: 'Access',
        description: 'Sharing, grants, and invitations',
        icon: 'users',
        href: '/tenants/tenant-one/inventories/inventory-one/settings/access?invitationStatus=pending',
        current: true
      },
      {
        value: 'fields',
        label: 'Fields',
        description: 'Custom asset types and fields',
        icon: 'sliders',
        href: '/tenants/tenant-one/inventories/inventory-one/settings/fields',
        current: false
      },
      {
        value: 'activity',
        label: 'Activity',
        description: 'Audit history for this workspace',
        icon: 'activity',
        href: '/tenants/tenant-one/inventories/inventory-one/settings/activity?auditScope=tenant',
        current: false
      },
      {
        value: 'administration',
        label: 'Admin',
        description: 'Tenant and inventory administration',
        icon: 'user-cog',
        href: '/tenants/tenant-one/inventories/inventory-one/settings/administration',
        current: false
      }
    ]);
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

  it('builds invitation status segmented options with durable hrefs', () => {
    expect(settingsInvitationStatusOptions({ tenantId: 'tenant-one', inventoryId: 'inventory-one' })).toEqual([
      {
        value: 'all',
        label: 'All',
        href: '/tenants/tenant-one/inventories/inventory-one/settings/access'
      },
      {
        value: 'pending',
        label: 'Pending',
        href: '/tenants/tenant-one/inventories/inventory-one/settings/access?invitationStatus=pending'
      },
      {
        value: 'accepted',
        label: 'Accepted',
        href: '/tenants/tenant-one/inventories/inventory-one/settings/access?invitationStatus=accepted'
      },
      {
        value: 'revoked',
        label: 'Revoked',
        href: '/tenants/tenant-one/inventories/inventory-one/settings/access?invitationStatus=revoked'
      },
      {
        value: 'cancelled',
        label: 'Cancelled',
        href: '/tenants/tenant-one/inventories/inventory-one/settings/access?invitationStatus=cancelled'
      },
      {
        value: 'expired',
        label: 'Expired',
        href: '/tenants/tenant-one/inventories/inventory-one/settings/access?invitationStatus=expired'
      }
    ]);
  });

  it('omits invitation status hrefs until an inventory route exists', () => {
    expect(settingsInvitationStatusOptions({ tenantId: 'tenant-one', inventoryId: null }).map((option) => option.href)).toEqual([
      undefined,
      undefined,
      undefined,
      undefined,
      undefined,
      undefined
    ]);
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
