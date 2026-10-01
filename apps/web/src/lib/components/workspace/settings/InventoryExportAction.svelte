<script lang="ts">
  import { t } from '$lib/presentation/localization';
  import { onDestroy } from 'svelte';
  import Download from '@lucide/svelte/icons/download';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import type { ExportInventory } from '$lib/application/exportInventory';
  import type { ExportFormat, ExportScope } from '$lib/ports/inventoryExport';

  let { command, scope }: { command: ExportInventory; scope: ExportScope } = $props();
  let pending = $state(false);
  let message = $state('');
  let error = $state(false);
  let lastFormat: ExportFormat = 'json';
  let active: AbortController | undefined;
  onDestroy(() => active?.abort());
  async function run(format: ExportFormat) {
    if (pending) return;
    const request = new AbortController(); active = request;
    lastFormat = format; pending = true; message = ''; error = false;
    try {
      await command.execute(scope, format, request.signal);
      if (!request.signal.aborted) message = 'Inventory download started.';
    } catch (caught) {
      if (!request.signal.aborted) {
        error = true;
        const status = (caught as { status?: number }).status;
        message = status === 401 ? 'Sign in again to export this inventory.' : status === 403 ? 'You no longer have access to export this inventory.' : status === 422 ? 'This inventory exceeds the server’s export limit. Ask your administrator to increase it.' : 'Could not export this inventory. Try again.';
      }
    } finally { if (active === request) { pending = false; active = undefined; } }
  }
</script>

<section class="settings-resource-group" aria-label={t('web.InventoryExportAction.exportInventory')}>
  <h2>{t('web.InventoryExportAction.exportInventory')}</h2>
  <p>{t('web.InventoryExportAction.downloadThisInventoryIncludingArchivedItemsJSONIncludesField')}</p>
  <div>
    <DropdownMenu.Root>
      <DropdownMenu.Trigger>
        {#snippet child({ props })}
          <Button.Root {...props} variant="outline" disabled={pending}><Download aria-hidden="true" />{pending ? 'Preparing export…' : 'Export inventory'}</Button.Root>
        {/snippet}
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="start">
        <DropdownMenu.Item onSelect={() => void run('json')}>{t('web.InventoryExportAction.jSONCompleteInventoryData')}</DropdownMenu.Item>
        <DropdownMenu.Item onSelect={() => void run('csv')}>{t('web.InventoryExportAction.cSVSpreadsheetRows')}</DropdownMenu.Item>
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  </div>
  {#if message}<p role={error ? 'alert' : 'status'}>{message}</p>{/if}
  {#if error}<div><Button.Root variant="outline" disabled={pending} onclick={() => void run(lastFormat)}>{t('web.InventoryExportAction.retryExport')}</Button.Root></div>{/if}
</section>
