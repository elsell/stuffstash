<script lang="ts">
  import StepProgress, { type StepProgressStep } from './step-progress.svelte';

  let {
    onNavigateStep = () => {},
    orientation = 'horizontal',
    density = 'compact'
  }: {
    onNavigateStep?: (stepId: string) => void;
    orientation?: 'horizontal' | 'vertical';
    density?: 'compact' | 'comfortable';
  } = $props();

  const steps: StepProgressStep[] = [
    { id: 'account', label: 'Account', description: 'Create your profile' },
    { id: 'home', label: 'Home', description: 'Name your household' },
    { id: 'invite', label: 'Invite', description: 'Invite trusted people' },
    { id: 'done', label: 'Done', description: 'Finish setup' }
  ];
</script>

<StepProgress
  {steps}
  current="invite"
  reachableStepIds={['account', 'home', 'invite']}
  ariaLabel="Onboarding progress"
  {orientation}
  {density}
  {onNavigateStep}
/>
