<script lang="ts">
  import { t } from '$lib/presentation/localization';
  import Play from '@lucide/svelte/icons/play';
  import type { ImportJob } from '$lib/domain/inventory';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import {
    previewReadinessBadge,
    previewReadinessDescription,
    previewReadinessTitle,
    reportedErrorCount,
    reportedWarningCount,
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
    availableSteps: Array<'source' | 'connect' | 'preview' | 'run'>;
    previewReady: boolean;
    previewStale: boolean;
    busy: boolean;
    onStart: () => void;
    onNavigateStep: (step: 'source' | 'connect' | 'preview' | 'run') => void;
    onBack: () => void;
  };

  let {
    previewJob,
    availableSteps,
    previewReady,
    previewStale,
    busy,
    onStart,
    onNavigateStep,
    onBack
  }: Props = $props();
</script>

<Card.Root>
  <Card.Header>
    <ImportFlowStepper current="preview" {availableSteps} {onNavigateStep} />
    <Card.Title>{t('web.ImportPreviewPanel.previewImport')}</Card.Title>
    <Card.Description>{t('web.ImportPreviewPanel.reviewThePlanBeforeStartingTheBackgroundImport')}</Card.Description>
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
        <ul class="source-option-list" aria-label={t('web.ImportPreviewPanel.selectedSourceOptions')}>
          {#each sourceOptionsSummary(previewJob) as option}
            <li>{option}</li>
          {/each}
        </ul>
      </div>
      <ImportCountGrid cells={visiblePreviewCountCells(previewJob)} />
      <section class="preview-issues-section" aria-label={t('web.ImportPreviewPanel.previewIssues')}>
        <div class="preview-section-heading">
          <h3>{t('web.ImportPreviewPanel.issues')}</h3>
          <small>{previewJob.counts.warnings + previewJob.counts.errors === 0 ? t('web.ImportPreviewPanel.noBlockers') : t('web.ImportPreviewPanel.groupedByCause')}</small>
        </div>
        <ImportMessagesList
          messages={visiblePreviewMessages(previewJob)}
          emptyText={t('web.ImportPreviewPanel.noBlockingIssuesFound')}
          truncated={previewJob.preview.messagesTruncated}
          reportedWarnings={reportedWarningCount(previewJob)}
          reportedErrors={reportedErrorCount(previewJob)}
        />
      </section>
      <section class="preview-plan-section" aria-label={t('web.ImportPreviewPanel.planPreview')}>
        <div class="preview-section-heading">
          <h3>{t('web.ImportPreviewPanel.planPreview')}</h3>
          <small>{t('web.ImportPreviewPanel.plannedRecordsBySection')}</small>
        </div>
        <ImportPreviewSamples preview={previewJob.preview} />
      </section>
    {/if}

    <div class="action-row">
      <Button.Root onclick={onStart} disabled={!previewReady || busy}>
        <Button.BusyContent {busy} icon={Play} label={t('web.ImportPreviewPanel.startBackgroundImport')} busyLabel={t('web.ImportPreviewPanel.startingImport')} />
      </Button.Root>
      <Button.Root variant="outline" onclick={onBack} disabled={busy}>{t('web.ImportPreviewPanel.back')}</Button.Root>
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
    scroll-margin-bottom: var(--mobile-scroll-clearance, 10rem);
  }

  .source-summary,
  .readiness-panel {
    align-items: center;
    display: flex;
    gap: 0.75rem;
    justify-content: space-between;
  }

  .readiness-panel {
    background: color-mix(in oklab, var(--primary) 6%, transparent);
    border: 1px solid color-mix(in oklab, var(--primary) 28%, transparent);
    border-radius: 8px;
    padding: 0.85rem;
  }

  .readiness-panel.needs-attention {
    background: color-mix(in oklab, var(--destructive) 6%, transparent);
    border-color: color-mix(in oklab, var(--destructive) 30%, transparent);
  }

  .readiness-panel span {
    color: var(--muted-foreground);
    display: block;
    font-size: 0.86rem;
    margin-top: 0.2rem;
  }

  .source-summary {
    background: color-mix(in oklab, var(--muted) 45%, transparent);
    border: 1px solid var(--border);
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
    border-left: 2px solid var(--border);
    color: var(--foreground);
    font-size: 0.78rem;
    line-height: 1.3;
    padding-left: 0.5rem;
  }

  .source-summary small {
    color: var(--muted-foreground);
    font-size: 0.78rem;
  }

  .preview-issues-section,
  .preview-plan-section {
    display: grid;
    gap: 0.75rem;
    min-width: 0;
  }

  .preview-section-heading {
    align-items: baseline;
    border-top: 1px solid var(--border);
    display: flex;
    gap: 0.5rem;
    justify-content: space-between;
    padding-top: 0.85rem;
  }

  .preview-section-heading h3 {
    font-size: 1rem;
    margin: 0;
  }

  .preview-section-heading small {
    color: var(--muted-foreground);
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

    .preview-section-heading {
      align-items: flex-start;
      display: grid;
      gap: 0.2rem;
    }
  }
</style>
