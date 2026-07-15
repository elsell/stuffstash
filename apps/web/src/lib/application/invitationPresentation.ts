import type { InvitationPreview } from '$lib/domain/invitation';

export type InvitationPresentationState = 'ready' | 'expired' | 'revoked' | 'cancelled' | 'accepted';
export type InvitationScreenState = 'loading' | 'signed_out' | 'invalid' | 'email_mismatch' | 'unavailable' | 'success' | InvitationPresentationState;

export function invitationPresentationState(preview: InvitationPreview): InvitationPresentationState {
  if (preview.status === 'accepted') return 'accepted';
  if (preview.status === 'revoked') return 'revoked';
  if (preview.status === 'cancelled') return 'cancelled';
  return preview.status === 'expired' || preview.isExpired ? 'expired' : 'ready';
}

export function invitationRelationshipLabel(preview: InvitationPreview): string {
  return preview.relationship === 'editor' ? 'Can edit' : 'Can view';
}

export function invitationExpirationLabel(preview: InvitationPreview, locale?: string): string {
  return new Intl.DateTimeFormat(locale, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(preview.expiresAt));
}
