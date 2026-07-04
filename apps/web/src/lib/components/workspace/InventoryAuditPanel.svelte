<script lang="ts">
  import Activity from '@lucide/svelte/icons/activity';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import {
    hasAccessPermission,
    type AuditRecord,
    type AuditScope,
    type Inventory,
    type Tenant
  } from '$lib/domain/inventory';
  import type { InventoryAuditRepository } from '$lib/ports/inventoryAuditRepository';
  import { settingsAuditScopeOptions } from '$lib/application/workspaceSettingsNavigation';
  import SegmentedControl from './SegmentedControl.svelte';

  let {
    tenant,
    inventory,
    repository,
    scope = $bindable<AuditScope>('inventory'),
    onScopeChange = (nextScope: AuditScope) => {
      scope = nextScope;
    }
  }: {
    tenant: Tenant | null;
    inventory: Inventory | null;
    repository: InventoryAuditRepository;
    scope?: AuditScope;
    onScopeChange?: (scope: AuditScope) => void;
  } = $props();

  let records = $state<AuditRecord[]>([]);
  let nextCursor = $state<string | null>(null);
  let busy = $state(false);
  let loaded = $state(false);
  let error = $state('');
  let requestId = 0;
  let controller: AbortController | null = null;
  let scopeOptions = $derived(
    settingsAuditScopeOptions({
      tenantId: tenant?.id ?? inventory?.tenantId ?? null,
      inventoryId: inventory?.id ?? null,
      hasTenant: !!tenant,
      hasInventory: !!inventory
    })
  );

  let canReadInventoryAudit = $derived(hasAccessPermission(inventory?.access, 'view'));
  let canReadTenantAudit = $derived(hasAccessPermission(tenant?.access, 'configure'));
  let canReadScope = $derived(scope === 'tenant' ? canReadTenantAudit : canReadInventoryAudit);
  let contextKey = $derived(
    tenant && (scope === 'tenant' || inventory)
      ? `${tenant.id}:${scope === 'tenant' ? 'tenant' : inventory?.id}:${scope}:${canReadScope}`
      : ''
  );

  $effect(() => {
    requestId += 1;
    controller?.abort();
    records = [];
    nextCursor = null;
    loaded = false;
    error = '';
    if (!contextKey || !canReadScope) {
      controller = null;
      return;
    }
    controller = new AbortController();
    void loadRecords(contextKey, undefined, controller.signal);
    return () => {
      controller?.abort();
      controller = null;
    };
  });

  async function loadRecords(expectedContext = contextKey, cursor?: string, signal?: AbortSignal): Promise<void> {
    if (!tenant || (scope === 'inventory' && !inventory) || !canReadScope) {
      return;
    }
    const current = requestId;
    const tenantId = tenant.id;
    const inventoryId = inventory?.id ?? '';
    const requestedScope = scope;
    busy = true;
    error = '';
    try {
      const page =
        requestedScope === 'tenant'
          ? await repository.listTenantAuditRecords(tenantId, cursor, signal)
          : await repository.listInventoryAuditRecords(tenantId, inventoryId, cursor, signal);
      if (current !== requestId || contextKey !== expectedContext) {
        return;
      }
      records = cursor ? [...records, ...page.items] : page.items;
      nextCursor = page.pagination.nextCursor;
      loaded = true;
    } catch (caught) {
      if (caught instanceof Error && caught.name === 'AbortError') {
        return;
      }
      if (current === requestId && contextKey === expectedContext) {
        error = caught instanceof Error ? caught.message : 'Unable to load audit history.';
      }
    } finally {
      if (controller?.signal === signal) {
        controller = null;
      }
      if (current === requestId && contextKey === expectedContext) {
        busy = false;
      }
    }
  }

  function loadNextPage(): void {
    if (!nextCursor || busy) {
      return;
    }
    controller?.abort();
    controller = new AbortController();
    void loadRecords(contextKey, nextCursor, controller.signal);
  }

  function selectScope(nextScope: AuditScope): void {
    scope = nextScope;
    onScopeChange(nextScope);
  }

  function formatDate(value: string): string {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return value;
    }
    return date.toLocaleString();
  }
</script>

<section class="settings-panel wide" aria-labelledby="settings-activity">
  <div class="settings-panel-heading">
    <Activity aria-hidden="true" />
    <div>
      <h2 id="settings-activity">Activity</h2>
      <p>Audit history for authorized tenant and inventory activity.</p>
    </div>
  </div>

  <SegmentedControl
    label="Audit scope"
    value={scope}
    options={scopeOptions}
    onSelect={(value) => selectScope(value as AuditScope)}
  />

  {#if !tenant || (scope === 'inventory' && !inventory)}
    <p class="denied-note">Select an inventory before viewing audit history.</p>
  {:else if !canReadScope}
    <p class="denied-note" role="alert">
      {scope === 'tenant' ? 'Tenant audit history requires tenant configuration access.' : 'Inventory audit history requires inventory view access.'}
    </p>
  {:else if error}
    <p class="denied-note" role="alert">{error}</p>
  {:else if busy && !loaded}
    <p class="muted-note">Loading audit history...</p>
  {:else if records.length === 0}
    <p class="muted-note">No audit records found.</p>
  {:else}
    <div class="audit-list" aria-label="Audit records">
      {#each records as record}
        <article class="audit-row">
          <div>
            <strong>{record.action}</strong>
            <small>{record.targetType} / {record.targetId}</small>
            <small>{formatDate(record.occurredAt)}</small>
          </div>
          <div class="audit-meta">
            <Badge variant="outline">{record.source}</Badge>
            <small>{record.principalId}</small>
            {#if record.requestId}
              <small>{record.requestId}</small>
            {/if}
          </div>
        </article>
      {/each}
    </div>
    {#if nextCursor}
      <Button.Root variant="outline" size="sm" disabled={busy} onclick={loadNextPage}>
        Load more history
      </Button.Root>
    {/if}
  {/if}
</section>
