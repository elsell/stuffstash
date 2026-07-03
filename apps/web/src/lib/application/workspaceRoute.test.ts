import { describe, expect, it } from 'vitest';
import { parseWorkspaceRoute, workspaceRouteHref } from './workspaceRoute';

describe('workspace route state', () => {
  it('parses a tenant inventory location deep link', () => {
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/locations'))).toMatchObject({
      mode: 'locations',
      tenantId: 'tenant_1',
      inventoryId: 'inv_1'
    });
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/locations/loc_1'))).toMatchObject({
      mode: 'location',
      tenantId: 'tenant_1',
      inventoryId: 'inv_1',
      locationId: 'loc_1'
    });
  });

  it('parses asset edit and search filters', () => {
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/assets/asset_1/edit'))).toMatchObject({
      mode: 'asset',
      tenantId: 'tenant_1',
      inventoryId: 'inv_1',
      assetId: 'asset_1',
      action: 'edit',
      assetAction: 'edit'
    });

    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/search?q=drill&lifecycle=all&mode=exact'))).toMatchObject({
      mode: 'search',
      tenantId: 'tenant_1',
      lifecycleState: 'active',
      searchQuery: 'drill',
      searchLifecycleState: 'all',
      searchMode: 'exact'
    });
  });

  it('parses location edit as the underlying location asset edit route', () => {
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/locations/location_1/edit'))).toMatchObject({
      mode: 'asset',
      tenantId: 'tenant_1',
      inventoryId: 'inv_1',
      locationId: 'location_1',
      assetId: 'location_1',
      action: 'edit',
      assetAction: 'edit'
    });
  });

  it('parses durable asset actions and settings sections', () => {
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/assets/asset_1/move'))).toMatchObject({
      mode: 'asset',
      assetId: 'asset_1',
      assetAction: 'move'
    });
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/assets/asset_1/delete'))).toMatchObject({
      mode: 'asset',
      assetId: 'asset_1',
      assetAction: 'delete'
    });
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/assets/asset_1/archive'))).toMatchObject({
      mode: 'asset',
      assetId: 'asset_1',
      assetAction: 'archive'
    });
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/assets/asset_1/restore'))).toMatchObject({
      mode: 'asset',
      assetId: 'asset_1',
      assetAction: 'restore'
    });
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/settings/fields'))).toMatchObject({
      mode: 'settings',
      settingsSection: 'fields'
    });
    expect(
      parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/settings/fields/asset-types/type_1/archive'))
    ).toMatchObject({
      mode: 'settings',
      settingsSection: 'fields',
      customizationAction: 'archive_asset_type',
      customAssetTypeId: 'type_1'
    });
    expect(
      parseWorkspaceRoute(
        new URL('https://app.test/tenants/tenant_1/inventories/inv_1/settings/fields/field-definitions/field_1/archive')
      )
    ).toMatchObject({
      mode: 'settings',
      settingsSection: 'fields',
      customizationAction: 'archive_field_definition',
      customFieldDefinitionId: 'field_1'
    });
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/settings/access?invitationStatus=revoked'))).toMatchObject({
      mode: 'settings',
      settingsSection: 'access',
      invitationStatus: 'revoked'
    });
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/settings/activity?auditScope=tenant'))).toMatchObject({
      mode: 'settings',
      settingsSection: 'activity',
      auditScope: 'tenant'
    });
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/settings/nope'))).toMatchObject({
      mode: 'settings',
      settingsSection: 'overview',
      invitationStatus: 'all',
      auditScope: 'inventory'
    });
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/settings/fields?invitationStatus=revoked'))).toMatchObject({
      mode: 'settings',
      settingsSection: 'fields',
      invitationStatus: 'all',
      auditScope: 'inventory'
    });
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/settings/access?auditScope=tenant'))).toMatchObject({
      mode: 'settings',
      settingsSection: 'access',
      auditScope: 'inventory'
    });
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/import'))).toMatchObject({
      mode: 'import',
      importSourceType: 'legacy_homebox'
    });
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/import/legacy-homebox'))).toMatchObject({
      mode: 'import',
      importSourceType: 'legacy_homebox'
    });
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/import/legacy-homebox-csv'))).toMatchObject({
      mode: 'import',
      importSourceType: 'legacy_homebox_csv'
    });
  });

  it('parses attachment delete confirmation routes under the parent asset', () => {
    expect(
      parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/assets/asset_1/attachments/file_1/delete'))
    ).toMatchObject({
      mode: 'asset',
      tenantId: 'tenant_1',
      inventoryId: 'inv_1',
      assetId: 'asset_1',
      attachmentId: 'file_1',
      attachmentAction: 'delete',
      assetAction: null
    });
  });

  it('formats stable workspace hrefs', () => {
    expect(workspaceRouteHref({ mode: 'asset', assetId: 'asset 1', action: 'edit' }, 'tenant 1', 'inv 1')).toBe(
      '/tenants/tenant%201/inventories/inv%201/assets/asset%201/edit'
    );
    expect(workspaceRouteHref({ mode: 'search', searchQuery: 'garage shelf', searchLifecycleState: 'archived' }, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/search?q=garage+shelf&lifecycle=archived'
    );
    expect(workspaceRouteHref({ action: 'add', addKind: 'location' }, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/add/location'
    );
    expect(workspaceRouteHref({ mode: 'locations' }, 'tenant_1', 'inv_1')).toBe('/tenants/tenant_1/inventories/inv_1/locations');
    expect(workspaceRouteHref({ mode: 'asset', assetId: 'asset_1', assetAction: 'move' }, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/assets/asset_1/move'
    );
    expect(workspaceRouteHref({ mode: 'asset', assetId: 'asset_1', assetAction: 'archive' }, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/assets/asset_1/archive'
    );
    expect(workspaceRouteHref({ mode: 'settings', settingsSection: 'activity' }, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/settings/activity'
    );
    expect(workspaceRouteHref({ mode: 'settings', settingsSection: 'access', invitationStatus: 'pending' }, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/settings/access?invitationStatus=pending'
    );
    expect(workspaceRouteHref({ mode: 'settings', settingsSection: 'access', invitationStatus: 'all' }, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/settings/access'
    );
    expect(workspaceRouteHref({ mode: 'settings', settingsSection: 'activity', auditScope: 'tenant' }, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/settings/activity?auditScope=tenant'
    );
    expect(workspaceRouteHref({ mode: 'settings', settingsSection: 'activity', auditScope: 'inventory' }, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/settings/activity'
    );
    expect(
      workspaceRouteHref(
        { mode: 'settings', settingsSection: 'fields', customizationAction: 'archive_asset_type', customAssetTypeId: 'type 1' },
        'tenant_1',
        'inv_1'
      )
    ).toBe('/tenants/tenant_1/inventories/inv_1/settings/fields/asset-types/type%201/archive');
    expect(
      workspaceRouteHref(
        {
          mode: 'settings',
          settingsSection: 'fields',
          customizationAction: 'archive_field_definition',
          customFieldDefinitionId: 'field 1'
        },
        'tenant_1',
        'inv_1'
      )
    ).toBe('/tenants/tenant_1/inventories/inv_1/settings/fields/field-definitions/field%201/archive');
    expect(workspaceRouteHref({ mode: 'import' }, 'tenant_1', 'inv_1')).toBe('/tenants/tenant_1/inventories/inv_1/import');
    expect(workspaceRouteHref({ mode: 'import', importSourceType: 'legacy_homebox' }, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/import/legacy-homebox'
    );
    expect(workspaceRouteHref({ mode: 'import', importSourceType: 'legacy_homebox_csv' }, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/import/legacy-homebox-csv'
    );
    expect(
      workspaceRouteHref(
        { mode: 'asset', assetId: 'asset_1', attachmentId: 'file_1', attachmentAction: 'delete' },
        'tenant_1',
        'inv_1'
      )
    ).toBe('/tenants/tenant_1/inventories/inv_1/assets/asset_1/attachments/file_1/delete');
    expect(
      workspaceRouteHref(
        { mode: 'asset', locationId: 'location_1', assetId: 'location_1', action: 'edit', assetAction: 'edit' },
        'tenant_1',
        'inv_1'
      )
    ).toBe('/tenants/tenant_1/inventories/inv_1/locations/location_1/edit');
  });

  it('accepts inventory-only compatibility aliases', () => {
    expect(parseWorkspaceRoute(new URL('https://app.test/inventories/inv_1/assets/asset_1/edit'))).toMatchObject({
      mode: 'asset',
      tenantId: null,
      inventoryId: 'inv_1',
      assetId: 'asset_1',
      action: 'edit'
    });

    expect(workspaceRouteHref({ mode: 'search', inventoryId: 'inv_1', searchQuery: 'drill' }, null, null)).toBe(
      '/inventories/inv_1/search?q=drill'
    );
  });

  it('falls back for malformed encoded paths', () => {
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/%E0%A4%A'))).toMatchObject({
      mode: 'home',
      tenantId: null,
      inventoryId: null
    });
  });

  it('falls back for unsupported trailing route segments', () => {
    const unsupported = [
      'https://app.test/tenants/tenant_1/inventories/inv_1/locations/loc_1/junk',
      'https://app.test/tenants/tenant_1/inventories/inv_1/assets/asset_1/edit/junk',
      'https://app.test/tenants/tenant_1/inventories/inv_1/assets/asset_1/share',
      'https://app.test/tenants/tenant_1/inventories/inv_1/assets/asset_1/attachments/file_1/archive',
      'https://app.test/tenants/tenant_1/inventories/inv_1/assets/asset_1/attachments/file_1/delete/junk',
      'https://app.test/tenants/tenant_1/inventories/inv_1/search/junk',
      'https://app.test/tenants/tenant_1/inventories/inv_1/settings/fields/junk',
      'https://app.test/tenants/tenant_1/inventories/inv_1/settings/fields/asset-types/type_1/edit',
      'https://app.test/tenants/tenant_1/inventories/inv_1/settings/fields/nope/type_1/archive',
      'https://app.test/tenants/tenant_1/inventories/inv_1/import/junk',
      'https://app.test/tenants/tenant_1/inventories/inv_1/import/legacy-homebox/junk',
      'https://app.test/tenants/tenant_1/inventories/inv_1/add/location/junk'
    ];

    for (const href of unsupported) {
      expect(parseWorkspaceRoute(new URL(href))).toMatchObject({
        mode: 'home',
        tenantId: null,
        inventoryId: null
      });
    }
  });
});
