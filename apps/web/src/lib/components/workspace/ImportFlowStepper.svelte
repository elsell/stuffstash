<script lang="ts">
  import { t } from '$lib/presentation/localization';
  import { StepProgress, type StepProgressStep } from '$lib/components/ui/step-progress/index.js';

  type StepID = 'source' | 'connect' | 'preview' | 'run';

  type Props = {
    current: StepID;
    availableSteps?: StepID[];
    onNavigateStep?: (step: StepID) => void;
  };

  const steps: StepProgressStep[] = [
    { id: 'source', label: t('web.ImportFlowStepper.source') },
    { id: 'connect', label: t('web.ImportFlowStepper.connect') },
    { id: 'preview', label: t('web.ImportFlowStepper.preview') },
    { id: 'run', label: t('web.ImportFlowStepper.run') }
  ];

  let { current, availableSteps = ['source'], onNavigateStep }: Props = $props();

  function navigate(stepId: string): void {
    if (stepId === 'source' || stepId === 'connect' || stepId === 'preview' || stepId === 'run') {
      onNavigateStep?.(stepId);
    }
  }
</script>

<StepProgress {steps} {current} reachableStepIds={availableSteps} ariaLabel="Import progress" onNavigateStep={navigate} />
