<script lang="ts">
  import CheckCircle2 from '@lucide/svelte/icons/check-circle-2';
  import Eye from '@lucide/svelte/icons/eye';
  import type { ImportJob } from '$lib/domain/inventory';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import { isTerminal, phaseLabel, progressSummary, sourceDescription, statusSentence } from './importWorkspacePresentation';
  import ImportFlowStepper from './ImportFlowStepper.svelte';

  type Props = {
    job: ImportJob;
    onViewHistory: () => void;
  };

  let { job, onViewHistory }: Props = $props();
</script>

<Card.Root>
  <Card.Header>
    <ImportFlowStepper current="run" />
    <Card.Title>{isTerminal(job) ? 'Import finished' : 'Import is running'}</Card.Title>
    <Card.Description>
      {isTerminal(job) ? statusSentence(job) : 'You can leave this page and return from import history.'}
    </Card.Description>
  </Card.Header>
  <Card.Content class="run-handoff-content">
    <div class="handoff-panel" role="status">
      <CheckCircle2 size={20} aria-hidden="true" />
      <div>
        <strong>{job.source.name}</strong>
        <span>{sourceDescription(job)} · {phaseLabel(job)} · {progressSummary(job)}</span>
      </div>
    </div>
    <Button.Root onclick={onViewHistory}>
      <Eye size={16} aria-hidden="true" />
      View in history
    </Button.Root>
  </Card.Content>
</Card.Root>

<style>
  :global(.run-handoff-content) {
    display: grid;
    gap: 1rem;
  }

  .handoff-panel {
    align-items: center;
    background: hsl(var(--primary) / 0.06);
    border: 1px solid hsl(var(--primary) / 0.28);
    border-radius: 8px;
    display: flex;
    gap: 0.75rem;
    padding: 0.85rem;
  }

  .handoff-panel span {
    color: hsl(var(--muted-foreground));
    display: block;
    font-size: 0.85rem;
    margin-top: 0.15rem;
  }

  @media (max-width: 640px) {
    .handoff-panel {
      align-items: flex-start;
    }
  }
</style>
