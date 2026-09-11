import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { NotificationBell } from './NotificationBell';
it('initializes preferences before counting and opens the inbox', async () => {
  const calls: string[] = [];
  const client = createMobileQueryClient();
  const harness = new MobileRenderHarness();
  try {
    await harness.render(<MobileServerStateProvider client={client} scopeId="server-user" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <NotificationBell tenantId="tenant" inventoryId="inventory" initialize={async () => { calls.push('initialize'); }} count={async () => { calls.push('count'); return 4; }} onOpen={() => calls.push('open')} />
    </MobileServerStateProvider>);
    await harness.run(() => new Promise((resolve) => setTimeout(resolve, 30)));
    await harness.press(harness.byLabel('Notifications, 4 unread'));
    expect(calls).toEqual(['initialize', 'count', 'count', 'open']);
  } finally { await harness.unmount(); }
});
it('keeps inbox access when registration fails and does not claim zero unread', async () => {
  const client = createMobileQueryClient();
  const harness = new MobileRenderHarness();
  let counted = false;
  let opened = false;
  try {
    await harness.render(<MobileServerStateProvider client={client} scopeId="server-user" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <NotificationBell tenantId="tenant" inventoryId="inventory" initialize={async () => { throw new Error('unavailable'); }} count={async () => { counted = true; return 0; }} onOpen={() => { opened = true; }} />
    </MobileServerStateProvider>);
    await harness.run(() => new Promise((resolve) => setTimeout(resolve, 30)));
    expect(counted).toBe(false);
    expect(harness.byLabel('Notifications, 0 unread')).toBeUndefined();
    await harness.press(harness.byLabel('Notifications, unread count unavailable'));
    expect(opened).toBe(true);
  } finally { await harness.unmount(); }
});
