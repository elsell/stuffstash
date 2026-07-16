<script lang="ts">
  import LogIn from '@lucide/svelte/icons/log-in';
  import * as Button from '$lib/components/ui/button/index.js';
  import AuthSurface from './AuthSurface.svelte';

  let {
    title = 'Sign in to Stuff Stash',
    description = 'Continue to your secure sign-in page. You’ll return here when you’re done.',
    error = '',
    canSignIn = true,
    onSignIn
  }: {
    title?: string;
    description?: string;
    error?: string;
    canSignIn?: boolean;
    onSignIn: () => Promise<void> | void;
  } = $props();

  let signingIn = $state(false);

  async function handleSignIn(): Promise<void> {
    if (!canSignIn || signingIn) {
      return;
    }
    signingIn = true;
    try {
      await onSignIn();
    } finally {
      signingIn = false;
    }
  }
</script>

<AuthSurface {title} {description}>
  <div class="auth-actions">
    {#if error}
      <p class="signin-error" role="alert">{error}</p>
    {/if}

    <div>
      <Button.Root class="signin-button" style="min-height: 48px" size="lg" disabled={!canSignIn || signingIn} onclick={() => { void handleSignIn(); }}>
        <LogIn aria-hidden="true" />
        {signingIn ? 'Opening sign-in…' : 'Continue to sign in'}
      </Button.Root>
    </div>
  </div>
</AuthSurface>

<style>
  .auth-actions {
    display: grid;
    gap: var(--space-4);
  }

  .signin-error {
    margin: 0;
    padding: var(--space-3);
    border: 1px solid color-mix(in oklab, var(--destructive) 24%, var(--border));
    border-radius: var(--radius-control);
    background: color-mix(in oklab, var(--destructive) 6%, var(--card));
    color: var(--destructive);
    font-size: var(--text-body-size);
    line-height: var(--text-body-line-height);
  }

  .auth-actions :global([data-slot='button']) {
    width: 100%;
  }

  .auth-actions :global(.signin-button) {
    min-height: 48px;
  }

  .auth-actions :global(.signin-button svg) {
    width: 16px;
    height: 16px;
  }
</style>
