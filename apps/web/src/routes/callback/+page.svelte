<script lang="ts">
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { completeSignIn } from '$lib/auth';
  import { loadRuntimeConfig } from '$lib/runtimeConfig';
  import {
    failedSignInCallbackPresentation,
    pendingSignInCallbackPresentation,
    type FailedSignInCallbackPresentation
  } from '$lib/application/signInCallbackPresentation';
  import LoaderCircle from '@lucide/svelte/icons/loader-circle';
  import * as Button from '$lib/components/ui/button/index.js';
  import AuthSurface from '$lib/components/auth/AuthSurface.svelte';
  import { BrowserAuthObserver, authFailureAttributes } from '$lib/observability/authObserver';

  const pending = pendingSignInCallbackPresentation();
  let failure: FailedSignInCallbackPresentation | null = null;

  onMount(async () => {
    const authObserver = new BrowserAuthObserver();
    try {
      const config = await loadRuntimeConfig();
      const returnTo = await completeSignIn(config, window.location.href);
      await goto(returnTo);
    } catch (caught) {
      authObserver.record('auth.callback_failed', authFailureAttributes(caught, 'callback_completion'));
      failure = failedSignInCallbackPresentation(caught);
    }
  });
</script>

<svelte:head>
  <title>{failure ? 'Sign-in failed · Stuff Stash' : 'Signing in · Stuff Stash'}</title>
</svelte:head>

<AuthSurface title={failure?.title ?? pending.title} description={failure?.description ?? pending.description}>
  <div class="callback-status">
    {#if failure}
      <p class="visually-hidden" role="alert">{failure.title} {failure.description}</p>
      <Button.Root href="/" class="callback-action" style="min-height: 48px" size="lg">{failure.actionLabel}</Button.Root>
    {:else}
      <p class="callback-progress" role="status" aria-live="polite">
        <LoaderCircle class="callback-spinner" aria-hidden="true" />
        Confirming session
      </p>
    {/if}
  </div>
</AuthSurface>

<style>
  .callback-status {
    display: grid;
    gap: var(--space-4);
  }

  :global(.callback-action) {
    width: 100%;
    min-height: 48px;
  }

  .callback-progress {
    display: flex;
    min-height: 48px;
    align-items: center;
    gap: var(--space-3);
    margin: 0;
    color: var(--muted-foreground);
    font-size: var(--text-body-size);
    line-height: var(--text-body-line-height);
  }

  :global(.callback-spinner) {
    width: 20px;
    height: 20px;
    animation: callback-spin 1s linear infinite;
    color: var(--primary);
  }

  @keyframes callback-spin {
    to {
      transform: rotate(360deg);
    }
  }

  @media (prefers-reduced-motion: reduce) {
    :global(.callback-spinner) {
      animation: none;
    }
  }
</style>
