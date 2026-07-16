<script lang="ts">
  import type { Snippet } from 'svelte';
  import AuthBrand from './AuthBrand.svelte';

  let {
    title,
    description,
    children
  }: {
    title: string;
    description: string;
    children?: Snippet;
  } = $props();
</script>

<main class="auth-shell">
  <section class="auth-panel" data-auth-track="readable" aria-labelledby="auth-surface-title">
    <AuthBrand />

    <header class="auth-copy">
      <h1 id="auth-surface-title">{title}</h1>
      <p>{description}</p>
    </header>

    {#if children}
      {@render children()}
    {/if}
  </section>
</main>

<style>
  .auth-shell {
    display: grid;
    min-height: 100svh;
    place-items: center;
    padding: var(--space-6);
    background: var(--background);
    color: var(--foreground);
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  }

  .auth-panel {
    display: grid;
    width: 100%;
    max-width: 26rem;
    gap: var(--space-6);
    padding: var(--space-8);
    border: 1px solid var(--border);
    border-radius: var(--radius-surface);
    background: var(--card);
  }

  .auth-copy {
    display: grid;
    gap: var(--space-2);
  }

  h1,
  p {
    margin: 0;
  }

  h1 {
    color: var(--foreground);
    font-size: var(--text-title-size);
    font-weight: 700;
    line-height: var(--text-title-line-height);
  }

  p {
    color: var(--muted-foreground);
    font-size: var(--text-body-size);
    line-height: var(--text-body-line-height);
  }

  @media (max-width: 640px) {
    .auth-shell {
      place-items: center stretch;
      padding: var(--space-4);
    }

    .auth-panel {
      max-width: none;
      padding: var(--space-6);
    }
  }
</style>
