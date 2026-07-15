import { describe, expect, it } from 'vitest';
import { parseInvitationLink } from './invitationLink';

const origin = 'https://stash.example.test';
const token = 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA';

describe('parseInvitationLink', () => {
  it('parses the canonical scoped invitation link', () => {
    expect(parseInvitationLink(`/invitations/accept?tenant=tenant-one&inventory=inventory-one&invitation=invite-one#token=${token}`, origin)).toEqual({
      tenantId: 'tenant-one',
      inventoryId: 'inventory-one',
      invitationId: 'invite-one',
      token
    });
  });

  it.each([
    `https://evil.example/invitations/accept?tenant=t&inventory=i&invitation=x#token=${token}`,
    `/other?tenant=t&inventory=i&invitation=x#token=${token}`,
    `/invitations/accept?tenant=t&tenant=t2&inventory=i&invitation=x#token=${token}`,
    `/invitations/accept?tenant=t&inventory=i&invitation=x&token=${token}#token=${token}`,
    `/invitations/accept?tenant=t&inventory=i&invitation=x#token=${token}&token=${token}`,
    `/invitations/accept?tenant=t&inventory=i&invitation=x#token=short`,
    `/invitations/accept?tenant=t&inventory=i&invitation=x#token=${'A'.repeat(42)}!`,
    `/invitations/accept?tenant=t%00bad&inventory=i&invitation=x#token=${token}`,
    `/invitations/accept?tenant=t&inventory=i&invitation=x&redirect=https://evil.example#token=${token}`,
    `/invitations/accept?tenant=t&inventory=i&invitation=x#token=${token}&extra=value`
  ])('rejects malformed or ambiguous material: %s', (value) => {
    expect(parseInvitationLink(value, origin)).toBeNull();
  });
});
