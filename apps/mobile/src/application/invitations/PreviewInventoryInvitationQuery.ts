import type {
  InventoryInvitationPreview,
  InventoryInvitationReference,
  InventoryInvitationRepository
} from './InventoryInvitationRepository';

export class PreviewInventoryInvitationQuery {
  constructor(private readonly invitations: InventoryInvitationRepository) {}

  execute(input: InventoryInvitationReference): Promise<InventoryInvitationPreview> {
    return this.invitations.preview(input);
  }
}
