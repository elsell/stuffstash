<script lang="ts">
  import AlertCircle from '@lucide/svelte/icons/alert-circle';
  import CheckCircle2 from '@lucide/svelte/icons/check-circle-2';
  import X from '@lucide/svelte/icons/x';
  import type { ImportMessage } from '$lib/domain/inventory';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Dialog from '$lib/components/ui/dialog/index.js';
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
  let selectedGroupKey = $state<string | null>(null);
  let visibleGroupLimit = $derived(expanded ? EXPANDED_GROUP_LIMIT : COLLAPSED_GROUP_LIMIT);
  let visibleGroups = $derived(groups.slice(0, visibleGroupLimit));
  let hiddenGroupCount = $derived(Math.max(0, groups.length - visibleGroups.length));
  let shouldBoundGroups = $derived(visibleMessages.length > 12 || groups.length > COLLAPSED_GROUP_LIMIT);
  let selectedGroup = $derived(groups.find((group) => group.key === selectedGroupKey) ?? null);

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

  function issueGuidance(group: MessageGroup): { meaning: string; impact: string; nextAction: string } {
    const message = group.messages[0];
    const code = message?.code ?? '';
    const cause = group.cause.toLowerCase();
    if (code === 'duplicate-asset' || code === 'source-link-duplicate' || cause.includes('already')) {
      return {
        meaning: 'Stuff Stash found records that look connected to an earlier import.',
        impact: 'Those records were skipped so the import would not create duplicates.',
        nextAction: 'Open the matching item in Stuff Stash or review the original Homebox record before importing it again.'
      };
    }
    if (code === 'partial-date' || group.summary.toLowerCase().includes('partial date')) {
      return {
        meaning: 'Homebox has a date that is incomplete or cannot be represented as a full Stuff Stash date.',
        impact: 'The value was kept as text instead of being saved as a structured date.',
        nextAction: 'Edit the date in Homebox or update the imported field in Stuff Stash after the import.'
      };
    }
    if (code === 'attachment-unavailable' || cause.includes('download')) {
      return {
        meaning: 'Stuff Stash could not download one or more files from the source.',
        impact: 'The related asset can still import, but the listed photos or files were skipped.',
        nextAction: 'Check that the file exists in Homebox and that the Homebox URL is reachable, then run a new preview if you still need the file.'
      };
    }
    if (cause.includes('attachment validation') || cause.includes('unsupported file type')) {
      return {
        meaning: 'A file was reachable, but it did not meet Stuff Stash attachment rules.',
        impact: 'The file was skipped and was not attached to the imported asset.',
        nextAction: 'Convert or replace the file with a supported format in Homebox, then preview the import again.'
      };
    }
    if (group.severity === 'error') {
      return {
        meaning: 'This issue blocked part of the import from completing safely.',
        impact: 'Stuff Stash stopped or skipped the affected work to avoid saving misleading data.',
        nextAction: 'Review the affected records, correct the source data if needed, then preview and run the import again.'
      };
    }
    return {
      meaning: 'Stuff Stash imported what it could and preserved this warning for review.',
      impact: 'The affected records may need follow-up, but the warning did not block the whole import.',
      nextAction: 'Review the affected records below and update the source or imported records if the result is not what you want.'
    };
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
          <Badge
            variant={group.severity === 'error' ? 'destructive' : 'secondary'}
            class={group.severity === 'warning' ? 'message-warning-badge' : ''}
          >
            {severityLabel(group.severity)}
          </Badge>
          <div>
            <strong>{group.summary}</strong>
            <span>{group.cause ? `${group.cause} · ${groupCountLabel(group.messages.length)}` : groupCountLabel(group.messages.length)}</span>
          </div>
          <Button.Root variant="ghost" size="sm" class="message-detail-button" onclick={() => (selectedGroupKey = group.key)}>
            Details
          </Button.Root>
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
  <Dialog.Root
    open={Boolean(selectedGroup)}
    ariaLabelledBy="issue-detail-title"
    class="issue-detail-dialog"
    onDismiss={() => (selectedGroupKey = null)}
  >
    {#if selectedGroup}
      {@const guidance = issueGuidance(selectedGroup)}
      <div class="issue-detail-heading">
        <div>
          <Badge
            variant={selectedGroup.severity === 'error' ? 'destructive' : 'secondary'}
            class={selectedGroup.severity === 'warning' ? 'message-warning-badge' : ''}
          >
            {severityLabel(selectedGroup.severity)}
          </Badge>
          <h3 id="issue-detail-title">{selectedGroup.summary}</h3>
          <p>{selectedGroup.cause || groupCountLabel(selectedGroup.messages.length)}</p>
        </div>
        <Button.Root variant="ghost" size="icon" aria-label="Close issue details" onclick={() => (selectedGroupKey = null)}>
          <X size={16} aria-hidden="true" />
        </Button.Root>
      </div>
      <div class="issue-detail-grid">
        <div>
          <span>Meaning</span>
          <p>{guidance.meaning}</p>
        </div>
        <div>
          <span>Impact</span>
          <p>{guidance.impact}</p>
        </div>
        <div>
          <span>Next action</span>
          <p>{guidance.nextAction}</p>
        </div>
      </div>
      <div class="issue-detail-records">
        <h4>Affected records</h4>
        <div>
          {#each selectedGroup.messages.slice(0, 8) as message}
            {@const diagnostic = messageDiagnostic(message, selectedGroup)}
            <div class="message-row compact">
              <span>{messageRowLabel(message, selectedGroup)}</span>
              {#if diagnostic}
                <small>{diagnostic}</small>
              {/if}
            </div>
          {/each}
        </div>
        {#if selectedGroup.messages.length > 8}
          <small>{selectedGroup.messages.length - 8} more affected {selectedGroup.messages.length - 8 === 1 ? 'record' : 'records'} in this group.</small>
        {/if}
      </div>
    {/if}
  </Dialog.Root>
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
    flex: 1 1 auto;
    min-width: 0;
  }

  :global(.message-detail-button) {
    flex: 0 0 auto;
  }

  :global(.message-warning-badge) {
    background: color-mix(in oklab, var(--color-warning) 16%, transparent);
    color: var(--color-warning-foreground);
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

  .message-row.compact {
    background: color-mix(in oklab, var(--muted) 20%, transparent);
    border: 1px solid color-mix(in oklab, var(--border) 72%, transparent);
    border-radius: 8px;
    padding: 0.5rem 0.65rem;
  }

  h3 {
    font-size: 1rem;
    margin: 0;
  }

  h4 {
    font-size: 0.88rem;
    margin: 0;
  }

  :global(.issue-detail-dialog) {
    background: var(--background);
    border: 1px solid var(--border);
    border-radius: 8px;
    box-shadow: 0 1.5rem 4rem color-mix(in oklab, var(--foreground) 18%, transparent);
    display: grid;
    gap: 1rem;
    max-height: min(42rem, calc(100vh - 2rem));
    max-width: min(36rem, 100%);
    padding: 1rem;
  }

  .issue-detail-heading {
    align-items: flex-start;
    display: flex;
    gap: 0.75rem;
    justify-content: space-between;
  }

  .issue-detail-heading > div {
    display: grid;
    gap: 0.35rem;
    min-width: 0;
  }

  .issue-detail-heading p,
  .issue-detail-grid p,
  .issue-detail-records > small {
    color: var(--muted-foreground);
    font-size: 0.86rem;
    margin: 0;
    overflow-wrap: anywhere;
  }

  .issue-detail-grid {
    display: grid;
    gap: 0.55rem;
  }

  .issue-detail-grid > div {
    border-top: 1px solid var(--border);
    display: grid;
    gap: 0.18rem;
    padding-top: 0.55rem;
  }

  .issue-detail-grid > div:first-child {
    border-top: 0;
    padding-top: 0;
  }

  .issue-detail-grid span {
    color: var(--foreground);
    font-size: 0.78rem;
    font-weight: 700;
    text-transform: uppercase;
  }

  .issue-detail-records {
    display: grid;
    gap: 0.5rem;
  }

  .issue-detail-records > div {
    display: grid;
    gap: 0.35rem;
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
