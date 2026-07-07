<script lang="ts">
  import Play from '@lucide/svelte/icons/play';
  import type { ImportJob } from '$lib/domain/inventory';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import {
    previewReadinessBadge,
    previewReadinessDescription,
    previewReadinessTitle,
    sourceDescription,
    sourceOptionsSummary,
    sourceSnapshotDescription,
    visiblePreviewCountCells,
    visiblePreviewMessages
  } from './importWorkspacePresentation';
  import ImportCountGrid from './ImportCountGrid.svelte';
  import ImportFlowStepper from './ImportFlowStepper.svelte';
  import ImportMessagesList from './ImportMessagesList.svelte';
  import ImportPreviewSamples from './ImportPreviewSamples.svelte';

  type Props = {
    previewJob: ImportJob | null;
    previewReady: boolean;
    previewStale: boolean;
    busy: boolean;
    onStart: () => void;
    onBack: () => void;
  };

  let {
    previewJob,
    previewReady,
    previewStale,
    busy,
    onStart,
    onBack
  }: Props = $props();
</script>

<Card.Root>
  <Card.Header>
    <ImportFlowStepper current="preview" />
    <Card.Title>Preview import</Card.Title>
    <Card.Description>Review the plan before starting the background import.</Card.Description>
  </Card.Header>
  <Card.Content class="import-preview-content">
    {#if previewJob}
      <div class={`readiness-panel ${previewStale || previewJob.counts.errors > 0 ? 'needs-attention' : ''}`}>
        <div>
          <strong>{previewReadinessTitle(previewJob, previewStale)}</strong>
          <span>{previewReadinessDescription(previewJob, previewStale)}</span>
        </div>
        <Badge variant={previewStale || previewJob.counts.errors > 0 ? 'destructive' : 'secondary'}>
          {previewReadinessBadge(previewJob, previewStale)}
        </Badge>
      </div>
      <div class="source-summary">
        <div>
          <span>{sourceDescription(previewJob)}</span>
          <small>{sourceSnapshotDescription(previewJob)}</small>
        </div>
        <ul class="source-option-list" aria-label="Selected source options">
          {#each sourceOptionsSummary(previewJob) as option}
            <li>{option}</li>
          {/each}
        </ul>
      </div>
      <ImportCountGrid cells={visiblePreviewCountCells(previewJob)} />
      <ImportPreviewSamples preview={previewJob.preview} />
      <ImportMessagesList messages={visiblePreviewMessages(previewJob)} emptyText="No blocking issues found." truncated={previewJob.preview.messagesTruncated} />
    {/if}

    <div class="action-row">
      <Button.Root onclick={onStart} disabled={!previewReady || busy}>
        <Play size={16} aria-hidden="true" />
        Start background import
      </Button.Root>
      <Button.Root variant="outline" onclick={onBack} disabled={busy}>Back</Button.Root>
    </div>
  </Card.Content>
</Card.Root>

<style>
  .import-preview-content {
    display: grid;
    gap: 1rem;
  }

  .action-row,
  .quiet-row,
  .message-row {
    align-items: center;
    display: flex;
    gap: 0.75rem;
  }

  .action-row {
    flex-wrap: wrap;
  }

  .source-summary,
  .readiness-panel {
    align-items: center;
    display: flex;
    gap: 0.75rem;
    justify-content: space-between;
  }

  .readiness-panel {
    background: hsl(var(--primary) / 0.06);
    border: 1px solid hsl(var(--primary) / 0.28);
    border-radius: 8px;
    padding: 0.85rem;
  }

  .readiness-panel.needs-attention {
    background: hsl(var(--destructive) / 0.06);
    border-color: hsl(var(--destructive) / 0.3);
  }

  .readiness-panel span {
    color: hsl(var(--muted-foreground));
    display: block;
    font-size: 0.86rem;
    margin-top: 0.2rem;
  }

  .source-summary {
    background: hsl(var(--muted) / 0.45);
    border: 1px solid hsl(var(--border));
    border-radius: 8px;
    padding: 0.75rem;
  }

  .source-summary span,
  .source-summary small {
    display: block;
  }

  .source-option-list {
    display: grid;
    gap: 0.35rem;
    justify-items: end;
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .source-option-list li {
    border-left: 2px solid hsl(var(--border));
    color: hsl(var(--foreground));
    font-size: 0.78rem;
    line-height: 1.3;
    padding-left: 0.5rem;
  }

  .source-summary small {
    color: hsl(var(--muted-foreground));
    font-size: 0.78rem;
  }

  @media (max-width: 860px) {
    .action-row {
      align-items: flex-start;
      flex-direction: column;
    }

    .source-summary,
    .readiness-panel {
      align-items: flex-start;
      flex-direction: column;
    }

    .source-option-list {
      justify-items: start;
    }

  }
</style>
