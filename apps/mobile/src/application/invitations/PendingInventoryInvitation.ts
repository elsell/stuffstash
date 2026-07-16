import type { InventoryInvitationReference } from './InventoryInvitationRepository';
import { parseInventoryInvitationLink } from './InvitationLinkParser';

export type PendingInventoryInvitationSnapshot = {
  readonly reference?: InventoryInvitationReference;
  readonly invalid: boolean;
  readonly initialized: boolean;
};

export class PendingInventoryInvitation {
  private snapshot: PendingInventoryInvitationSnapshot = { invalid: false, initialized: false };

  capture(
    source: string,
    configuredPublicOrigin?: string,
    allowInsecureLocalHTTP = false
  ): PendingInventoryInvitationSnapshot {
    try {
      this.snapshot = {
        reference: parseInventoryInvitationLink(
          source,
          configuredPublicOrigin ?? 'https://unconfigured.invalid',
          allowInsecureLocalHTTP
        ),
        invalid: false,
        initialized: true
      };
    } catch {
      this.snapshot = { invalid: true, initialized: true };
    }
    return this.snapshot;
  }

  current(): PendingInventoryInvitationSnapshot {
    return this.snapshot;
  }

  clear(): PendingInventoryInvitationSnapshot {
    this.snapshot = { invalid: false, initialized: true };
    return this.snapshot;
  }
}
