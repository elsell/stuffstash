<script lang="ts">
  import Activity from '@lucide/svelte/icons/activity';
  import Boxes from '@lucide/svelte/icons/boxes';
  import Shield from '@lucide/svelte/icons/shield';
  import SlidersHorizontal from '@lucide/svelte/icons/sliders-horizontal';
  import Users from '@lucide/svelte/icons/users';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import type { Inventory, Tenant } from '$lib/domain/inventory';
  import { canEditAsset, hasAccessPermission } from '$lib/domain/inventory';

  let {
    tenant,
    inventory,
    inventoryCount
  }: {
    tenant: Tenant | null;
    inventory: Inventory | null;
    inventoryCount: number;
  } = $props();

  let canShare = $derived(hasAccessPermission(inventory?.access, 'share'));
  let canConfigureInventory = $derived(hasAccessPermission(inventory?.access, 'configure'));
  let canConfigureTenant = $derived(hasAccessPermission(tenant?.access, 'configure'));
  let canEditAssets = $derived(canEditAsset(inventory));
</script>

<section class="workspace-main" aria-labelledby="settings-title">
  <div class="section-heading">
    <div>
      <h1 id="settings-title">Inventory settings</h1>
      <p>{inventory?.name ?? 'No inventory selected'}</p>
    </div>
    {#if inventory}
      <Badge variant={canConfigureInventory ? 'secondary' : 'outline'}>{inventory.access.relationship}</Badge>
    {/if}
  </div>

  {#if !inventory}
    <div class="empty-state spacious">
      <h2>No inventory selected</h2>
      <p>Select or create an inventory before managing settings.</p>
    </div>
  {:else}
    <div class="settings-grid">
      <section class="settings-panel" aria-labelledby="settings-overview">
        <div class="settings-panel-heading">
          <Boxes aria-hidden="true" />
          <div>
            <h2 id="settings-overview">Overview</h2>
            <p>{tenant?.name ?? 'No tenant'} / {inventory.name}</p>
          </div>
        </div>
        <dl class="detail-list">
          <div><dt>Tenant</dt><dd>{tenant?.name ?? 'Not available'}</dd></div>
          <div><dt>Inventories</dt><dd>{inventoryCount}</dd></div>
          <div><dt>Access</dt><dd>{inventory.access.relationship}</dd></div>
          <div><dt>Asset edits</dt><dd>{canEditAssets ? 'Allowed' : 'View only'}</dd></div>
        </dl>
      </section>

      <section class="settings-panel" aria-labelledby="settings-access">
        <div class="settings-panel-heading">
          <Users aria-hidden="true" />
          <div>
            <h2 id="settings-access">Sharing</h2>
            <p>
              {canShare
                ? 'Direct grants and invitations are planned for the access workflow.'
                : 'Sharing requires inventory share access.'}
            </p>
          </div>
        </div>
        <Button.Root variant="outline" disabled={true}>Manage sharing unavailable</Button.Root>
      </section>

      <section class="settings-panel" aria-labelledby="settings-activity">
        <div class="settings-panel-heading">
          <Activity aria-hidden="true" />
          <div>
            <h2 id="settings-activity">Activity</h2>
            <p>Audit history and undoable operations are not connected yet.</p>
          </div>
        </div>
        <Button.Root variant="outline" disabled={true}>View activity unavailable</Button.Root>
      </section>

      <section class="settings-panel" aria-labelledby="settings-customization">
        <div class="settings-panel-heading">
          <SlidersHorizontal aria-hidden="true" />
          <div>
            <h2 id="settings-customization">Customization</h2>
            <p>
              {canConfigureInventory
                ? 'Custom fields and asset types are planned for configuration.'
                : 'Configuration requires inventory configure access.'}
            </p>
          </div>
        </div>
        <Button.Root variant="outline" disabled={true}>Manage fields unavailable</Button.Root>
      </section>

      <section class="settings-panel" aria-labelledby="settings-admin">
        <div class="settings-panel-heading">
          <Shield aria-hidden="true" />
          <div>
            <h2 id="settings-admin">Administration</h2>
            <p>
              {canConfigureTenant
                ? 'Tenant-level administration is planned for this workspace.'
                : 'Tenant administration is not available for this account.'}
            </p>
          </div>
        </div>
        <Button.Root variant="outline" disabled={true}>Tenant administration unavailable</Button.Root>
      </section>
    </div>
  {/if}
</section>
