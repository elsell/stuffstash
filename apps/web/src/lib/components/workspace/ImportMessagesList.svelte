<script lang="ts">
  import AlertCircle from '@lucide/svelte/icons/alert-circle';
  import CheckCircle2 from '@lucide/svelte/icons/check-circle-2';
  import type { ImportMessage } from '$lib/domain/inventory';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Button from '$lib/components/ui/button/index.js';
  import { uniqueImportMessages } from './importWorkspacePresentation';

  const COLLAPSED_GROUP_LIMIT = 5;
  const EXPANDED_GROUP_LIMIT = 15;
  const COLLAPSED_RECORD_LIMIT = 3;

  type Props = {
    messages: ImportMessage[];
    emptyText: string;
    truncated?: boolean;
    truncatedText?: string;
    reportedWarnings?: number;
    reportedErrors?: number;
  };

  let {
    messages,
    emptyText,
    truncated = false,
    truncatedText = 'Showing a sample of import messages.',
    reportedWarnings,
    reportedErrors
  }: Props = $props();

  let visibleMessages = $derived(uniqueImportMessages(messages));
  let groups = $derived(groupMessages(visibleMessages));
  let visibleWarningCount = $derived(visibleMessages.filter((message) => message.severity === 'warning').length);
  let visibleErrorCount = $derived(visibleMessages.filter((message) => message.severity === 'error').length);
  let warningCount = $derived(Math.max(reportedWarnings ?? 0, visibleWarningCount));
  let errorCount = $derived(Math.max(reportedErrors ?? 0, visibleErrorCount));
  let expanded = $state(false);
  let expandedGroupKeys = $state<string[]>([]);
  let visibleGroupLimit = $derived(expanded ? EXPANDED_GROUP_LIMIT : COLLAPSED_GROUP_LIMIT);
  let visibleGroups = $derived(groups.slice(0, visibleGroupLimit));
  let hiddenGroupCount = $derived(Math.max(0, groups.length - visibleGroups.length));
  let shouldBoundGroups = $derived(visibleMessages.length > 12 || groups.length > COLLAPSED_GROUP_LIMIT);

  type MessageGroup = {
    key: string;
    severity: ImportMessage['severity'];
    summary: string;
    cause: string;
    messages: ImportMessage[];
  };

  function groupMessages(items: ImportMessage[]): MessageGroup[] {
    const grouped = new Map<string, MessageGroup>();
    for (const message of items) {
      const cause = friendlyCause(message);
      const key = `${message.severity}:${message.summary}:${cause}`;
      const group = grouped.get(key);
      if (group) {
        group.messages.push(message);
      } else {
        grouped.set(key, {
          key,
          severity: message.severity,
          summary: message.summary,
          cause,
          messages: [message]
        });
      }
    }
    return Array.from(grouped.values());
  }

  function severityLabel(severity: ImportMessage['severity']): string {
    return severity === 'error' ? 'Blocking' : 'Warning';
  }

  function groupCountLabel(count: number): string {
    return count === 1 ? '1 item' : `${count} items`;
  }

  function friendlyCause(message: ImportMessage): string {
    const detail = message.detail || '';
    if (message.code === 'duplicate-asset' || detail.toLowerCase().includes('homebox-source-id')) {
      return 'Already linked to an earlier import';
    }
    if (message.code === 'source-link-duplicate') {
      return 'Already imported from this source';
    }
    if (message.code === 'attachment-unavailable') {
      return 'Could not download from the source';
    }
    if (detail.toLowerCase().includes('import validation failed')) {
      return 'File did not pass attachment validation';
    }
    return detail;
  }

  function messageRowLabel(message: ImportMessage, group: MessageGroup): string {
    if (message.sourceName) return message.sourceName;
    if (message.sourceId) return 'Homebox record';
    return message.detail || group.summary;
  }

  function messageDiagnostic(message: ImportMessage, group: MessageGroup): string {
    if (message.sourceName && message.detail && friendlyCause(message) !== group.cause) return friendlyCause(message);
    if (message.sourceId) return `Source ID ${message.sourceId}`;
    return '';
  }

  function groupExpanded(group: MessageGroup): boolean {
    return expandedGroupKeys.includes(group.key);
  }

  function toggleGroup(group: MessageGroup): void {
    expandedGroupKeys = groupExpanded(group)
      ? expandedGroupKeys.filter((key) => key !== group.key)
      : [...expandedGroupKeys, group.key];
  }
</script>

<div class="message-list">
  {#if groups.length > 0}
    <div class="message-list-summary">
      <div class="issue-stat">
        <span>Groups</span>
        <strong>{groups.length}</strong>
      </div>
      <div class="issue-stat">
        <span>Affected</span>
        <strong>{visibleMessages.length}</strong>
      </div>
      {#if errorCount > 0}
        <div class="issue-stat blocking">
          <span>Blocking</span>
          <strong>{errorCount}</strong>
        </div>
      {/if}
      {#if warningCount > 0}
        <div class="issue-stat warning">
          <span>Warnings</span>
          <strong>{warningCount}</strong>
        </div>
      {/if}
      <span class="sr-only">{visibleMessages.length === 1 ? '1 affected record' : `${visibleMessages.length} affected records`}</span>
    </div>
  {/if}
  <!-- svelte-ignore a11y_no_noninteractive_tabindex (bounded overflow regions need a keyboard focus target) -->
  <div
    class:bounded-message-groups={shouldBoundGroups}
    role={shouldBoundGroups ? 'region' : undefined}
    aria-label={shouldBoundGroups ? 'Grouped import issues' : undefined}
    tabindex={shouldBoundGroups ? 0 : undefined}
  >
    {#each visibleGroups as group (group.key)}
      {@const isGroupExpanded = groupExpanded(group)}
      {@const visibleMessages = isGroupExpanded ? group.messages : group.messages.slice(0, COLLAPSED_RECORD_LIMIT)}
      <div class="message-group">
        <div class="message-group-heading">
          <Badge variant={group.severity === 'error' ? 'destructive' : 'secondary'}>{severityLabel(group.severity)}</Badge>
          <div>
            <strong>{group.summary}</strong>
            <span>{group.cause ? `${group.cause} · ${groupCountLabel(group.messages.length)}` : groupCountLabel(group.messages.length)}</span>
          </div>
        </div>
        <div class="message-group-items">
          {#each visibleMessages as message}
            {@const diagnostic = messageDiagnostic(message, group)}
            <div class="message-row">
              <span>{messageRowLabel(message, group)}</span>
              {#if diagnostic}
                <small>{diagnostic}</small>
              {/if}
            </div>
          {/each}
          {#if group.messages.length > visibleMessages.length}
            <Button.Root variant="ghost" size="sm" class="message-group-toggle" onclick={() => toggleGroup(group)}>
              Show {group.messages.length - visibleMessages.length} more in this group
            </Button.Root>
          {:else if isGroupExpanded && group.messages.length > COLLAPSED_RECORD_LIMIT}
            <Button.Root variant="ghost" size="sm" class="message-group-toggle" onclick={() => toggleGroup(group)}>
              Show fewer in this group
            </Button.Root>
          {/if}
        </div>
      </div>
    {/each}
  </div>
  {#if hiddenGroupCount > 0}
    <div class="message-overflow-action">
      <span>{hiddenGroupCount} more issue {hiddenGroupCount === 1 ? 'group' : 'groups'} hidden.</span>
      {#if expanded}
        <Button.Root variant="outline" size="sm" onclick={() => (expanded = false)}>Show fewer</Button.Root>
      {:else}
        <Button.Root variant="outline" size="sm" onclick={() => (expanded = true)}>Show more issues</Button.Root>
      {/if}
    </div>
  {:else if expanded && groups.length > COLLAPSED_GROUP_LIMIT}
    <div class="message-overflow-action">
      <span>All issue groups are shown.</span>
      <Button.Root variant="outline" size="sm" onclick={() => (expanded = false)}>Show fewer</Button.Root>
    </div>
  {/if}
  {#if visibleMessages.length === 0}
    <div class="quiet-row"><CheckCircle2 size={16} aria-hidden="true" /> {emptyText}</div>
  {/if}
  {#if truncated}
    <div class="quiet-row"><AlertCircle size={16} aria-hidden="true" /> {truncatedText}</div>
  {/if}
</div>

<style>
  .message-list {
    display: grid;
    gap: 0.75rem;
  }

  .message-list-summary {
    display: grid;
    gap: 0.5rem;
    grid-template-columns: repeat(auto-fit, minmax(7rem, 1fr));
  }

  .message-overflow-action {
    align-items: center;
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    justify-content: space-between;
  }

  .issue-stat {
    background: color-mix(in oklab, var(--muted) 38%, transparent);
    border: 1px solid var(--border);
    border-radius: 8px;
    display: grid;
    gap: 0.15rem;
    padding: 0.55rem 0.65rem;
  }

  .issue-stat.warning {
    background: color-mix(in oklab, var(--color-warning) 7%, transparent);
    border-color: color-mix(in oklab, var(--color-warning) 26%, transparent);
  }

  .issue-stat.warning strong {
    color: var(--color-warning-foreground);
  }

  .issue-stat.blocking {
    background: color-mix(in oklab, var(--destructive) 9%, transparent);
    border-color: color-mix(in oklab, var(--destructive) 32%, transparent);
  }

  .issue-stat span {
    color: var(--muted-foreground);
    font-size: 0.72rem;
    font-weight: 700;
    text-transform: uppercase;
  }

  .issue-stat strong {
    font-size: 1rem;
    line-height: 1.1;
  }

  .sr-only {
    border: 0;
    clip: rect(0 0 0 0);
    height: 1px;
    margin: -1px;
    overflow: hidden;
    padding: 0;
    position: absolute;
    white-space: nowrap;
    width: 1px;
  }

  .message-overflow-action span {
    color: var(--muted-foreground);
    font-size: 0.82rem;
  }

  :global(.message-group-toggle) {
    justify-self: start;
    padding-inline: 0;
  }

  .message-group {
    border-top: 1px solid var(--border);
    display: grid;
    gap: 0.55rem;
    padding-top: 0.75rem;
  }

  .message-group:first-child {
    border-top: 0;
    padding-top: 0;
  }

  .bounded-message-groups {
    border: 1px solid var(--border);
    border-radius: 8px;
    display: grid;
    gap: 0.75rem;
    max-height: min(26rem, 56vh);
    overflow-y: auto;
    padding: 0.75rem;
  }

  .bounded-message-groups .message-group:first-child {
    padding-top: 0;
  }

  .message-group-heading {
    align-items: flex-start;
    display: flex;
    gap: 0.75rem;
    min-width: 0;
  }

  .message-group-heading > div {
    min-width: 0;
  }

  .message-group-heading strong,
  .message-group-heading span,
  .message-row span,
  .message-row small {
    display: block;
    overflow-wrap: anywhere;
  }

  .message-group-heading span,
  .message-row small {
    color: var(--muted-foreground);
    font-size: 0.78rem;
  }

  .message-group-items {
    display: grid;
    gap: 0.4rem;
  }

  .message-row,
  .quiet-row {
    align-items: flex-start;
    display: flex;
    gap: 0.75rem;
  }

  @media (max-width: 640px) {
    .message-overflow-action {
      align-items: flex-start;
      display: grid;
      justify-content: stretch;
    }

    .message-list-summary {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .bounded-message-groups {
      max-height: min(22rem, 48vh);
    }

    .message-group-heading {
      display: grid;
      gap: 0.45rem;
    }

    .message-row {
      display: block;
    }
  }
</style>
