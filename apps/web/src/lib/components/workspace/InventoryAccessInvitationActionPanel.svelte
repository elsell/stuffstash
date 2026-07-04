<script lang="ts" module>
  import type { AccessInvitationRouteAction } from '$lib/application/workspaceRoute';
  import type { InventoryAccessInvitation } from '$lib/domain/inventory';

  export type InventoryAccessInvitationActionPanelProps = {
    action: AccessInvitationRouteAction;
    invitation: InventoryAccessInvitation | null;
    busy: boolean;
    accessHref: string;
    panelElement: HTMLElement | null;
    onClose: (event: MouseEvent) => void;
    onConfirm: (action: AccessInvitationRouteAction, invitation: InventoryAccessInvitation) => Promise<void>;
  };
</script>

<script lang="ts">
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import * as Button from '$lib/components/ui/button/index.js';
  import { invitationActionIsAvailable } from './invitationActionPolicy';

  let {
    action,
    invitation,
    busy,
    accessHref,
    panelElement = $bindable(),
    onClose,
    onConfirm
  }: InventoryAccessInvitationActionPanelProps = $props();

  function title(nextAction: AccessInvitationRouteAction): string {
    if (nextAction === 'expire') {
      return 'Expire invitation';
    }
    if (nextAction === 'cancel') {
      return 'Cancel invitation';
    }
    if (nextAction === 'delete') {
      return 'Delete invitation';
    }
    return 'Invitation action';
  }

  function description(nextAction: AccessInvitationRouteAction, target: InventoryAccessInvitation): string {
    if (nextAction === 'expire') {
      return `Set the invitation for ${target.email} to expire immediately.`;
    }
    if (nextAction === 'cancel') {
      return `Cancel the pending invitation for ${target.email}.`;
    }
    if (nextAction === 'delete') {
      return `Permanently remove the invitation record for ${target.email}.`;
    }
    return 'This invitation action is unavailable.';
  }

  function disabled(nextAction: AccessInvitationRouteAction, target: InventoryAccessInvitation): boolean {
    return busy || !invitationActionIsAvailable(nextAction, target);
  }

  function buttonLabel(nextAction: AccessInvitationRouteAction): string {
    if (nextAction === 'expire') {
      return 'Expire';
    }
    if (nextAction === 'cancel') {
      return 'Cancel invitation';
    }
    if (nextAction === 'delete') {
      return 'Delete';
    }
    return 'Continue';
  }

</script>

<section
  bind:this={panelElement}
  class="settings-panel archive-confirmation"
  aria-labelledby="access-invitation-action-title"
  tabindex="-1"
>
  {#if invitation && invitationActionIsAvailable(action, invitation)}
    <div class="settings-panel-heading">
      <Trash2 aria-hidden="true" />
      <div>
        <h3 id="access-invitation-action-title">{title(action)}</h3>
        <p>{invitation.email}</p>
      </div>
    </div>
    <p class="muted-note">{description(action, invitation)}</p>
    <div class="heading-actions">
      <Button.Root href={accessHref} variant="outline" onclick={onClose}>Cancel</Button.Root>
      <Button.Root
        variant={action === 'delete' ? 'destructive' : 'secondary'}
        disabled={disabled(action, invitation)}
        onclick={() => { void onConfirm(action, invitation); }}
      >
        {buttonLabel(action)}
      </Button.Root>
    </div>
  {:else}
    <div class="settings-panel-heading">
      <Trash2 aria-hidden="true" />
      <div>
        <h3 id="access-invitation-action-title">Invitation unavailable</h3>
        <p>This invitation is not available in the current access list.</p>
      </div>
    </div>
    <Button.Root href={accessHref} variant="outline" onclick={onClose}>Back to invitations</Button.Root>
  {/if}
</section>
