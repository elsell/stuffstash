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
  import { invitationActionConfirmation, invitationActionIsAvailable } from '$lib/application/workspaceInvitationActions';

  let {
    action,
    invitation,
    busy,
    accessHref,
    panelElement = $bindable(),
    onClose,
    onConfirm
  }: InventoryAccessInvitationActionPanelProps = $props();

  let confirmation = $derived(invitation ? invitationActionConfirmation(action, invitation, busy) : null);
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
        <h3 id="access-invitation-action-title">{confirmation?.title}</h3>
        <p>{invitation.email}</p>
      </div>
    </div>
    <p class="muted-note">{confirmation?.description}</p>
    <div class="heading-actions">
      <Button.Root href={accessHref} variant="outline" onclick={onClose}>Cancel</Button.Root>
      <Button.Root
        variant={confirmation?.destructive ? 'destructive' : 'secondary'}
        disabled={confirmation?.disabled}
        onclick={() => { void onConfirm(action, invitation); }}
      >
        {confirmation?.buttonLabel}
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
