<script lang="ts">
  import { t } from '$lib/presentation/localization';
  import Check from '@lucide/svelte/icons/check';
  import LogIn from '@lucide/svelte/icons/log-in';
  import Users from '@lucide/svelte/icons/users';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import AuthBrand from '$lib/components/auth/AuthBrand.svelte';
  import type { InvitationPreview } from '$lib/domain/invitation';
  import { invitationExpirationLabel, invitationRelationshipLabel, type InvitationScreenState } from '$lib/application/invitationPresentation';

  let {
    state,
    preview = null,
    busy = false,
    onSignIn,
    onSwitchAccount,
    onAccept,
    onRetry,
    openInventoryHref = '/'
  }: {
    state: InvitationScreenState;
    preview?: InvitationPreview | null;
    busy?: boolean;
    onSignIn?: () => Promise<void> | void;
    onSwitchAccount?: () => Promise<void> | void;
    onAccept?: () => Promise<void> | void;
    onRetry?: () => Promise<void> | void;
    openInventoryHref?: string;
  } = $props();
</script>

<svelte:head>
  <title>{t('web.InvitationAcceptSurface.inventoryInvitationStuffStash')}</title>
</svelte:head>

<main class="invitation-shell">
  <Card.Root class="invitation-card" aria-live="polite" aria-busy={state === 'loading' || busy}>
    <Card.Header class="invitation-header">
      <AuthBrand />
      <div class:success-mark={state === 'success'} class="invitation-mark">
        {#if state === 'success'}<Check aria-hidden="true" />{:else}<Users aria-hidden="true" />{/if}
      </div>
      <div class="heading-copy">
        <Card.Title role="heading" aria-level={1}>
          {#if state === 'loading'}{t('web.InvitationAcceptSurface.checkingInvitation')} {:else if state === 'signed_out'}{t('web.InvitationAcceptSurface.youVeBeenInvited')} {:else if state === 'ready' && preview}{t('web.InvitationAcceptSurface.joinFull', { inventoryName: preview.inventoryName })} {:else if state === 'accepted' && preview}{t('web.InvitationAcceptSurface.youAlreadyJoinedFull', { inventoryName: preview.inventoryName })} {:else if state === 'success' && preview}{t('web.InvitationAcceptSurface.youJoinedFull', { inventoryName: preview.inventoryName })} {:else if state === 'expired'}{t('web.InvitationAcceptSurface.thisInvitationExpired')} {:else if state === 'revoked'}{t('web.InvitationAcceptSurface.thisInvitationWasRevoked')} {:else if state === 'cancelled'}{t('web.InvitationAcceptSurface.thisInvitationWasCancelled')} {:else if state === 'email_mismatch'}{t('web.InvitationAcceptSurface.thisInvitationIsForAnotherAccount')} {:else if state === 'unavailable'}{t('web.InvitationAcceptSurface.invitationCouldNotBeChecked')} {:else}{t('web.InvitationAcceptSurface.thisInvitationLinkIsInvalid')}{/if}
        </Card.Title>
        <Card.Description>
          {#if state === 'loading'}{t('web.InvitationAcceptSurface.thisWillOnlyTakeAMoment')} {:else if state === 'signed_out'}{t('web.InvitationAcceptSurface.signInToViewTheInventoryAndAccessLevel')} {:else if state === 'ready'}{t('web.InvitationAcceptSurface.reviewTheDetailsThenAcceptWhenYouReReady')} {:else if state === 'accepted' || state === 'success'}{t('web.InvitationAcceptSurface.youCanOpenTheInventoryNow')} {:else if state === 'expired'}{t('web.InvitationAcceptSurface.askTheInventoryOwnerToSendANewInvitation')} {:else if state === 'revoked'}{t('web.InvitationAcceptSurface.theInventoryOwnerRevokedThisInvitationAskThemTo')} {:else if state === 'cancelled'}{t('web.InvitationAcceptSurface.theInventoryOwnerCancelledThisInvitationAskThemTo')} {:else if state === 'email_mismatch'}{t('web.InvitationAcceptSurface.signOutThenUseTheAccountThatReceivedThe')} {:else if state === 'unavailable'}{t('web.InvitationAcceptSurface.yourAccessHasNotChangedCheckYourConnectionAnd')} {:else}{t('web.InvitationAcceptSurface.checkThatYouOpenedTheCompleteLinkOrAsk')}{/if}
        </Card.Description>
      </div>
    </Card.Header>

    {#if preview && ['ready', 'accepted', 'success'].includes(state)}
      <Card.Content>
        <dl class="invitation-details">
          <div><dt>{t('web.InvitationAcceptSurface.inventory')}</dt><dd>{preview.inventoryName}</dd></div>
          <div><dt>{t('web.InvitationAcceptSurface.access')}</dt><dd><span class="access-pill">{invitationRelationshipLabel(preview)}</span></dd></div>
          <div><dt>{t('web.InvitationAcceptSurface.invitationExpires')}</dt><dd>{invitationExpirationLabel(preview)}</dd></div>
        </dl>
      </Card.Content>
    {/if}

    {#if state === 'signed_out' || state === 'email_mismatch' || state === 'ready' || state === 'unavailable' || state === 'accepted' || state === 'success'}
      <Card.Footer class="invitation-actions">
        {#if state === 'signed_out'}
          <Button.Root size="lg" disabled={busy} onclick={() => { void onSignIn?.(); }}><LogIn aria-hidden="true" />{busy ? 'Opening sign-in…' : 'Continue to sign in'}</Button.Root>
        {:else if state === 'email_mismatch'}
          <Button.Root variant="outline" size="lg" disabled={busy} onclick={() => { void onSwitchAccount?.(); }}><LogIn aria-hidden="true" />{busy ? 'Opening sign-in…' : 'Switch account'}</Button.Root>
        {:else if state === 'ready'}
          <Button.Root size="lg" disabled={busy} onclick={() => { void onAccept?.(); }}>{busy ? 'Accepting…' : 'Accept invitation'}</Button.Root>
        {:else if state === 'unavailable'}
          <Button.Root variant="outline" size="lg" disabled={busy} onclick={() => { void onRetry?.(); }}>{busy ? 'Checking…' : 'Try again'}</Button.Root>
        {:else}
          <Button.Root href={openInventoryHref} size="lg">{t('web.InvitationAcceptSurface.openInventory')}</Button.Root>
        {/if}
      </Card.Footer>
    {/if}
  </Card.Root>
</main>

<style>
  .invitation-shell { display: grid; min-height: 100svh; place-items: center; padding: 24px; background: var(--background); }
  :global(.invitation-card) { width: 100%; max-width: 30rem; border-radius: var(--radius-overlay); box-shadow: var(--shadow-panel); }
  :global(.invitation-header) { gap: 20px; }
  .invitation-mark { display: grid; width: 48px; height: 48px; place-items: center; border-radius: var(--radius-pill); background: var(--accent); color: var(--primary); }
  .invitation-mark.success-mark { background: color-mix(in oklab, var(--color-brand-contained) 18%, var(--card)); color: var(--color-brand-frame); }
  .invitation-mark :global(svg) { width: 22px; height: 22px; }
  .heading-copy { display: grid; gap: 8px; }
  .heading-copy :global([data-slot='card-title']) { font-size: var(--text-title-size); line-height: var(--text-title-line-height); }
  .heading-copy :global([data-slot='card-description']) { font-size: var(--text-body-size); line-height: var(--text-body-line-height); }
  .invitation-details { display: grid; margin: 0; border-top: 1px solid var(--border); }
  .invitation-details div { display: grid; grid-template-columns: minmax(7rem, 0.8fr) minmax(0, 1.2fr); gap: 16px; align-items: center; min-height: 52px; padding: 10px 0; border-bottom: 1px solid var(--border); }
  .invitation-details dt { color: var(--muted-foreground); font-size: var(--text-metadata-size); line-height: var(--text-metadata-line-height); }
  .invitation-details dd { margin: 0; color: var(--foreground); font-weight: 600; overflow-wrap: anywhere; }
  .access-pill { display: inline-flex; min-height: 24px; align-items: center; padding: 2px 10px; border-radius: var(--radius-pill); background: var(--secondary); color: var(--secondary-foreground); font-size: var(--text-caption-size); line-height: var(--text-caption-line-height); font-weight: 600; }
  :global(.invitation-actions [data-slot='button']) { width: 100%; min-height: 48px; }
  @media (max-width: 640px) { .invitation-shell { place-items: center stretch; padding: 16px; } :global(.invitation-card) { max-width: none; } .invitation-details div { grid-template-columns: 1fr; gap: 2px; align-items: start; } }
</style>
