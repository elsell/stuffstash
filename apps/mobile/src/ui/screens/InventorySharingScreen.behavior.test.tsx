import { AppFeedbackProvider } from '../feedback/AppFeedback';
import React from 'react';
import { pressAlertButton } from '../../test-support/react-native';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { createMobileQueryClient, mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { QueryClientInvitationMutationObserver } from '../../adapters/serverState/QueryClientInvitationMutationObserver';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { InventorySharingScreen } from './InventorySharingScreen';
import { CreateInventoryInvitationCommand, CancelInventoryInvitationCommand, ListInventoryInvitationsQuery, type InventoryInvitationManagementRepository, type InventoryInvitationSummary, type InventorySharingScope } from '../../application/sharing/InventorySharing';

const scope: InventorySharingScope = { tenantId: 'tenant', inventoryId: 'inventory', inventoryName: 'Garage', permissions: ['share'] };
const item: InventoryInvitationSummary = { id: 'one', email: 'old@example.test', relationship: 'viewer', status: 'pending', isExpired: false, expiresAt: '2027-01-01' };
const settle = async (h: MobileRenderHarness) => { await h.run(() => new Promise(r => setTimeout(r, 10))); };
it('reuses safe invitation pages and keeps a created secret out of cache', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); const calls: (string | undefined)[] = []; let rows = [item]; let failRefresh = false;
  const repository: InventoryInvitationManagementRepository = {
    list: async (_scope, request) => { calls.push(request?.cursor); if (failRefresh) throw new Error('offline'); return { items: request?.cursor ? [{ ...item, id: 'two', email: 'older@example.test' }] : rows, nextCursor: request?.cursor ? undefined : 'next' }; },
    create: async (_scope, input) => { const created = { ...item, id: 'new', email: input.email }; rows = [created, ...rows]; return { ...created, inviteUrl: 'https://example.test/#token=secret' }; },
    cancel: async () => undefined
  };
  const observer = new QueryClientInvitationMutationObserver(client, 'scope');
  const query = new ListInventoryInvitationsQuery(repository);
  const props = { listQuery: query, createCommand: new CreateInventoryInvitationCommand(repository, observer), cancelCommand: new CancelInventoryInvitationCommand(repository, observer), linkActions: { copy: async () => undefined, share: async () => undefined }, scope };
  const view = (show: boolean, composition = 'scope') => <MobileServerStateProvider client={client} scopeId={composition} loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><AppFeedbackProvider>{show && <InventorySharingScreen {...props} />}</AppFeedbackProvider></MobileServerStateProvider>;
  try {
    await h.render(view(true)); await settle(h); expect(h.allText()).toContain('old@example.test'); expect(calls).toEqual([undefined]);
    await h.press(h.byLabel('Load older invitations')); await settle(h); expect(h.allText()).toContain('older@example.test');
    await h.render(view(false)); await h.render(view(true)); await settle(h); expect(calls).toEqual([undefined, 'next']);
    await h.changeText(h.byLabel('Invitee email'), 'new@example.test'); await h.press(h.byLabel('Create Invitation')); await settle(h);
    expect(h.allText()).toContain('https://example.test/#token=secret');
    expect(JSON.stringify(client.getQueryCache().getAll().map(q => q.state.data))).not.toContain('secret');
    failRefresh = true;
    await h.press(h.byLabel('Cancel invitation for new@example.test'));
    await h.run(() => pressAlertButton('Cancel Invitation')); await settle(h);
    expect(h.byLabel('Cancel invitation for new@example.test')).toBeUndefined();
    failRefresh = false;
    await h.render(view(true, 'replacement')); await settle(h);
    expect(h.allText()).not.toContain('https://example.test/#token=secret');
    await h.render(view(false)); await h.render(view(true)); await settle(h); expect(h.allText()).not.toContain('https://example.test/#token=secret');
  } finally { await h.unmount(); }
});

it('hides cached invitations after denial and cancels a departed scope read', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); let denied = false; let signal: AbortSignal | undefined;
  const repository: InventoryInvitationManagementRepository = {
    list: async (selected, request) => { if (selected.inventoryId === 'other') { signal = request?.signal; return new Promise(() => undefined); } if (denied) throw Object.assign(new Error('denied'), { status: 403 }); return { items: [item] }; },
    create: async () => ({ ...item, inviteUrl: 'secret' }), cancel: async () => undefined
  };
  const query = new ListInventoryInvitationsQuery(repository);
  const view = (selected: InventorySharingScope, show = true) => <MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: selected.inventoryId })}><AppFeedbackProvider>{show && <InventorySharingScreen scope={selected} listQuery={query} createCommand={new CreateInventoryInvitationCommand(repository)} cancelCommand={new CancelInventoryInvitationCommand(repository)} linkActions={{ copy: async () => undefined, share: async () => undefined }} />}</AppFeedbackProvider></MobileServerStateProvider>;
  try {
    await h.render(view(scope)); await settle(h); expect(h.allText()).toContain('old@example.test');
    denied = true; await h.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.invitations('scope', 'tenant', 'inventory') })); await settle(h);
    expect(h.allText()).not.toContain('old@example.test');
    await h.render(view({ ...scope, inventoryId: 'other' })); await settle(h); expect(signal?.aborted).toBe(false);
    await h.render(view(scope, false)); expect(signal?.aborted).toBe(true);
  } finally { await h.unmount(); }
});

it('chooses access in place and preserves the submitted draft while creation fails', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  let rejectCreate: ((error: Error) => void) | undefined;
  const submitted: { email: string; relationship: string }[] = [];
  const repository: InventoryInvitationManagementRepository = {
    list: async () => ({ items: [] }),
    create: async (_scope, input) => { submitted.push(input); return new Promise((_resolve, reject) => { rejectCreate = reject; }); },
    cancel: async () => undefined
  };
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><AppFeedbackProvider><InventorySharingScreen scope={scope} listQuery={new ListInventoryInvitationsQuery(repository)} createCommand={new CreateInventoryInvitationCommand(repository)} cancelCommand={new CancelInventoryInvitationCommand(repository)} linkActions={{ copy: async () => undefined, share: async () => undefined }} /></AppFeedbackProvider></MobileServerStateProvider>);
    await settle(h);
    await h.press(h.byLabel('Choose invitation access'));
    await h.press(h.byLabel('Editor'));
    await h.changeText(h.byLabel('Invitee email'), 'friend@example.test');
    await h.press(h.byLabel('Create Invitation'));
    expect(submitted).toEqual([{ email: 'friend@example.test', relationship: 'editor' }]);
    expect(h.byLabel('Invitee email')?.props.editable).toBe(false);
    expect(h.byLabel('Choose invitation access')?.props.disabled).toBe(true);
    await h.changeText(h.byLabel('Invitee email'), 'replacement@example.test');
    await h.run(() => rejectCreate?.(new Error('offline')));
    expect(h.byLabel('Invitee email')?.props.value).toBe('friend@example.test');
    expect(h.byLabel('Invitee email')?.props.editable).toBe(true);
    expect(h.byLabel('Choose invitation access')?.props.disabled).toBe(false);
  } finally { await h.unmount(); client.clear(); }
});
