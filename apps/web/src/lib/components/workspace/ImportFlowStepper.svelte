<script lang="ts">
  import { StepProgress, type StepProgressStep } from '$lib/components/ui/step-progress/index.js';

  type StepID = 'source' | 'connect' | 'preview' | 'run';

  type Props = {
    current: StepID;
    availableSteps?: StepID[];
    onNavigateStep?: (step: StepID) => void;
  };

  const steps: StepProgressStep[] = [
    { id: 'source', label: 'Source' },
    { id: 'connect', label: 'Connect' },
    { id: 'preview', label: 'Preview' },
    { id: 'run', label: 'Run' }
  ];

  let { current, availableSteps = ['source'], onNavigateStep }: Props = $props();

  function navigate(stepId: string): void {
    if (stepId === 'source' || stepId === 'connect' || stepId === 'preview' || stepId === 'run') {
      onNavigateStep?.(stepId);
    }
  }
</script>

<StepProgress {steps} {current} reachableStepIds={availableSteps} ariaLabel="Import progress" onNavigateStep={navigate} />
