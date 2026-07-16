import { describe, expect, it } from 'vitest';
import type { InvitationPreview } from '$lib/domain/invitation';
import { invitationPresentationState, invitationRelationshipLabel } from './invitationPresentation';

const preview: InvitationPreview = {
  inventoryId: 'inventory-one', inventoryName: 'Workshop tools', relationship: 'viewer', status: 'pending',
  isExpired: false, expiresAt: '2026-07-21T12:00:00Z'
};

describe('invitation presentation', () => {
  it.each([
    [{ ...preview }, 'ready'],
    [{ ...preview, isExpired: true }, 'expired'],
    [{ ...preview, status: 'revoked' as const }, 'revoked'],
    [{ ...preview, status: 'cancelled' as const }, 'cancelled'],
    [{ ...preview, status: 'accepted' as const }, 'accepted']
  ] as const)('maps preview state', (value, expected) => {
    expect(invitationPresentationState(value)).toBe(expected);
  });

  it('uses plain-language relationship labels', () => {
    expect(invitationRelationshipLabel(preview)).toBe('Can view');
    expect(invitationRelationshipLabel({ ...preview, relationship: 'editor' })).toBe('Can edit');
  });
});
