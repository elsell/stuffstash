import type {
  InventoryInvitationAcceptance,
  InventoryInvitationReference,
  InventoryInvitationRepository
} from './InventoryInvitationRepository';

export class AcceptInventoryInvitationCommand {
  constructor(private readonly invitations: InventoryInvitationRepository) {}

  execute(input: InventoryInvitationReference): Promise<InventoryInvitationAcceptance> {
    return this.invitations.accept(input);
  }
}
