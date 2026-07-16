import { describe, expect, it } from 'vitest';
import { PendingInventoryInvitation } from './PendingInventoryInvitation';

const link =
  'https://stash.example/invitations/accept?tenant=tenant-1&inventory=inventory-1&invitation=invite-1#token=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA';

describe('PendingInventoryInvitation', () => {
  it('keeps accepted link material only in the in-memory instance', () => {
    const pending = new PendingInventoryInvitation();

    expect(pending.capture(link, 'https://stash.example')).toEqual({
      reference: {
        tenantId: 'tenant-1',
        inventoryId: 'inventory-1',
        invitationId: 'invite-1',
        acceptanceToken: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA'
      },
      invalid: false,
      initialized: true
    });
    expect(new PendingInventoryInvitation().current()).toEqual({ invalid: false, initialized: false });
  });

  it('replaces old material when an invalid link arrives and clears after use', () => {
    const pending = new PendingInventoryInvitation();
    pending.capture(link, 'https://stash.example');

    expect(pending.capture('https://evil.example/invitations/accept', 'https://stash.example'))
      .toEqual({ invalid: true, initialized: true });
    expect(pending.clear()).toEqual({ invalid: false, initialized: true });
  });
});
