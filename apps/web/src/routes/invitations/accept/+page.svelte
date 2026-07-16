<script lang="ts">
  import { onMount } from 'svelte';
  import { getStoredSession, signOut, startSignIn } from '$lib/auth';
  import { loadRuntimeConfig, type RuntimeConfig } from '$lib/runtimeConfig';
  import { invitationReturnPath, parseInvitationLink } from '$lib/application/invitationLink';
  import { invitationPresentationState, type InvitationScreenState } from '$lib/application/invitationPresentation';
  import { InvitationFailure, type InvitationLinkMaterial, type InvitationPreview } from '$lib/domain/invitation';
  import { StuffStashInventoryInvitationRepository } from '$lib/adapters/api/stuffStashInventoryInvitationRepository';
  import type { InventoryInvitationRepository } from '$lib/ports/inventoryInvitationRepository';
  import InvitationAcceptSurface from '$lib/components/invitations/InvitationAcceptSurface.svelte';

  let screenState = $state<InvitationScreenState>('loading');
  let busy = $state(false);
  let config = $state<RuntimeConfig | null>(null);
  let material = $state<InvitationLinkMaterial | null>(null);
  let preview = $state<InvitationPreview | null>(null);
  let repository = $state<InventoryInvitationRepository | null>(null);
  let inventoryDestination = $state('/');

  const openInventoryHref = $derived(inventoryDestination);

  onMount(() => { void initialize(); });

  async function initialize(): Promise<void> {
    screenState = 'loading';
    try {
      if (!material) {
        material = parseInvitationLink(window.location.href, window.location.origin);
        history.replaceState(history.state, '', window.location.pathname + window.location.search);
      }
      if (!material) {
        screenState = 'invalid';
        return;
      }
      config = await loadRuntimeConfig();
      if (!getStoredSession()) {
        screenState = 'signed_out';
        return;
      }
      repository = new StuffStashInventoryInvitationRepository(config.apiBaseUrl, () => getStoredSession()?.idToken ?? null);
      await loadPreview();
    } catch {
      screenState = 'unavailable';
    }
  }

  async function loadPreview(): Promise<void> {
    if (!repository || !material) return;
    busy = true;
    try {
      preview = await repository.preview(material);
      screenState = invitationPresentationState(preview);
      if (screenState === 'expired' || screenState === 'revoked' || screenState === 'cancelled' || screenState === 'accepted') {
        if (screenState === 'accepted') rememberInventoryDestination(material);
        clearInvitationSecret();
      }
    } catch (error) {
      handleFailure(error);
    } finally {
      busy = false;
    }
  }

  async function acceptInvitation(): Promise<void> {
    if (!repository || !material || !preview || busy) return;
    busy = true;
    try {
      await repository.accept(material);
      rememberInventoryDestination(material);
      clearInvitationSecret();
      screenState = 'success';
      history.replaceState(history.state, '', window.location.pathname + window.location.search);
    } catch (error) {
      handleFailure(error);
    } finally {
      busy = false;
    }
  }

  async function retryInvitation(): Promise<void> {
    if (repository && material) {
      await loadPreview();
      return;
    }
    await initialize();
  }

  async function signIn(): Promise<void> {
    if (!config || !material || busy) return;
    busy = true;
    try {
      const returnTo = invitationReturnPath(material);
      await startSignIn(config, window.location, window.sessionStorage, window.history, returnTo);
      clearInvitationSecret();
    } finally {
      busy = false;
    }
  }

  async function switchAccount(): Promise<void> {
    signOut();
    await signIn();
  }

  function handleFailure(error: unknown): void {
    if (!(error instanceof InvitationFailure)) {
      screenState = 'unavailable';
      return;
    }
    if (error.kind === 'authentication_required') {
      signOut();
      screenState = 'signed_out';
    } else if (error.kind === 'email_mismatch') {
      screenState = 'email_mismatch';
    } else if (error.kind === 'invalid') {
      clearInvitationSecret();
      screenState = 'invalid';
    } else {
      screenState = 'unavailable';
    }
  }

  function rememberInventoryDestination(value: InvitationLinkMaterial): void {
    inventoryDestination = `/tenants/${encodeURIComponent(value.tenantId)}/inventories/${encodeURIComponent(value.inventoryId)}`;
  }

  function clearInvitationSecret(): void {
    if (!material) return;
    material.token = '';
    material = null;
  }
</script>

<InvitationAcceptSurface
  state={screenState}
  {preview}
  {busy}
  onSignIn={signIn}
  onSwitchAccount={switchAccount}
  onAccept={acceptInvitation}
  onRetry={retryInvitation}
  {openInventoryHref}
/>
