<script lang="ts">
  import { onMount } from 'svelte';
  import { getStoredSession, signOut, startSignIn, type AuthSession } from '$lib/auth';
  import { loadRuntimeConfig, type RuntimeConfig } from '$lib/runtimeConfig';
  import AuthSignInScreen from '$lib/components/auth/AuthSignInScreen.svelte';
  import InventoryWorkspaceApp from '$lib/components/workspace/InventoryWorkspaceApp.svelte';
  import * as Alert from '$lib/components/ui/alert/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import { InMemoryWorkspaceObserver } from '$lib/observability/workspaceObserver';
  import type { WorkspaceData } from '$lib/domain/inventory';
  import { hasRecentlyCompletedSignIn } from '$lib/auth';
  import { isAuthenticationRequiredError } from '$lib/application/authenticationRequired';
  import { StuffStashInventoryRepository } from '$lib/adapters/api/stuffStashInventoryRepository';
  import type { InventoryAccessRepository } from '$lib/ports/inventoryAccessRepository';
  import type { InventoryAuditRepository } from '$lib/ports/inventoryAuditRepository';
  import type { InventoryCustomizationRepository } from '$lib/ports/inventoryCustomizationRepository';
  import type { InventoryRepository } from '$lib/ports/inventoryRepository';

  let config = $state<RuntimeConfig | null>(null);
  let session = $state<AuthSession | null>(null);
  let repository = $state<(InventoryRepository & InventoryAccessRepository & InventoryAuditRepository & InventoryCustomizationRepository) | null>(null);
  let workspaceData = $state<WorkspaceData | null>(null);
  let loading = $state(true);
  let error = $state('');
  let authNotice = $state<'expired' | 'rejected' | null>(null);

  onMount(async () => {
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
        expireSession();
      } else {
        error = caught instanceof Error ? caught.message : 'Unable to load Stuff Stash.';
      }
    } finally {
      loading = false;
    }
  });

  async function signIn(): Promise<void> {
    if (config) {
      authNotice = null;
      await startSignIn(config);
    }
  }

  function signOutAndReset(): void {
    signOut();
    session = null;
    repository = null;
    workspaceData = null;
    error = '';
    authNotice = null;
  }

  function expireSession(): void {
    const notice = hasRecentlyCompletedSignIn() ? 'rejected' : 'expired';
    signOut();
    session = null;
    repository = null;
    workspaceData = null;
    error = '';
    authNotice = notice;
  }

  const authTitle = $derived(authNotice === 'expired' ? 'Session expired.' : authNotice === 'rejected' ? 'Sign-in was rejected.' : 'Sign in to continue.');
  const authDescription = $derived(
    authNotice === 'expired'
      ? 'Sign in again to continue.'
      : authNotice === 'rejected'
        ? 'Dex completed sign-in, but the API rejected the new session. Check that the API accepts this web client ID.'
        : 'Use your configured identity provider.'
  );
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
    title={authTitle}
    description={authDescription}
    {error}
    canSignIn={Boolean(config)}
    onSignIn={signIn}
  />
{:else if repository && workspaceData}
  <InventoryWorkspaceApp {repository} initialData={workspaceData} onSignOut={signOutAndReset} onSessionExpired={expireSession} />
{:else if error}
  <main class="loading-shell">
    <Card.Root>
      <Card.Content>
        <p class="muted">{error}</p>
      </Card.Content>
    </Card.Root>
  </main>
{/if}

{#if error && repository}
  <Alert.Root class="toast" variant="destructive">
    <Alert.Description>{error}</Alert.Description>
  </Alert.Root>
{/if}
