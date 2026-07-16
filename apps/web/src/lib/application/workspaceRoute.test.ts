import { describe, expect, expectTypeOf, it } from 'vitest';
import type { WorkspaceMode } from '$lib/domain/inventory';
import { parseWorkspaceRoute, workspaceRouteHref } from './workspaceRoute';

describe('workspace route state', () => {
  it('parses and formats the canonical Browse state', () => {
    const route = parseWorkspaceRoute(
      new URL(
        'https://app.test/tenants/tenant_1/inventories/inv_1/browse?surface=map&scope=containers&q=paint&tag=tag_2&tag=tag_1&lifecycle=all&availability=available&sort=id_asc&mode=exact'
      )
    );

    expect(route).toMatchObject({
      mode: 'browse',
      browseSurface: 'map',
      browseScope: 'containers',
      searchQuery: 'paint',
      browseTagIds: ['tag_2', 'tag_1'],
      searchLifecycleState: 'all',
      searchCheckoutState: 'available',
      browseSort: 'id_asc',
      searchMode: 'exact'
    });
    expect(workspaceRouteHref(route, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/browse?surface=map&scope=containers&q=paint&tag=tag_2&tag=tag_1&lifecycle=all&availability=available&sort=id_asc&mode=exact'
    );
  });

  it('normalizes duplicate and empty Browse tag query values at the route boundary', () => {
    const route = parseWorkspaceRoute(new URL(
      'https://app.test/tenants/tenant_1/inventories/inv_1/browse?tag=&tag=tag_2&tag=tag_2&tag=tag_1'
    ));
    expect(route.browseTagIds).toEqual(['tag_2', 'tag_1']);
    expect(workspaceRouteHref(route, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/browse?tag=tag_2&tag=tag_1'
    );
  });

  it('parses legacy Locations and Search routes as Browse compatibility state', () => {
    const locationsAlias = parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/locations'));
    expect(locationsAlias).toMatchObject({
      mode: 'browse',
      browseScope: 'places',
      compatibilityAlias: true
    });
    expect(workspaceRouteHref(locationsAlias, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/browse?scope=places'
    );

    const searchAlias = parseWorkspaceRoute(
      new URL('https://app.test/tenants/tenant_1/inventories/inv_1/search?q=drill&tagId=tag_tools&tagId=tag_camping&lifecycle=all&checkout=checked_out')
    );
    expect(searchAlias).toMatchObject({
      mode: 'browse',
      searchQuery: 'drill',
      browseTagIds: ['tag_tools', 'tag_camping'],
      searchLifecycleState: 'all',
      searchCheckoutState: 'checked_out',
      compatibilityAlias: true
    });
    expect(workspaceRouteHref(searchAlias, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/browse?q=drill&tag=tag_tools&tag=tag_camping&lifecycle=all&availability=checked_out'
    );
    expectTypeOf<Extract<WorkspaceMode, 'search' | 'locations'>>().toEqualTypeOf<never>();
  });

  it('parses a tenant inventory location deep link', () => {
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/locations'))).toMatchObject({
      mode: 'browse',
      browseScope: 'places',
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
      mode: 'browse',
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

  it('parses and builds the canonical location move-here route', () => {
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/locations/location_1/move-here'))).toMatchObject({
      mode: 'location',
      locationId: 'location_1',
      assetAction: 'move-here'
    });
    expect(workspaceRouteHref({ mode: 'location', locationId: 'location_1', assetAction: 'move-here' }, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/locations/location_1/move-here'
    );
  });

  it('parses durable asset actions and settings sections', () => {
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/assets/asset_1/move'))).toMatchObject({
      mode: 'asset',
      assetId: 'asset_1',
      assetAction: 'move'
    });
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/assets/asset_1/move-here'))).toMatchObject({
      mode: 'asset',
      assetId: 'asset_1',
      assetAction: 'move-here'
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
    expect(
      parseWorkspaceRoute(
        new URL('https://app.test/tenants/tenant_1/inventories/inv_1/settings/access/invitations/invite_1/cancel')
      )
    ).toMatchObject({
      mode: 'settings',
      settingsSection: 'access',
      accessInvitationAction: 'cancel',
      accessInvitationId: 'invite_1'
    });
    expect(
      parseWorkspaceRoute(
        new URL(
          'https://app.test/tenants/tenant_1/inventories/inv_1/settings/access/invitations/invite_1/delete?invitationStatus=pending'
        )
      )
    ).toMatchObject({
      mode: 'settings',
      settingsSection: 'access',
      invitationStatus: 'pending',
      accessInvitationAction: 'delete',
      accessInvitationId: 'invite_1'
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
      importSource: null
    });
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/import/homebox'))).toMatchObject({
      mode: 'import',
      importSource: 'homebox'
    });
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/import/homebox-csv'))).toMatchObject({
      mode: 'import',
      importSource: 'homebox-csv'
    });
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/import/jobs/job_1?tab=records'))).toMatchObject({
      mode: 'import',
      importSource: null,
      importJobId: 'job_1',
      importTab: 'records'
    });
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/import/jobs/job_1?tab=junk'))).toMatchObject({
      mode: 'import',
      importSource: null,
      importJobId: 'job_1',
      importTab: null
    });
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/add/item?parent=location_1'))).toMatchObject({
      action: 'add',
      addKind: 'item',
      addParentAssetId: 'location_1'
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
    expect(workspaceRouteHref({ action: 'add', addKind: 'location' }, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/add/location'
    );
    expect(workspaceRouteHref({ action: 'add', addKind: 'item', addParentAssetId: 'location 1' }, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/add/item?parent=location+1'
    );
    expect(workspaceRouteHref({ mode: 'asset', assetId: 'asset_1', assetAction: 'move' }, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/assets/asset_1/move'
    );
    expect(workspaceRouteHref({ mode: 'asset', assetId: 'asset_1', assetAction: 'move-here' }, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/assets/asset_1/move-here'
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
    expect(
      workspaceRouteHref(
        { mode: 'settings', settingsSection: 'access', accessInvitationAction: 'expire', accessInvitationId: 'invite 1' },
        'tenant_1',
        'inv_1'
      )
    ).toBe('/tenants/tenant_1/inventories/inv_1/settings/access/invitations/invite%201/expire');
    expect(
      workspaceRouteHref(
        {
          mode: 'settings',
          settingsSection: 'access',
          invitationStatus: 'pending',
          accessInvitationAction: 'delete',
          accessInvitationId: 'invite 1'
        },
        'tenant_1',
        'inv_1'
      )
    ).toBe('/tenants/tenant_1/inventories/inv_1/settings/access/invitations/invite%201/delete?invitationStatus=pending');
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
    expect(workspaceRouteHref({ mode: 'import', importSource: 'homebox' }, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/import/homebox'
    );
    expect(workspaceRouteHref({ mode: 'import', importSource: 'homebox-csv' }, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/import/homebox-csv'
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

    expect(workspaceRouteHref({ mode: 'browse', inventoryId: 'inv_1', searchQuery: 'drill' }, null, null)).toBe(
      '/inventories/inv_1/browse?q=drill'
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
      'https://app.test/tenants/tenant_1/inventories/inv_1/settings/access/invitations/invite_1/archive',
      'https://app.test/tenants/tenant_1/inventories/inv_1/settings/access/grants/grant_1/delete',
      'https://app.test/tenants/tenant_1/inventories/inv_1/settings/fields/junk',
      'https://app.test/tenants/tenant_1/inventories/inv_1/settings/fields/asset-types/type_1/edit',
      'https://app.test/tenants/tenant_1/inventories/inv_1/settings/fields/nope/type_1/archive',
      'https://app.test/tenants/tenant_1/inventories/inv_1/import/junk',
      'https://app.test/tenants/tenant_1/inventories/inv_1/import/legacy-homebox',
      'https://app.test/tenants/tenant_1/inventories/inv_1/import/legacy-homebox-csv',
      'https://app.test/inventories/inv_1/import',
      'https://app.test/inventories/inv_1/import/homebox',
      'https://app.test/inventories/inv_1/import/homebox-csv',
      'https://app.test/tenants/tenant_1/inventories/inv_1/add/location/junk'
    ];

    for (const href of unsupported) {
      const expectedTenantId = href.includes('/tenants/') ? 'tenant_1' : null;
      expect(parseWorkspaceRoute(new URL(href))).toMatchObject({
        mode: 'home',
        tenantId: expectedTenantId,
        inventoryId: 'inv_1'
      });
    }
  });
});
