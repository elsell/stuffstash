import { describe, expect, it } from 'vitest';
import {
  settingsInvitationStatusOptions,
  settingsAuditScopeOptions,
  settingsAuditScopeHref,
  settingsInvitationStatusHref,
  settingsAdministrationPresentation,
  settingsOverviewPresentation,
  settingsSectionHref,
  settingsSectionOptions,
  settingsShellPresentation
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

  it('builds settings shell context and missing-inventory presentation', () => {
    const activeSection = {
      label: 'Access',
      description: 'Sharing, grants, and invitations'
    };
    expect(
      settingsShellPresentation({
        tenant: { name: 'Household' },
        inventory: { name: 'Garage' },
        activeSection
      })
    ).toEqual({
      title: 'Settings',
      contextLabel: 'Garage / Access',
      liveAnnouncement: 'Access: Sharing, grants, and invitations',
      overviewContextLabel: 'Household / Garage',
      emptyState: null
    });
    expect(
      settingsShellPresentation({
        tenant: null,
        inventory: { name: 'Garage' },
        activeSection
      }).contextLabel
    ).toBe('Garage / Access');
    expect(
      settingsShellPresentation({
        tenant: null,
        inventory: null,
        activeSection
      })
    ).toEqual({
      title: 'Settings',
      contextLabel: 'No inventory selected',
      liveAnnouncement: 'Access: Sharing, grants, and invitations',
      overviewContextLabel: 'Not available',
      emptyState: {
        title: 'No inventory selected',
        message: 'Select or create an inventory before managing settings.'
      }
    });
  });

  it('builds overview panel presentation without component-local fallback copy', () => {
    expect(
      settingsOverviewPresentation({
        tenantName: 'Household',
        inventoryCount: 3,
        accessRelationship: 'owner',
        canEditAssets: true,
        contextLabel: 'Household / Garage'
      })
    ).toEqual({
      title: 'Overview',
      contextLabel: 'Household / Garage',
      rows: [
        { label: 'Tenant', value: 'Household' },
        { label: 'Inventories', value: '3' },
        { label: 'Access', value: 'owner' },
        { label: 'Asset edits', value: 'Allowed' }
      ]
    });

    expect(
      settingsOverviewPresentation({
        tenantName: null,
        inventoryCount: 0,
        accessRelationship: 'viewer',
        canEditAssets: false,
        contextLabel: 'Not available'
      }).rows
    ).toEqual([
      { label: 'Tenant', value: 'Not available' },
      { label: 'Inventories', value: '0' },
      { label: 'Access', value: 'viewer' },
      { label: 'Asset edits', value: 'View only' }
    ]);
  });

  it('builds administration panel presentation from tenant configuration availability', () => {
    expect(settingsAdministrationPresentation({ canConfigureTenant: true })).toEqual({
      title: 'Administration',
      description: 'There are no administration actions available in the web app yet.'
    });
    expect(settingsAdministrationPresentation({ canConfigureTenant: false }).description).toBe(
      'This account does not have access to tenant administration.'
    );
  });
});
