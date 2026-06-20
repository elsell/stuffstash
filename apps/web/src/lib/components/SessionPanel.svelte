<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import type { Principal } from '@stuff-stash/api-client';

  export let principal: Principal | null;
  export let tenantId: string;
  export let busy: boolean;
  export let onLoadIdentity: () => void;
  export let onRefreshInventories: () => void;
</script>

<Card.Root class="identity-panel" size="sm">
  <Card.Header class="p-0">
    <Card.Title>Session</Card.Title>
  </Card.Header>
  <Card.Content class="grid gap-4 p-0">
    {#if principal}
      <dl>
        <div>
          <dt>User</dt>
          <dd>{principal.email || principal.id}</dd>
        </div>
      </dl>
    {:else}
      <Button type="button" onclick={onLoadIdentity} disabled={busy}>Load identity</Button>
    {/if}

    <div class="field-stack">
      <Label for="tenant-id">Tenant ID</Label>
      <Input id="tenant-id" bind:value={tenantId} placeholder="Created tenant ID appears here" />
    </div>
    <Button variant="outline" type="button" onclick={onRefreshInventories} disabled={busy || !tenantId}>
      Refresh inventories
    </Button>
  </Card.Content>
</Card.Root>
