<script lang="ts">
  import StepProgress, { type StepProgressStep } from './step-progress.svelte';

  let {
    onNavigateStep = () => {}
  }: {
    onNavigateStep?: (stepId: string) => void;
  } = $props();

  const steps: StepProgressStep[] = [
    { id: 'account', label: 'Account' },
    { id: 'home', label: 'Home' },
    { id: 'invite', label: 'Invite' },
    { id: 'done', label: 'Done' }
  ];
</script>

<StepProgress
  {steps}
  current="invite"
  reachableStepIds={['account', 'home', 'invite']}
  ariaLabel="Onboarding progress"
  {onNavigateStep}
/>
