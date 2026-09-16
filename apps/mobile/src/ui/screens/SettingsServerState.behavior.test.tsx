import { AboutSettingsScreen, ConnectionSettingsScreen, DiagnosticsSettingsScreen } from './SettingsDetailScreens';
import { AppFeedbackProvider } from '../feedback/AppFeedback';
import { ApiSettingsScopeRepository } from '../../adapters/settings/ApiSettingsScopeRepository';
import React from 'react';
import { Text } from 'react-native';
import { expect, it } from 'vitest';
import { SettingsQuery } from '../../application/settings/SettingsQuery';
import { useSettingsModel } from './SettingsScreenState';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { createMobileQueryClient, mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';

const settle = (harness: MobileRenderHarness) => harness.run(() => new Promise((resolve) => setTimeout(resolve, 10)));
it('keeps local diagnostics visible while identity loads, fails and recovers', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  let fail!: (error: Error) => void;
  const pending = new Promise<never>((_, reject) => { fail = reject; });
  let recovering = false;
  const query = new SettingsQuery({ getCurrentPrincipal: async () => {
    if (!recovering) return pending;
    return { id: 'principal-recovered' };
  } }, { getDiagnostics: () => ({ apiBaseUrl: 'https://example.test', appVersion: 'local-version', authenticationMode: 'oidc-sso' }) }, { getSelectedScope: async () => {
    if (!recovering) return pending;
    return { tenant: { id: 'tenant-recovered', name: 'Home', permissions: [] }, inventory: { id: 'inventory', name: 'Garage', permissions: [] } };
  } });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><DiagnosticsSettingsScreen settingsQuery={query} /></MobileServerStateProvider>);
    await settle(h);
    expect(h.allText()).toContain('local-version');
    expect(h.allText()).toContain('https://example.test');
    expect(h.allText()).toContain('Loading account identity');
    expect(h.allText()).toContain('Loading household identity');
    await h.run(() => fail(new Error('Private transport detail'))); await settle(h);
    expect(h.allText()).toContain('local-version');
    expect(h.allText().join(' ')).not.toContain('Private transport detail');
    recovering = true;
    await h.press(h.byLabel('Retry account identity')); await settle(h);
    await h.press(h.byLabel('Retry household identity')); await settle(h);
    expect(h.allText()).toContain('principal-recovered');
    expect(h.allText()).toContain('tenant-recovered');
  } finally { await h.unmount(); }
});

it('suppresses cached diagnostic identity on access failure while preserving local values', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  let denied = false; let unavailable = false;
  const query = new SettingsQuery({ getCurrentPrincipal: async () => {
    if (denied) throw Object.assign(new Error('Denied'), { status: 403 });
    if (unavailable) throw new Error('Unavailable');
    return { id: 'private-principal' };
  } }, { getDiagnostics: () => ({ apiBaseUrl: 'https://example.test', appVersion: 'local-version', authenticationMode: 'oidc-sso' }) }, { getSelectedScope: async () => {
    if (denied) throw Object.assign(new Error('Denied'), { status: 403 });
    if (unavailable) throw new Error('Unavailable');
    return { tenant: { id: 'private-tenant', name: 'Home', permissions: [] }, inventory: { id: 'inventory', name: 'Garage', permissions: [] } };
  } });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><DiagnosticsSettingsScreen settingsQuery={query} /></MobileServerStateProvider>);
    await settle(h); await settle(h);
    expect(h.allText()).toContain('private-principal');
    denied = true;
    await h.run(() => client.invalidateQueries()); await settle(h);
    expect(h.allText()).toContain('local-version');
    expect(h.allText().join(' ')).not.toContain('private-principal');
    expect(h.allText().join(' ')).not.toContain('private-tenant');
    denied = false; unavailable = true;
    await h.press(h.byLabel('Retry account identity')); await settle(h);
    await h.press(h.byLabel('Retry household identity')); await settle(h);
    expect(h.allText().join(' ')).not.toContain('private-principal');
    expect(h.allText().join(' ')).not.toContain('private-tenant');
    unavailable = false;
    await h.press(h.byLabel('Retry account identity')); await settle(h);
    await h.press(h.byLabel('Retry household identity')); await settle(h);
    expect(h.allText()).toContain('private-principal');
    expect(h.allText()).toContain('private-tenant');
  } finally { await h.unmount(); }
});

it('shares scope across Settings surfaces without waiting for principal identity', async () => {
  const harness = new MobileRenderHarness();
  const client = createMobileQueryClient();
  let scopeReads = 0;
  let principalReads = 0;
  const query = new SettingsQuery({ getCurrentPrincipal: () => { principalReads++; return new Promise(() => undefined); } }, { getDiagnostics: () => ({ apiBaseUrl: 'https://example.test', appVersion: 'test', authenticationMode: 'oidc-sso' }) }, { getSelectedScope: async () => { scopeReads++; return { tenant: { id: 'tenant', name: 'Home', permissions: [] }, inventory: { id: 'inventory', name: 'Garage', permissions: [] } }; } });
  function Surface() { const model = useSettingsModel(query); return <Text>{model.state.status === 'ready' ? model.state.settings.selectedInventory.name : 'Loading'}</Text>; }
  try {
    await harness.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><Surface /><Surface /></MobileServerStateProvider>);
    await settle(harness); await settle(harness);
    expect(harness.allText()).toEqual(['Garage', 'Garage']);
    expect(scopeReads).toBe(1);
    expect(principalReads).toBe(1);
  } finally { await harness.unmount(); }
});

it('does not expose cached permission controls after a denied scope refresh', async () => {
  const harness = new MobileRenderHarness(); const client = createMobileQueryClient();
  let denied = false;
  const query = new SettingsQuery({ getCurrentPrincipal: async () => ({ id: 'principal' }) }, { getDiagnostics: () => ({ apiBaseUrl: 'https://example.test', appVersion: 'test', authenticationMode: 'oidc-sso' }) }, { getSelectedScope: async () => { if (denied) throw Object.assign(new Error('denied'), { status: 403 }); return { tenant: { id: 'tenant', name: 'Home', permissions: ['configure'] }, inventory: { id: 'inventory', name: 'Garage', permissions: [] } }; } });
  function Surface() { const model = useSettingsModel(query); return <Text>{model.state.status === 'ready' ? model.state.settings.selectedTenant.permissions.join(',') : model.state.status}</Text>; }
  try {
    await harness.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><Surface /></MobileServerStateProvider>);
    await settle(harness); await settle(harness); expect(harness.allText()).toContain('configure');
    denied = true;
    await harness.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.settingsScope('scope', 'tenant', 'inventory') })); await settle(harness);
    expect(harness.allText()).toEqual(['error']);
  } finally { await harness.unmount(); }
});

it('hides warmed Settings permissions when the selected inventory disappears from authorized discovery', async () => {
  const harness = new MobileRenderHarness(); const client = createMobileQueryClient(); let removed = false;
  const scope = new ApiSettingsScopeRepository({
    getTenant: async () => ({ id: 'tenant', name: 'Home', access: { relationship: 'owner', permissions: ['configure'] } }),
    listInventories: async () => ({ items: removed ? [] : [{ id: 'inventory', tenantId: 'tenant', name: 'Garage', access: { relationship: 'owner', permissions: ['configure'] } }], pagination: { hasMore: false, nextCursor: null, limit: 100 } })
  }, { getCurrentSettingsScope: async () => ({ tenantId: 'tenant', inventory: { id: 'inventory', name: 'Old scope', permissions: ['configure'] } }) });
  const query = new SettingsQuery({ getCurrentPrincipal: async () => ({ id: 'principal' }) }, { getDiagnostics: () => ({ apiBaseUrl: 'https://example.test', appVersion: 'test', authenticationMode: 'oidc-sso' }) }, scope);
  function Surface() { const model = useSettingsModel(query); return <Text>{model.state.status === 'ready' ? model.state.settings.selectedInventory.permissions.join(',') : model.state.status}</Text>; }
  try {
    await harness.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><Surface /></MobileServerStateProvider>);
    await settle(harness); await settle(harness); expect(harness.allText()).toContain('configure');
    removed = true;
    await harness.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.settingsScope('scope', 'tenant', 'inventory') })); await settle(harness);
    expect(harness.allText()).toEqual(['error']);
  } finally { await harness.unmount(); }
});

it('hides warmed Settings controls when expired directory discovery loses the selected inventory', async () => {
  const { ApiInventoryDirectory } = await import('../../adapters/inventories/ApiInventoryDirectory');
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); let now = 0; let removed = false;
  const tenant = { id: 'tenant', name: 'Home', access: { relationship: 'owner' as const, permissions: ['configure'] } };
  const inventory = { id: 'inventory', tenantId: 'tenant', name: 'Garage', access: tenant.access };
  const directory = new ApiInventoryDirectory({ listMyTenants: async () => ({ items: [tenant], pagination: { limit: 100, hasMore: false, nextCursor: null } }), listInventories: async () => ({ items: removed ? [] : [inventory], pagination: { limit: 100, hasMore: false, nextCursor: null } }) }, 'tenant', () => now);
  const query = new SettingsQuery({ getCurrentPrincipal: async () => ({ id: 'principal' }) }, { getDiagnostics: () => ({ apiBaseUrl: 'https://example.test', appVersion: 'test', authenticationMode: 'oidc-sso' }) }, { getSelectedScope: async () => { const value = await directory.selected(); return { tenant: { id: value.tenant.id, name: value.tenant.name, permissions: value.tenant.access.permissions }, inventory: { id: value.inventory.id, name: value.inventory.name, permissions: value.inventory.access.permissions } }; } });
  function Surface() { const model = useSettingsModel(query); return <Text>{model.state.status === 'ready' ? model.state.settings.selectedInventory.permissions.join(',') : model.state.status}</Text>; }
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><Surface /></MobileServerStateProvider>); await settle(h); await settle(h); expect(h.allText()).toContain('configure');
    removed = true; now = 300_001;
    await h.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.settingsScope('scope', 'tenant', 'inventory') })); await settle(h);
    expect(h.allText()).toEqual(['error']);
  } finally { await h.unmount(); }
});

it('renders About and Connection from local diagnostics without discovery or account reads', async () => {
  const h = new MobileRenderHarness(); let reads = 0;
  const query = new SettingsQuery({ getCurrentPrincipal: async () => { reads++; throw new Error('Unnecessary account read'); } }, { getDiagnostics: () => ({ apiBaseUrl: 'https://example.test', appVersion: 'local-version', authenticationMode: 'oidc-sso' }) }, { getSelectedScope: async () => { reads++; throw new Error('Unnecessary scope read'); } });
  try {
    await h.render(<AppFeedbackProvider><AboutSettingsScreen settingsQuery={query} /><ConnectionSettingsScreen settingsQuery={query} onChangeServer={async () => undefined} /></AppFeedbackProvider>);
    expect(h.allText()).toContain('local-version'); expect(h.allText()).toContain('https://example.test'); expect(reads).toBe(0);
  } finally { await h.unmount(); }
});
