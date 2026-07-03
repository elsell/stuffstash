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
      error = caught instanceof Error ? caught.message : 'Unable to load Stuff Stash.';
    } finally {
      loading = false;
    }
  });

  async function signIn(): Promise<void> {
    if (config) {
      await startSignIn(config);
    }
  }

  function signOutAndReset(): void {
    signOut();
    session = null;
    repository = null;
    workspaceData = null;
    error = '';
  }
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
  <AuthSignInScreen {error} canSignIn={Boolean(config)} onSignIn={signIn} />
{:else if repository && workspaceData}
  <InventoryWorkspaceApp {repository} initialData={workspaceData} onSignOut={signOutAndReset} />
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
