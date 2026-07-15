<script lang="ts">
  import Activity from '@lucide/svelte/icons/activity';
  import { safeWorkspaceErrorMessage } from '$lib/application/workspaceSafeError';
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
  import {
    auditRecordPresentation,
    auditStatusPresentation,
    groupAuditRecordsByDay
  } from '$lib/application/workspaceAuditPresentation';
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
  let visibleRecordCount = $state(20);
  let visibleRecords = $derived(records.slice(0, visibleRecordCount));
  let auditDayGroups = $derived(groupAuditRecordsByDay(visibleRecords));
  let hasBufferedRecords = $derived(visibleRecordCount < records.length);
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
  let auditStatus = $derived(
    auditStatusPresentation({
      hasTenant: !!tenant,
      hasInventory: !!inventory,
      scope,
      canReadScope,
      error,
      busy,
      loaded,
      recordCount: records.length
    })
  );
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
    visibleRecordCount = 20;
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
      if (cursor) {
        visibleRecordCount += 20;
      }
      nextCursor = page.pagination.nextCursor;
      loaded = true;
    } catch (caught) {
      if (caught instanceof Error && caught.name === 'AbortError') {
        return;
      }
      if (current === requestId && contextKey === expectedContext) {
        error = safeWorkspaceErrorMessage(caught, 'Activity could not be loaded. Try again.');
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

  function showMoreBufferedRecords(): void {
    visibleRecordCount += 20;
  }

  function selectScope(nextScope: AuditScope): void {
    scope = nextScope;
    onScopeChange(nextScope);
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

  {#if auditStatus.kind !== 'none'}
    <p class={auditStatus.role === 'alert' ? 'denied-note' : 'muted-note'} role={auditStatus.role}>{auditStatus.message}</p>
  {:else}
    <div class="audit-list" aria-label="Audit records">
      {#each auditDayGroups as group}
        <section class="audit-day-group" aria-labelledby={`audit-day-${group.key}`}>
          <h3 class="audit-day-heading" id={`audit-day-${group.key}`}>{group.label}</h3>
          <div class="audit-day-records">
            {#each group.records as record}
              {@const presented = auditRecordPresentation(record)}
              <article class="audit-row">
                <div data-audit-primary>
                  <strong>{presented.title}</strong>
                  <small class="audit-target">{presented.targetLabel}</small>
                  <small>{presented.occurredAtLabel}</small>
                </div>
                <div class="audit-meta">
                  <Badge variant="outline">{presented.sourceLabel}</Badge>
                  <small>{presented.actorLabel}</small>
                  {#if presented.technicalDetails.length > 0}
                    <details class="audit-technical">
                      <summary>Technical details</summary>
                      <dl>
                        {#each presented.technicalDetails as detail}
                          <div>
                            <dt>{detail.label}</dt>
                            <dd>{detail.value}</dd>
                          </div>
                        {/each}
                      </dl>
                    </details>
                  {/if}
                </div>
              </article>
            {/each}
          </div>
        </section>
      {/each}
    </div>
    {#if hasBufferedRecords}
      <Button.Root variant="outline" disabled={busy} onclick={showMoreBufferedRecords}>
        Show more activity
      </Button.Root>
    {:else if nextCursor}
      <Button.Root variant="outline" size="sm" disabled={busy} onclick={loadNextPage}>
        Load older activity
      </Button.Root>
    {/if}
  {/if}
</section>
