<script lang="ts" module>
  import type { AccessInvitationRouteAction } from '$lib/application/workspaceRoute';
  import type { InventoryAccessInvitation } from '$lib/domain/inventory';

  export type InventoryAccessInvitationActionPanelProps = {
    action: AccessInvitationRouteAction;
    invitation: InventoryAccessInvitation | null;
    busy: boolean;
    error: string;
    accessHref: string;
    onClose: (event: MouseEvent) => void;
    onDismiss: () => void;
    onCloseAutoFocus: (event: Event) => void;
    onConfirm: (action: AccessInvitationRouteAction, invitation: InventoryAccessInvitation) => Promise<boolean>;
  };

  export function invitationActionFocusTarget(
    trigger: HTMLElement | null,
    fallback: HTMLElement | null
  ): HTMLElement | null {
    return trigger?.isConnected ? trigger : fallback;
  }
</script>

<script lang="ts">
  import { onDestroy } from 'svelte';
  import * as Button from '$lib/components/ui/button/index.js';
  import { invitationActionConfirmation, invitationActionIsAvailable } from '$lib/application/workspaceInvitationActions';
  import WorkspaceConfirmationDialog from './action-surface/WorkspaceConfirmationDialog.svelte';

  let {
    action: routeAction,
    invitation,
    busy,
    error,
    accessHref,
    onClose,
    onDismiss,
    onCloseAutoFocus,
    onConfirm
  }: InventoryAccessInvitationActionPanelProps = $props();

  let available = $derived(Boolean(invitation && invitationActionIsAvailable(routeAction, invitation)));
  let confirmation = $derived(invitation ? invitationActionConfirmation(routeAction, invitation, busy) : null);
  let title = $derived(available ? confirmation?.title ?? 'Confirm invitation action' : 'Invitation unavailable');
  let description = $derived(
    available
      ? confirmation?.description ?? ''
      : 'This invitation is not available in the current access list.'
  );
  let open = $state(true);
  let dismissAfterClose = false;
  let dismissTimer: number | null = null;

  onDestroy(() => {
    if (dismissTimer !== null) window.clearTimeout(dismissTimer);
  });

  function requestDismiss(): void {
    dismissAfterClose = true;
    open = false;
  }

  function handleClose(event: MouseEvent): void {
    onClose(event);
    if (event.defaultPrevented) requestDismiss();
  }

  function handleCloseAutoFocus(event: Event): void {
    onCloseAutoFocus(event);
    if (!dismissAfterClose) return;
    dismissAfterClose = false;
    if (dismissTimer !== null) window.clearTimeout(dismissTimer);
    dismissTimer = window.setTimeout(() => {
      dismissTimer = null;
      onDismiss();
    }, 16);
  }

  async function confirm(): Promise<void> {
    if (available && invitation && await onConfirm(routeAction, invitation)) requestDismiss();
  }
</script>

<WorkspaceConfirmationDialog
  {open}
  {title}
  {description}
  {busy}
  onOpenChange={(nextOpen) => { if (!nextOpen) requestDismiss(); }}
  onCloseAutoFocus={handleCloseAutoFocus}
>
  {#snippet children()}
    {#if invitation}<p class="muted-note">{invitation.email}</p>{/if}
    {#if error}<p class="denied-note" role="alert">{error}</p>{/if}
  {/snippet}
  {#snippet cancel()}
    <Button.Root href={accessHref} variant="outline" class="min-h-11" disabled={busy} onclick={handleClose} autofocus>
      {available ? 'Cancel' : 'Back to invitations'}
    </Button.Root>
  {/snippet}
  {#snippet action()}
    {#if available && invitation}
      <Button.Root
        variant={confirmation?.destructive ? 'destructive' : 'secondary'}
        class="min-h-11"
        disabled={confirmation?.disabled}
        onclick={() => { void confirm(); }}
      >
        {confirmation?.buttonLabel}
      </Button.Root>
    {/if}
  {/snippet}
</WorkspaceConfirmationDialog>
