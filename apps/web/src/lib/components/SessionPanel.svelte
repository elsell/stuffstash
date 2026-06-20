<script lang="ts">
  import type { Principal } from '@stuff-stash/api-client';

  export let principal: Principal | null;
  export let tenantId: string;
  export let busy: boolean;
  export let onLoadIdentity: () => void;
  export let onRefreshInventories: () => void;
</script>

<aside class="panel identity-panel">
  <h2>Session</h2>
  {#if principal}
    <dl>
      <div>
        <dt>User</dt>
        <dd>{principal.email || principal.id}</dd>
      </div>
    </dl>
  {:else}
    <button type="button" onclick={onLoadIdentity} disabled={busy}>Load identity</button>
  {/if}

  <label>
    Tenant ID
    <input bind:value={tenantId} placeholder="Created tenant ID appears here" />
  </label>
  <button class="secondary" type="button" onclick={onRefreshInventories} disabled={busy || !tenantId}>
    Refresh inventories
  </button>
</aside>
