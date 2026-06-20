<script lang="ts">
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { completeSignIn } from '$lib/auth';
  import { loadRuntimeConfig } from '$lib/runtimeConfig';

  let error = '';

  onMount(async () => {
    try {
      const config = await loadRuntimeConfig();
      const returnTo = await completeSignIn(config, window.location.href);
      await goto(returnTo);
    } catch (caught) {
      error = caught instanceof Error ? caught.message : 'Unable to complete sign-in.';
    }
  });
</script>

<main class="callback">
  {#if error}
    <h1>Sign-in failed</h1>
    <p>{error}</p>
    <a href="/">Back to Stuff Stash</a>
  {:else}
    <h1>Signing you in…</h1>
    <p>Hang tight while Dex hands the browser session back to Stuff Stash.</p>
  {/if}
</main>

<style>
  .callback {
    display: grid;
    min-height: 100vh;
    place-content: center;
    gap: 10px;
    padding: 24px;
    text-align: center;
  }

  h1,
  p {
    margin: 0;
  }
</style>
