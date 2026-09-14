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
