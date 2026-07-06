<script lang="ts">
  import LogIn from '@lucide/svelte/icons/log-in';
  import * as Button from '$lib/components/ui/button/index.js';

  let {
    title = 'Sign in to continue.',
    description = 'Use your configured identity provider.',
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

<main class="auth-shell">
  <section class="signin-panel" aria-labelledby="signin-title">
    <div class="brand-row">
      <span class="brand-mark" aria-hidden="true">
        <span></span>
      </span>
      <span class="brand-name">Stuff Stash</span>
    </div>

    <div class="signin-copy">
      <h1 id="signin-title">{title}</h1>
      <p>{description}</p>
    </div>

    {#if error}
      <p class="signin-error" role="alert">{error}</p>
    {/if}

    <div class="signin-action">
      <Button.Root class="signin-button" size="lg" disabled={!canSignIn || signingIn} onclick={() => { void handleSignIn(); }}>
        <LogIn aria-hidden="true" />
        {signingIn ? 'Opening sign-in...' : 'Sign in'}
      </Button.Root>
    </div>
  </section>
</main>

<style>
  .auth-shell {
    min-height: 100vh;
    display: grid;
    place-items: center;
    padding: clamp(24px, 5vw, 56px);
    background:
      linear-gradient(180deg, rgba(107, 144, 170, 0.08), rgba(255, 255, 255, 0) 36%),
      #f8fafc;
    color: #111827;
  }

  .signin-panel {
    width: min(100%, 380px);
    display: grid;
    gap: 24px;
    padding: 28px;
    border: 1px solid rgba(48, 58, 65, 0.16);
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.92);
    box-shadow: 0 24px 60px rgba(15, 23, 42, 0.10);
  }

  .brand-row {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .brand-mark {
    width: 34px;
    height: 34px;
    display: grid;
    place-items: center;
    border-radius: 8px;
    background: #303a41;
  }

  .brand-mark span {
    width: 15px;
    height: 15px;
    border-radius: 4px;
    background: #f5ab4b;
    box-shadow: 8px -6px 0 #6b90aa;
  }

  .brand-name {
    font-size: 0.9rem;
    font-weight: 700;
  }

  .signin-copy {
    display: grid;
    gap: 8px;
  }

  h1,
  p {
    margin: 0;
  }

  h1 {
    font-size: 1.7rem;
    line-height: 1.12;
    font-weight: 760;
  }

  .signin-copy p {
    color: #53616d;
    line-height: 1.5;
  }

  .signin-error {
    padding: 10px 12px;
    border: 1px solid rgba(185, 28, 28, 0.22);
    border-radius: 8px;
    background: #fef2f2;
    color: #991b1b;
    font-size: 0.9rem;
    line-height: 1.45;
  }

  .signin-action :global([data-slot='button']) {
    width: 100%;
  }

  .signin-action :global(.signin-button) {
    display: inline-flex;
    justify-content: center;
    align-items: center;
    gap: 8px;
    text-align: center;
  }

  .signin-action :global(.signin-button svg) {
    width: 16px;
    height: 16px;
  }

  @media (max-width: 520px) {
    .auth-shell {
      align-items: stretch;
      place-items: center stretch;
      padding: 18px;
    }

    .signin-panel {
      padding: 22px;
    }

    h1 {
      font-size: 1.45rem;
    }
  }
</style>
