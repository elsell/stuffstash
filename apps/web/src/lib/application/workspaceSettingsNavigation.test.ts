import { describe, expect, it } from 'vitest';
import {
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
});
