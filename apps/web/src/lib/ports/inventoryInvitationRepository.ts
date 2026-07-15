import type { InvitationAcceptance, InvitationLinkMaterial, InvitationPreview } from '$lib/domain/invitation';

export interface InventoryInvitationRepository {
  preview(material: InvitationLinkMaterial): Promise<InvitationPreview>;
  accept(material: InvitationLinkMaterial): Promise<InvitationAcceptance>;
}
