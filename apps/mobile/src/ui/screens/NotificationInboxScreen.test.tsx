import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { NotificationInboxQueries } from '../../application/notifications/NotificationInboxQueries';
import type { ExpirationNotification } from '../../domain/notifications/Notification';
import { NotificationInboxScreen } from './NotificationInboxScreen';
const alert: ExpirationNotification = { id: 'notice', assetId: 'item', title: 'Tylenol', parentAssetId: 'bin', customAssetTypeId: 'medicine', expiration: { date: '2027-02', precision: 'month' }, milestone: 'upcoming', createdAt: '2027-01-01T00:00:00Z' };
it('shows a calendar month and opens only the currently resolved item after marking read', async () => {
  const events: string[] = [];
  const queries = new NotificationInboxQueries({
    async listInbox() { return { items: [alert], pagination: { limit: 20, hasMore: false, nextCursor: null } }; },
    async getNotification() { events.push('resolve'); return { ...alert, assetId: 'current-item' }; },
    async markUnread() { throw new Error('Unread not exercised by this fixture'); }, async markRead() { events.push('read'); },
    async countUnreadPage() { return { count: 1, nextCursor: null }; },
    async markAllReadPage() { return { complete: true, nextCursor: null }; }
  }, { record() {} });
  const harness = new MobileRenderHarness();
  try {
    await harness.render(<NotificationInboxScreen tenantId="tenant" inventoryId="inventory" queries={queries} onOpenAsset={(id) => events.push(id)} onChanged={() => events.push('changed')} onSettings={() => undefined} />);
    await harness.settle();
    expect(harness.allText().join(' ')).toContain('February 2027');
    expect(harness.byLabel('Open Tylenol')?.props.accessibilityValue).toEqual({text:'Expires February 2027. Unread'});
    await harness.press(harness.byLabel('Open Tylenol'));
    expect(events).toEqual(['resolve', 'read', 'changed', 'current-item']);
    expect(harness.allText()).not.toContain('Unread');
    expect(harness.byLabel('Mark Tylenol unread')).toBeDefined();
  } finally { await harness.unmount(); }
});
it('retains a retry surface without claiming an empty inbox after a failed load', async () => {
  const harness = new MobileRenderHarness();
  const queries = { async list() { throw new Error('private transport error'); }, async open() { return ''; }, async setRead() { throw new Error('Read state not exercised by this fixture'); }, async markAllRead() {} };
  try {
    await harness.render(<NotificationInboxScreen tenantId="tenant" inventoryId="inventory" queries={queries} onOpenAsset={() => undefined} onChanged={() => undefined} onSettings={() => undefined} />);
    await harness.settle();
    expect(harness.allText().join(' ')).toContain('could not be loaded');
    expect(harness.allText().join(' ')).not.toContain('No notifications');
    expect(harness.allText().join(' ')).not.toContain('private transport');
    expect(harness.byLabel('Retry notifications')?.props.disabled).toBe(false);
  } finally { await harness.unmount(); }
});
it('refreshes unread results after marking all alerts read', async () => {
  const harness = new MobileRenderHarness();
  let read = false;
  const filters: boolean[] = [];
  const queries = {
    async list(_tenant: string, _inventory: string, options: { unreadOnly?: boolean } = {}) {
      filters.push(!!options.unreadOnly);
      return { items: read && options.unreadOnly ? [] : [{ ...alert, readAt: read ? '2027-01-02T00:00:00Z' : undefined }], pagination: { limit: 20, hasMore: false, nextCursor: null } };
    }, async open() { return 'item'; }, async setRead() { throw new Error('Read state not exercised by this fixture'); }, async markAllRead() { read = true; }
  };
  let changes = 0;
  try {
    await harness.render(<NotificationInboxScreen tenantId="tenant" inventoryId="inventory" queries={queries} onOpenAsset={() => undefined} onChanged={() => changes++} onSettings={() => undefined} />);
    await harness.settle();
    const segment = harness.all().find((node) => Array.isArray(node.props.values) && node.props.values.includes('Unread'));
    await harness.run(() => segment?.props.onValueChange('Unread'));
    await harness.press(harness.byLabel('Mark all read'));
    expect(filters).toEqual([false, true, true]);
    expect(changes).toBe(1);
    expect(harness.allText().join(' ')).toContain('No unread notifications.');
  } finally { await harness.unmount(); }
});
it('opens the immediate parent separately without marking the alert read',async()=>{
 const events:string[]=[];const harness=new MobileRenderHarness();
 const queries={async list(){return {items:[{...alert,parentTrail:[{assetId:'closet',title:'Hall closet',kind:'location' as const},{assetId:'bin',title:'Bin 8',kind:'container' as const}],parentTrailIncomplete:true}],pagination:{limit:20,hasMore:false,nextCursor:null}};},async open(){events.push('read');return 'item';},async setRead() { throw new Error('Read state not exercised by this fixture'); }, async markAllRead(){}};
 try {await harness.render(<NotificationInboxScreen tenantId="tenant" inventoryId="inventory" queries={queries} onOpenAsset={id=>events.push(id)} onChanged={()=>{}} onSettings={()=>{}}/>);await harness.settle();
 expect(harness.allText().join(' ')).toContain('Partial location path');
 await harness.press(harness.byLabel('Open location Bin 8'));expect(events).toEqual(['bin']);
 }finally{await harness.unmount();}
});
it('marks a read notification unread without opening the asset', async () => {
  const harness = new MobileRenderHarness();
  let read = true;
  let opened = false;
  let changes = 0;
  const queries = {
    async list() { return { items: [{ ...alert, readAt: read ? '2027-01-02T00:00:00Z' : undefined }], pagination: { limit: 20, hasMore: false, nextCursor: null } }; },
    async open() { opened = true; return 'item'; },
    async markAllRead() {},
    async setRead(_tenant: string, _inventory: string, id: string, value: boolean) { expect(id).toBe('notice'); read = value; }
  };
  try {
    await harness.render(<NotificationInboxScreen tenantId="tenant" inventoryId="inventory" queries={queries} onOpenAsset={() => { opened = true; }} onChanged={() => changes++} onSettings={() => undefined} />);
    await harness.settle();
    await harness.press(harness.byLabel('Mark Tylenol unread'));
    expect(read).toBe(false);
    expect(opened).toBe(false);
    expect(changes).toBe(1);
    expect(harness.byLabel('Open Tylenol')?.props.accessibilityValue).toEqual({text:'Expires February 2027. Unread'});
  } finally { await harness.unmount(); }
});
it('accepts another device marking a previously opened notification unread on refresh', async () => {
 const harness = new MobileRenderHarness();
 const queries = {
  async list() { return { items: [alert], pagination: { limit: 20, hasMore: false, nextCursor: null } }; },
  async open() { return 'item'; }, async markAllRead() {}, async setRead() {}
 };
 try {
  await harness.render(<NotificationInboxScreen tenantId="tenant" inventoryId="inventory" queries={queries} onOpenAsset={() => {}} onChanged={() => {}} onSettings={() => {}} />);
  await harness.settle();
  await harness.press(harness.byLabel('Open Tylenol'));
  expect(harness.byLabel('Mark Tylenol unread')).toBeDefined();
  await harness.run(() => harness.byType('ScrollView')?.props.refreshControl.props.onRefresh());
  await harness.settle();
  expect(harness.byLabel('Mark Tylenol read')).toBeDefined();
 } finally { await harness.unmount(); }
});
it('keeps empty and short inbox refresh surfaces as large as the viewport', async () => {
 const harness=new MobileRenderHarness();let loads=0;
 const queries={async list(){loads++;return {items:loads===1?[]:[alert],pagination:{limit:20,hasMore:false,nextCursor:null}};},async open(){return '';},async setRead(){},async markAllRead(){}};
 try{
  await harness.render(<NotificationInboxScreen tenantId="tenant" inventoryId="inventory" queries={queries} onOpenAsset={()=>{}} onChanged={()=>{}} onSettings={()=>{}}/>);
  await harness.settle();
  const scroll=harness.byType('ScrollView')!;
  expect(scroll.props.style).toMatchObject({flex:1});
  expect(scroll.props.contentContainerStyle).toMatchObject({flexGrow:1});
  expect(scroll.props.alwaysBounceVertical).toBe(true);
  await harness.run(()=>scroll.props.refreshControl.props.onRefresh());
  expect(loads).toBe(2);
  expect(harness.byLabel('Open Tylenol')).toBeDefined();
 }finally{await harness.unmount();}
});
