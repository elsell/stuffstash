<script lang="ts">
  import { onMount } from 'svelte';
  import { getStoredSession, signOut, startSignIn, type AuthSession } from '$lib/auth';
  import { loadRuntimeConfig, type RuntimeConfig } from '$lib/runtimeConfig';
  import AuthSignInScreen from '$lib/components/auth/AuthSignInScreen.svelte';
  import InventoryWorkspaceApp from '$lib/components/workspace/InventoryWorkspaceApp.svelte';
  import * as Card from '$lib/components/ui/card/index.js';
  import { InMemoryWorkspaceObserver } from '$lib/observability/workspaceObserver';
  import { BrowserAuthObserver, authFailureAttributes, type AuthObserver } from '$lib/observability/authObserver';
  import type { WorkspaceData } from '$lib/domain/inventory';
  import { hasRecentlyCompletedSignIn } from '$lib/auth';
  import { isAuthenticationRequiredError } from '$lib/application/authenticationRequired';
  import {
    signInFailureMessage,
    signInPresentation,
    type SignInFailure,
    type SignInState
  } from '$lib/application/signInPresentation';
  import { StuffStashInventoryRepository } from '$lib/adapters/api/stuffStashInventoryRepository';
  import type { InventoryAccessRepository } from '$lib/ports/inventoryAccessRepository';
  import type { InventoryAuditRepository } from '$lib/ports/inventoryAuditRepository';
  import type { InventoryCustomizationRepository } from '$lib/ports/inventoryCustomizationRepository';
  import type { InventoryRepository } from '$lib/ports/inventoryRepository';
  import type { InventoryBrowseRepository } from '$lib/ports/inventoryBrowseRepository';

  let config = $state<RuntimeConfig | null>(null);
  let session = $state<AuthSession | null>(null);
  let repository = $state<(InventoryRepository & InventoryBrowseRepository & InventoryAccessRepository & InventoryAuditRepository & InventoryCustomizationRepository) | null>(null);
  let workspaceData = $state<WorkspaceData | null>(null);
  let loading = $state(true);
  let authFailure = $state<SignInFailure | null>(null);
  let workspaceError = $state('');
  let authNotice = $state<Exclude<SignInState, 'default'> | null>(null);
  let authObserver: AuthObserver | null = null;

  onMount(async () => {
    authObserver = new BrowserAuthObserver();
    try {
      config = await loadRuntimeConfig();
      session = getStoredSession();
      if (session) {
        const observer = new InMemoryWorkspaceObserver();
        repository = new StuffStashInventoryRepository(config, () => getStoredSession()?.idToken ?? null, observer);
        workspaceData = await repository.loadWorkspace();
      }
    } catch (caught) {
      if (isAuthenticationRequiredError(caught)) {
        expireSession(caught);
      } else if (session) {
        authObserver.record('auth.workspace_load_failed', authFailureAttributes(caught, 'workspace_transport'));
        workspaceError = signInFailureMessage('workspace');
      } else {
        authObserver.record('auth.runtime_configuration_failed', authFailureAttributes(caught, 'runtime_configuration'));
        authFailure = 'configuration';
      }
    } finally {
      loading = false;
    }
  });

  async function signIn(): Promise<void> {
    if (config) {
      authNotice = null;
      authFailure = null;
      try {
        await startSignIn(config);
      } catch (caught) {
        authObserver?.record('auth.sign_in_start_failed', authFailureAttributes(caught, 'sign_in_navigation'));
        authFailure = 'start';
      }
    }
  }

  function signOutAndReset(): void {
    signOut();
    session = null;
    repository = null;
    workspaceData = null;
    authFailure = null;
    workspaceError = '';
    authNotice = null;
  }

  function expireSession(failure?: unknown): void {
    const notice = hasRecentlyCompletedSignIn() ? 'rejected' : 'expired';
    authObserver?.record(
      'auth.session_invalidated',
      authFailureAttributes(failure, notice === 'rejected' ? 'post_callback_rejected' : 'session_expired')
    );
    signOut();
    session = null;
    repository = null;
    workspaceData = null;
    authFailure = null;
    workspaceError = '';
    authNotice = notice;
  }

  const authPresentation = $derived(signInPresentation(authNotice ?? 'default'));
  const authError = $derived(authFailure ? signInFailureMessage(authFailure) : '');
</script>

<svelte:head>
  <title>Stuff Stash</title>
</svelte:head>

{#if loading}
  <main class="loading-shell">
    <Card.Root>
      <Card.Content>
        <p class="muted">Loading Stuff Stash...</p>
      </Card.Content>
    </Card.Root>
  </main>
{:else if !session}
  <AuthSignInScreen
    title={authPresentation.title}
    description={authPresentation.description}
    error={authError}
    canSignIn={Boolean(config)}
    onSignIn={signIn}
  />
{:else if repository && workspaceData}
  <InventoryWorkspaceApp {repository} initialData={workspaceData} onSignOut={signOutAndReset} onSessionExpired={expireSession} />
{:else if workspaceError}
  <main class="loading-shell">
    <Card.Root>
      <Card.Content>
        <p class="muted" role="alert">{workspaceError}</p>
      </Card.Content>
    </Card.Root>
  </main>
{/if}
