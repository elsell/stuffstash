import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { StuffStashClient } from '@stuff-stash/api-client';
import { ApiNotificationRepository } from '../../adapters/notifications/ApiNotificationRepository';
import { NotificationPreferencesSession } from '../../application/notifications/NotificationPreferencesSession';
import { NotificationSettingsScreen } from './NotificationSettingsScreen';
import { setScreenFocused } from '../../test-support/navigation';
import { setAppStateForTest } from '../../test-support/react-native';
it('loads personal timing and saves a changed threshold without changing timezone', async () => {
  let preferences = { revision: 1, defaults: { enabled: true, upcoming: true, expired: true, advanceDays: 30 }, timezone: 'America/New_York', pushEnabled: false, overrides: [] };
  const repository = new ApiNotificationRepository(new StuffStashClient({ baseUrl: 'https://api.test', tokenProvider: () => 'token', fetch: async (input, init) => {
    const request = new Request(input, init);
    if (request.method === 'PUT') preferences = { ...preferences, ...await request.json(), revision: preferences.revision + 1 };
    return Response.json({ data: preferences, meta: {} });
  }}));
  const session = new NotificationPreferencesSession(repository, { record() {} }, 'tenant','inventory');
  await session.initialize('America/New_York');
  const harness = new MobileRenderHarness();
  try {
    await harness.render(<NotificationSettingsScreen page={{kind:"timing"}} onBack={()=>{}} onNavigate={()=>{}} tenantId="tenant" inventoryId="inventory" session={session} assetTypesQuery={{ async execute() { return []; } }} />);
    await harness.settle();
    await harness.press(harness.byLabel('Custom days'));
    expect(harness.byLabel('Days before expiration')).toBeDefined();
    await harness.changeText(harness.byLabel('Days before expiration'), '7');
    await harness.press(harness.byLabel('Save reminder days'));
    expect(preferences.defaults.advanceDays).toBe(7);
    expect(preferences.timezone).toBe('America/New_York');
  } finally { await harness.unmount(); }
});
it('handles denied permission, successful enablement and disabling without changing inbox defaults', async () => {
  let preferences = { revision: 1, defaults: { enabled: true, upcoming: true, expired: true, advanceDays: 30 }, timezone: 'UTC', pushEnabled: true, overrides: [] };
  const repository = new ApiNotificationRepository(new StuffStashClient({baseUrl:'https://api.test',tokenProvider:()=> 'token',fetch:async(input,init)=>{
    const request=new Request(input,init);
    if(request.method==='PUT') preferences={...preferences,...await request.json(),revision:preferences.revision+1};
    return Response.json({data:preferences,meta:{}});
  }}));
  const session = new NotificationPreferencesSession(repository,{record(){}},'tenant','inventory');
  let allowed=false;
  const pushSession={async enable(tenant:string,inventory:string){
    expect([tenant,inventory]).toEqual(['tenant','inventory']);
    if(!allowed)return 'denied' as const;
    preferences={...preferences,pushEnabled:true,revision:preferences.revision+1};
    return 'enabled' as const;
  }};
  const harness=new MobileRenderHarness();
  try {
    await harness.render(<NotificationSettingsScreen onBack={()=>{}} onNavigate={()=>{}} tenantId="tenant" inventoryId="inventory" session={session} assetTypesQuery={{async execute(){return [];}}} pushSession={pushSession} />);
    await harness.settle();
    await harness.press(harness.byLabel('Set up alerts on this device'));
    await harness.settle();
    expect(preferences.pushEnabled).toBe(true);
    expect(harness.allText()).toContain('Allow notifications for Stuff Stash in your device settings, then try again.');
    allowed=true;
    await harness.press(harness.byLabel('Set up alerts on this device'));
    await harness.settle();
    expect(harness.byLabel('Push notifications')!.props.value).toBe(true);
    await harness.run(()=>harness.byLabel('Push notifications')!.props.onValueChange(false));
    await harness.settle();
    expect(preferences.pushEnabled).toBe(false);
    expect(preferences.defaults).toEqual({enabled:true,upcoming:true,expired:true,advanceDays:30});
    expect(preferences.timezone).toBe('UTC');
    await harness.run(()=>harness.byLabel('Push notifications')!.props.onValueChange(true));
    await harness.settle();
    expect(preferences.pushEnabled).toBe(true);
    expect(harness.allText()).toContain('Open device settings');
    await harness.run(() => setScreenFocused(false));
    await harness.run(() => setScreenFocused(true));
    await harness.settle();
    expect(harness.allText()).toContain('Enable on this device');
    await harness.press(harness.byLabel('Set up alerts on this device'));
    await harness.settle();
    expect(harness.allText()).toContain('Open device settings');
    await harness.run(() => setAppStateForTest('background'));
    allowed = false;
    await harness.run(() => setAppStateForTest('active'));
    expect(harness.allText()).not.toContain('Notifications allowed on this device.');
    expect(harness.allText()).toContain('Enable on this device');
    expect(preferences.pushEnabled).toBe(true);
    await harness.press(harness.byLabel('Set up alerts on this device'));
    await harness.settle();
    expect(harness.allText()).toContain('Allow notifications for Stuff Stash in your device settings, then try again.');
  } finally {await harness.unmount(); setAppStateForTest('active'); setScreenFocused(true);}
});

it.each([['background', 'enabled'], ['inactive', 'enabled'], ['background', 'denied'], ['inactive', 'denied']] as const)('scopes delayed setup feedback across %s transitions with %s result', async (state, outcome) => {
  const preferences = { revision: 1, defaults: { enabled: true, upcoming: true, expired: true, advanceDays: 30 }, timezone: 'UTC', pushEnabled: true, overrides: [] };
  const repository = new ApiNotificationRepository(new StuffStashClient({ baseUrl: 'https://api.test', tokenProvider: () => 'token', fetch: async () => Response.json({ data: preferences, meta: {} }) }));
  const session = new NotificationPreferencesSession(repository, { record() {} }, 'tenant', 'inventory');
  let finish!: () => void;
  const pushSession = { async enable() { await new Promise<void>(resolve => { finish = resolve; }); return outcome; } };
  const h = new MobileRenderHarness();
  try {
    await h.render(<NotificationSettingsScreen onBack={() => {}} onNavigate={() => {}} tenantId="tenant" inventoryId="inventory" session={session} assetTypesQuery={{ async execute() { return []; } }} pushSession={pushSession} />);
    await h.settle();
    await h.press(h.byLabel('Set up alerts on this device'));
    await h.run(() => setAppStateForTest(state));
    await h.run(() => setAppStateForTest('active'));
    await h.run(() => finish());
    await h.settle();
    expect(h.allText().includes('Device setup completed.')).toBe(state === 'inactive' && outcome === 'enabled');
    expect(h.allText().includes('Allow notifications for Stuff Stash in your device settings, then try again.')).toBe(state === 'inactive' && outcome === 'denied');
    expect(h.allText()).toContain(state === 'inactive' && outcome === 'enabled' ? 'Open device settings' : 'Enable on this device');
    expect(Boolean(h.byLabel('Open notification system settings'))).toBe(state === 'inactive' && outcome === 'denied');
    expect(h.byLabel('Push notifications')?.props.value).toBe(true);
  } finally { await h.unmount(); setAppStateForTest('active'); }
});
