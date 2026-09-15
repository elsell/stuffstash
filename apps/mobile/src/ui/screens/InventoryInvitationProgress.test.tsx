import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { InventoryInvitationScreen } from './InventoryInvitationScreen';
import { AcceptInventoryInvitationCommand } from '../../application/invitations/AcceptInventoryInvitationCommand';
import { PreviewInventoryInvitationQuery } from '../../application/invitations/PreviewInventoryInvitationQuery';
import type { InventoryInvitationAcceptance, InventoryInvitationRepository } from '../../application/invitations/InventoryInvitationRepository';

it('keeps joining and opening named during progress and permits opening recovery', async () => {
  const h = new MobileRenderHarness();
  let finishJoin: ((value: InventoryInvitationAcceptance) => void) | undefined;
  let failOpen: ((error: Error) => void) | undefined;
  const reference = { tenantId: 'tenant', inventoryId: 'inventory', invitationId: 'invite', acceptanceToken: 'token' };
  const repository: InventoryInvitationRepository = {
    preview: async () => ({ inventoryId: 'inventory', inventoryName: 'Kitchen', relationship: 'viewer', status: 'pending', isExpired: false, expiresAt: '2027-01-01' }),
    accept: async () => new Promise(resolve => { finishJoin = resolve; })
  };
  try {
    await h.render(<InventoryInvitationScreen initialized invalidLink={false} reference={reference}
      previewQuery={new PreviewInventoryInvitationQuery(repository)} acceptCommand={new AcceptInventoryInvitationCommand(repository)}
      onAccepted={async () => new Promise((_resolve, reject) => { failOpen = reject; })} onDismiss={() => {}} onSwitchAccount={() => {}} />);
    await h.press(h.byLabel('Join inventory'));
    expect(h.byLabel('Join inventory')?.props.accessibilityState).toEqual({ busy: true, disabled: true });
    expect(h.allText()).toContain('Joining…');
    await h.run(() => finishJoin?.({ ...reference, principalId: 'principal', relationship: 'viewer', status: 'accepted' }));
    await h.press(h.byLabel('Open inventory'));
    expect(h.byLabel('Open inventory')?.props.accessibilityState).toEqual({ busy: true, disabled: true });
    expect(h.allText()).toContain('Opening…');
    await h.run(() => failOpen?.(new Error('offline')));
    expect(h.byLabel('Open inventory')?.props.accessibilityState).toEqual({ busy: false, disabled: false });
    expect(h.allText()).toContain('Try opening again');
    expect(h.allText()).toContain('You now have access to ');
  } finally { await h.unmount(); }
});

it.each(['opening', 'start-over'] as const)('keeps a new invitation visible after an old %s failure', async operation => {
  const h = new MobileRenderHarness();
  let fail: ((error: Error) => void) | undefined;
  const reference = { tenantId: 'tenant', inventoryId: 'old', invitationId: 'old', acceptanceToken: 'token' };
  const repository: InventoryInvitationRepository = {
    preview: async selected => ({ inventoryId: selected.inventoryId, inventoryName: selected.inventoryId, relationship: 'viewer', status: operation === 'opening' && selected.inventoryId === 'old' ? 'accepted' : 'pending', isExpired: false, expiresAt: '2027-01-01' }),
    accept: async () => { throw new Error('not used'); }
  };
  const delayed = async () => new Promise<void>((_resolve, reject) => { fail = reject; });
  const view = (selected: typeof reference) => <InventoryInvitationScreen initialized invalidLink={false} reference={selected}
    previewQuery={new PreviewInventoryInvitationQuery(repository)} acceptCommand={new AcceptInventoryInvitationCommand(repository)}
    onAccepted={delayed} onStartOver={delayed} onDismiss={() => {}} onSwitchAccount={() => {}} />;
  try {
    await h.render(view(reference));
    await h.press(h.byLabel(operation === 'opening' ? 'Open inventory' : 'Sign out and start over'));
    await h.render(view({ ...reference, inventoryId: 'new', invitationId: 'new' }));
    expect(h.allText()).toContain('new');
    expect(h.byLabel('Join inventory')?.props.disabled).toBe(false);
    await h.run(() => fail?.(new Error('old operation failed')));
    expect(h.allText()).toContain('new');
    expect(h.allText()).not.toContain('Could not start over');
    expect(h.allText()).not.toContain('Try opening again');
  } finally { await h.unmount(); }
});
