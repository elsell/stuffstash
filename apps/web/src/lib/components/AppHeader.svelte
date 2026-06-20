<script lang="ts">
  import type { AuthSession } from '$lib/auth';

  export let session: AuthSession | null;
  export let onSignOut: () => void;
</script>

<header class="topbar">
  <div class="brand-lockup">
    <div class="brand-mark" aria-hidden="true">
      <span></span>
    </div>
    <div>
      <h1>Stuff Stash</h1>
      <p>Keep track of what you own and where it lives.</p>
    </div>
  </div>
  {#if session}
    <button class="secondary" type="button" onclick={onSignOut}>Sign out</button>
  {/if}
</header>
