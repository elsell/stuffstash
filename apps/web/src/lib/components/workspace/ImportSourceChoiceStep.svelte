<script lang="ts">
  import Database from '@lucide/svelte/icons/database';
  import FileText from '@lucide/svelte/icons/file-text';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import type { ImportSourceChoice } from '$lib/application/workspaceImportRequest';
  import ImportFlowStepper from './ImportFlowStepper.svelte';

  type Props = {
    liveHref: string;
    csvHref: string;
    onChoose: (event: MouseEvent, choice: ImportSourceChoice) => void;
    onCancel: () => void;
  };

  let { liveHref, csvHref, onChoose, onCancel }: Props = $props();
</script>

<Card.Root>
  <Card.Header>
    <ImportFlowStepper current="source" />
    <Card.Title>Choose import method</Card.Title>
    <Card.Description>Pick the path that matches the data you have right now.</Card.Description>
  </Card.Header>
  <Card.Content>
    <div class="import-source-choice-content">
      <div class="source-choice-grid" role="group" aria-label="Homebox import method">
        <Button.Root variant="outline" class="source-card" href={liveHref} onclick={(event) => onChoose(event, 'homebox_live')}>
          <span class="source-choice-icon"><Database size={24} aria-hidden="true" /></span>
          <span class="source-choice-copy">
            <strong>Connect to Homebox</strong>
            <small>Use your Homebox URL and credentials. Best when the instance is reachable and you want photos from the live API.</small>
            <em>Can include photos · checks the source before running</em>
          </span>
        </Button.Root>
        <Button.Root variant="outline" class="source-card" href={csvHref} onclick={(event) => onChoose(event, 'homebox_csv')}>
          <span class="source-choice-icon"><FileText size={24} aria-hidden="true" /></span>
          <span class="source-choice-copy">
            <strong>Upload Homebox CSV</strong>
            <small>Use an exported CSV file. Best for offline imports, migrations from an older instance, or when the Homebox API is not reachable.</small>
            <em>No photos in CSV · works without a live server</em>
          </span>
        </Button.Root>
      </div>
      <div class="action-row">
        <Button.Root variant="outline" onclick={onCancel}>Cancel</Button.Root>
      </div>
    </div>
  </Card.Content>
</Card.Root>

<style>
  .import-source-choice-content {
    display: grid;
    gap: 1rem;
  }

  .source-choice-grid {
    display: grid;
    gap: 1rem;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .source-choice-grid :global(.source-card) {
    align-items: flex-start;
    display: flex;
    gap: 0.85rem;
    height: auto;
    justify-content: flex-start;
    min-height: 8rem;
    padding: 1rem;
    text-align: left;
    white-space: normal;
  }

  .source-choice-icon {
    display: inline-flex;
    flex: 0 0 auto;
    margin-top: 0.1rem;
  }

  .source-choice-copy {
    min-width: 0;
  }

  .source-choice-grid strong,
  .source-choice-grid small {
    display: block;
  }

  .source-choice-grid small,
  .source-choice-grid em {
    color: hsl(var(--muted-foreground));
    font-size: 0.85rem;
    font-weight: 400;
    line-height: 1.35;
    margin-top: 0.35rem;
  }

  .source-choice-grid em {
    color: hsl(var(--foreground));
    font-size: 0.78rem;
    font-style: normal;
    font-weight: 600;
  }

  .action-row {
    align-items: center;
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem;
  }

  @media (max-width: 860px) {
    .action-row {
      align-items: flex-start;
      flex-direction: column;
    }

    .source-choice-grid {
      grid-template-columns: 1fr;
    }

    .source-choice-grid :global(.source-card) {
      gap: 0.75rem;
      min-height: 0;
      padding: 0.85rem;
    }

    .source-choice-icon {
      height: 1.25rem;
      width: 1.25rem;
    }

    .source-choice-grid small {
      font-size: 0.82rem;
      line-height: 1.32;
    }

    .source-choice-grid em {
      font-size: 0.78rem;
      line-height: 1.28;
    }
  }
</style>
