import { InventoryInvitationLinkUnavailableError } from '../../application/sharing/InventorySharing';
import { setScreenFocused } from '../../test-support/navigation';
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
async function openCancellation(h: MobileRenderHarness, email: string) {
  await h.press(h.byLabel(`Invitation actions for ${email}`));
  await h.press(h.byText('Cancel invitation')?.parent ?? undefined);
}
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
    await openCancellation(h, 'new@example.test');
    await h.run(() => pressAlertButton('Cancel Invitation')); await settle(h);
    expect(h.byLabel('Invitation actions for new@example.test')).toBeUndefined();
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
    expect(h.allText()).toContain('Could not create invitation');
    expect(h.byLabel('Invitee email')?.props.value).toBe('friend@example.test');
    expect(h.byLabel('Invitee email')?.props.editable).toBe(true);
    expect(h.byLabel('Choose invitation access')?.props.disabled).toBe(false);
  } finally { await h.unmount(); client.clear(); }
});

it.each(['scope', 'leave', 'return'] as const)('does not announce a late creation failure after %s', async departure => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  let rejectCreate: ((error: Error) => void) | undefined;
  const repository: InventoryInvitationManagementRepository = {
    list: async () => ({ items: [] }),
    create: async () => new Promise((_resolve, reject) => { rejectCreate = reject; }),
    cancel: async () => undefined
  };
  const view = (selected = scope, show = true) => <MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: selected.inventoryId })}><AppFeedbackProvider>{show && <InventorySharingScreen scope={selected} listQuery={new ListInventoryInvitationsQuery(repository)} createCommand={new CreateInventoryInvitationCommand(repository)} cancelCommand={new CancelInventoryInvitationCommand(repository)} linkActions={{ copy: async () => undefined, share: async () => undefined }} />}</AppFeedbackProvider></MobileServerStateProvider>;
  try {
    await h.render(view()); await settle(h);
    await h.changeText(h.byLabel('Invitee email'), 'friend@example.test');
    await h.press(h.byLabel('Create Invitation'));
    if (departure === 'scope') await h.render(view({ ...scope, inventoryId: 'other' }));
    else {
      await h.run(() => setScreenFocused(false));
      if (departure === 'return') await h.run(() => setScreenFocused(true));
      else await h.render(view(scope, false));
    }
    await h.run(() => rejectCreate?.(new Error('old inventory unavailable')));
    expect(h.allText()).not.toContain('Could not create invitation');
  } finally { await h.unmount(); client.clear(); setScreenFocused(true); }
});

it.each(['copy', 'share', 'cancel'] as const)('suppresses late %s feedback after leaving and returning', async action => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  let finish: (() => void) | undefined;
  const delayed = () => new Promise<void>((resolve, reject) => { finish = () => action === 'copy' ? resolve() : reject(new Error('old operation failed')); });
  const repository: InventoryInvitationManagementRepository = {
    list: async () => ({ items: [item] }),
    create: async () => ({ ...item, inviteUrl: 'https://example.test/#token=secret' }),
    cancel: delayed
  };
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><AppFeedbackProvider><InventorySharingScreen scope={scope} listQuery={new ListInventoryInvitationsQuery(repository)} createCommand={new CreateInventoryInvitationCommand(repository)} cancelCommand={new CancelInventoryInvitationCommand(repository)} linkActions={{ copy: delayed, share: delayed }} /></AppFeedbackProvider></MobileServerStateProvider>);
    await settle(h);
    if (action === 'cancel') {
      await openCancellation(h, 'old@example.test');
      await h.run(() => pressAlertButton('Cancel Invitation'));
    } else {
      await h.changeText(h.byLabel('Invitee email'), 'friend@example.test');
      await h.press(h.byLabel('Create Invitation')); await settle(h);
      const label = action === 'copy' ? 'Copy link' : 'Share invitation';
      await h.press(h.allByType('Pressable').find(node => node.children.some(child => typeof child === 'object' && child !== null && 'children' in child && child.children.includes(label))));
    }
    expect(finish).toBeDefined();
    await h.run(() => setScreenFocused(false));
    await h.run(() => setScreenFocused(true));
    await h.run(() => finish?.());
    expect(h.allText()).not.toContain('Invitation link copied');
    expect(h.allText()).not.toContain(`Could not ${action} invitation`);
  } finally { await h.unmount(); client.clear(); setScreenFocused(true); }
});


it('refreshes safe metadata and shows link-unavailable recovery inside the sharing form', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); let creations = 0;
  const repository: InventoryInvitationManagementRepository = {
    list: async () => ({ items: creations === 2 ? [{ ...item, id: 'second', email: 'second@example.test' }, item] : creations ? [item] : [] }),
    create: async () => { if (++creations === 1) return { ...item, inviteUrl: 'https://example.test/#token=previous' }; throw new InventoryInvitationLinkUnavailableError(); }, cancel: async () => undefined
  };
  const observer = new QueryClientInvitationMutationObserver(client, 'scope');
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><AppFeedbackProvider><InventorySharingScreen scope={scope} listQuery={new ListInventoryInvitationsQuery(repository)} createCommand={new CreateInventoryInvitationCommand(repository, observer)} cancelCommand={new CancelInventoryInvitationCommand(repository)} linkActions={{ copy: async () => { throw new Error('Must not copy'); }, share: async () => { throw new Error('Must not share'); } }} /></AppFeedbackProvider></MobileServerStateProvider>);
    await settle(h); await h.changeText(h.byLabel('Invitee email'), 'first@example.test');
    await h.press(h.byLabel('Create Invitation')); await settle(h);
    expect(h.byLabel('Complete invitation link')).toBeDefined();
    await h.changeText(h.byLabel('Invitee email'), 'friend@example.test');
    await h.press(h.byLabel('Create Invitation')); await settle(h);
    const recovery = h.byText('Invitation created, link unavailable');
    expect(recovery).toBeDefined();
    let parent = recovery?.parent;
    while (parent && parent.type !== 'ScrollView') parent = parent.parent;
    expect(parent?.type).toBe('ScrollView');
    expect(h.allText()).toContain('second@example.test');
    expect(h.byLabel('Invitee email')?.props.value).toBe('friend@example.test');
    expect(h.byLabel('Complete invitation link')).toBeUndefined();
    expect(h.allText()).not.toContain('Could not create invitation');
  } finally { await h.unmount(); client.clear(); }
});

it.each(['copy', 'share', 'cancel'] as const)('keeps %s recovery beside its task and clears it on retry', async action => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); let fail = true;
  const operation = async () => { if (fail) throw new Error('Try this action again'); };
  const repository: InventoryInvitationManagementRepository = {
    list: async () => ({ items: [item] }),
    create: async () => ({ ...item, inviteUrl: 'https://example.test/#token=secret' }), cancel: operation
  };
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><AppFeedbackProvider><InventorySharingScreen scope={scope} listQuery={new ListInventoryInvitationsQuery(repository)} createCommand={new CreateInventoryInvitationCommand(repository)} cancelCommand={new CancelInventoryInvitationCommand(repository)} linkActions={{ copy: operation, share: operation }} /></AppFeedbackProvider></MobileServerStateProvider>);
    await settle(h);
    if (action !== 'cancel') {
      await h.changeText(h.byLabel('Invitee email'), 'friend@example.test');
      await h.press(h.byLabel('Create Invitation')); await settle(h);
    }
    const perform = async () => {
      if (action === 'cancel') {
        await openCancellation(h, 'old@example.test');
        await h.run(() => pressAlertButton('Cancel Invitation'));
      } else {
        const label = action === 'copy' ? 'Copy link' : 'Share invitation';
        await h.press(h.allByType('Pressable').find(node => node.children.some(child => typeof child === 'object' && child !== null && 'children' in child && child.children.includes(label))));
      }
      await settle(h);
    };
    await perform();
    const recovery = h.byText(`Could not ${action} invitation`);
    expect(recovery).toBeDefined();
    let parent = recovery?.parent;
    while (parent && parent.type !== 'ScrollView') parent = parent.parent;
    expect(parent?.type, 'task feedback must scroll with the form, outside the navigation overlay').toBe('ScrollView');
    if (action !== 'cancel') expect(h.byLabel('Complete invitation link')).toBeDefined();
    else expect(h.byLabel('Invitation actions for old@example.test')).toBeDefined();
    fail = false; await perform();
    expect(h.byText(`Could not ${action} invitation`)).toBeUndefined();
    if (action === 'copy') {
      const copied = h.byText('Invitation link copied');
      expect(copied).toBeDefined();
      let ancestor = copied?.parent;
      while (ancestor && ancestor.type !== 'ScrollView') ancestor = ancestor.parent;
      expect(ancestor?.type).toBe('ScrollView');
    }
  } finally { await h.unmount(); client.clear(); }
});

it.each(['success', 'failure'] as const)('does not attach delayed link %s feedback to a replacement invitation', async outcome => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); let finishCopy: (() => void) | undefined;
  const repository: InventoryInvitationManagementRepository = {
    list: async () => ({ items: [] }),
    create: async (_scope, input) => ({ ...item, id: input.email, email: input.email, inviteUrl: `https://example.test/#token=${input.email}` }),
    cancel: async () => undefined
  };
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><AppFeedbackProvider><InventorySharingScreen scope={scope} listQuery={new ListInventoryInvitationsQuery(repository)} createCommand={new CreateInventoryInvitationCommand(repository)} cancelCommand={new CancelInventoryInvitationCommand(repository)} linkActions={{ copy: () => new Promise((resolve, reject) => { finishCopy = () => outcome === 'success' ? resolve() : reject(new Error('First link failed')); }), share: async () => undefined }} /></AppFeedbackProvider></MobileServerStateProvider>);
    await settle(h);
    await h.changeText(h.byLabel('Invitee email'), 'first@example.test');
    await h.press(h.byLabel('Create Invitation')); await settle(h);
    await h.press(h.allByType('Pressable').find(node => node.children.some(child => typeof child === 'object' && child !== null && 'children' in child && child.children.includes('Copy link'))));
    expect(finishCopy).toBeDefined();
    await h.changeText(h.byLabel('Invitee email'), 'second@example.test');
    await h.press(h.byLabel('Create Invitation')); await settle(h);
    const finishOldCopy = finishCopy;
    await h.press(h.byLabel('Copy link'));
    expect(h.byLabel('Copying…')?.props.disabled).toBe(true);
    await h.run(() => finishOldCopy?.());
    expect(h.byLabel('Copying…')?.props.disabled).toBe(true);
    expect(h.byLabel('Share invitation')?.props.disabled).toBe(true);
    expect(h.allText()).not.toContain('Invitation link copied');
    expect(h.allText()).not.toContain('Could not copy invitation');
    expect(h.allText()).not.toContain('First link failed');
    expect(h.byLabel('Complete invitation link')?.children.join('')).toContain('second@example.test');
    await h.run(() => finishCopy?.());
    expect(h.byLabel('Copy link')?.props.disabled).toBe(false);
    expect(h.byLabel('Share invitation')?.props.disabled).toBe(false);
  } finally { await h.unmount(); client.clear(); }
});

it.each(['copy', 'share'] as const)('locks both link commands during %s and recovers after failure', async action => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); let rejectAction: ((error: Error) => void) | undefined; let calls = 0;
  const operation = () => { calls++; return new Promise<void>((_resolve, reject) => { rejectAction = reject; }); };
  const repository: InventoryInvitationManagementRepository = {
    list: async () => ({ items: [] }),
    create: async () => ({ ...item, inviteUrl: 'https://example.test/#token=secret' }), cancel: async () => undefined
  };
  const command = (label: string) => h.allByType('Pressable').find(node => node.props.accessibilityLabel === label || node.children.some(child => typeof child === 'object' && child !== null && 'children' in child && child.children.includes(label)));
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><AppFeedbackProvider><InventorySharingScreen scope={scope} listQuery={new ListInventoryInvitationsQuery(repository)} createCommand={new CreateInventoryInvitationCommand(repository)} cancelCommand={new CancelInventoryInvitationCommand(repository)} linkActions={{ copy: operation, share: operation }} /></AppFeedbackProvider></MobileServerStateProvider>);
    await settle(h); await h.changeText(h.byLabel('Invitee email'), 'friend@example.test');
    await h.press(h.byLabel('Create Invitation')); await settle(h);
    const copy = command('Copy link'); const share = command('Share invitation');
    await h.press(action === 'copy' ? copy : share);
    expect(command(action === 'copy' ? 'Copying…' : 'Sharing…')?.props.disabled).toBe(true);
    expect(command(action === 'copy' ? 'Share invitation' : 'Copy link')?.props.disabled).toBe(true);
    await h.press(copy); await h.press(share);
    expect(calls).toBe(1);
    await h.run(() => rejectAction?.(new Error('Link action unavailable')));
    expect(command('Copy link')?.props.disabled).toBe(false);
    expect(command('Share invitation')?.props.disabled).toBe(false);
    expect(h.byLabel('Complete invitation link')).toBeDefined();
    expect(h.byText(`Could not ${action} invitation`)).toBeDefined();
  } finally { await h.unmount(); client.clear(); }
});

it('keeps each invitation locked independently and rejects duplicate confirmations', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); const calls: string[] = [];
  const finish = new Map<string, (error: Error) => void>();
  const repository: InventoryInvitationManagementRepository = {
    list: async () => ({ items: [item, { ...item, id: 'two', email: 'second@example.test' }] }),
    create: async () => ({ ...item, inviteUrl: 'unused' }),
    cancel: (_scope, id) => { calls.push(id); return new Promise((_resolve, reject) => { finish.set(id, reject); }); }
  };
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><AppFeedbackProvider><InventorySharingScreen scope={scope} listQuery={new ListInventoryInvitationsQuery(repository)} createCommand={new CreateInventoryInvitationCommand(repository)} cancelCommand={new CancelInventoryInvitationCommand(repository)} linkActions={{ copy: async () => undefined, share: async () => undefined }} /></AppFeedbackProvider></MobileServerStateProvider>);
    await settle(h);
    await openCancellation(h, 'old@example.test');
    await h.run(() => pressAlertButton('Cancel Invitation'));
    await h.run(() => pressAlertButton('Cancel Invitation'));
    expect(calls).toEqual(['one']);
    await openCancellation(h, 'second@example.test');
    await h.run(() => pressAlertButton('Cancel Invitation'));
    expect(h.byLabel('Invitation actions for old@example.test')?.props.disabled).toBe(true);
    expect(h.byLabel('Invitation actions for second@example.test')?.props.disabled).toBe(true);
    await h.run(() => finish.get('one')?.(new Error('First cancellation failed')));
    expect(h.byLabel('Invitation actions for old@example.test')?.props.disabled).toBe(false);
    expect(h.byLabel('Invitation actions for second@example.test')?.props.disabled).toBe(true);
    await h.run(() => finish.get('two')?.(new Error('Second cancellation failed')));
    expect(h.byLabel('Invitation actions for second@example.test')?.props.disabled).toBe(false);
    await h.run(() => pressAlertButton('Cancel Invitation'));
    expect(calls).toEqual(['one', 'two']);
  } finally { await h.unmount(); client.clear(); }
});

it.each(['leave', 'return'] as const)('does not start a cancellation from an old confirmation after %s', async departure => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); let cancellations = 0;
  const repository: InventoryInvitationManagementRepository = {
    list: async () => ({ items: [item] }), create: async () => ({ ...item, inviteUrl: 'unused' }),
    cancel: async () => { cancellations++; }
  };
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><AppFeedbackProvider><InventorySharingScreen scope={scope} listQuery={new ListInventoryInvitationsQuery(repository)} createCommand={new CreateInventoryInvitationCommand(repository)} cancelCommand={new CancelInventoryInvitationCommand(repository)} linkActions={{ copy: async () => undefined, share: async () => undefined }} /></AppFeedbackProvider></MobileServerStateProvider>);
    await settle(h); await openCancellation(h, 'old@example.test');
    await h.run(() => setScreenFocused(false));
    if (departure === 'return') await h.run(() => setScreenFocused(true));
    await h.run(() => pressAlertButton('Cancel Invitation'));
    expect(cancellations).toBe(0);
  } finally { await h.unmount(); client.clear(); setScreenFocused(true); }
});
