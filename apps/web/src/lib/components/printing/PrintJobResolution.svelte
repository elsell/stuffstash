<script lang="ts">
  import {t} from '$lib/presentation/localization';
  import type {PrintJob,ReportedPrintOutcome} from '$lib/domain/printing';
  import {Button} from '$lib/components/ui/button/index.js';
  import {Label} from '$lib/components/ui/label/index.js';
  import {Checkbox} from '$lib/components/ui/checkbox/index.js';
  import PairingChoice from './PairingChoice.svelte';
  let {job,busy,onResolve}:{job:PrintJob;busy:boolean;onResolve:(job:PrintJob,outcome:ReportedPrintOutcome)=>Promise<void>}=$props();
  let outcome=$state(''),acknowledged=$state(false);
  let submitted=$state<{job:PrintJob;outcome:ReportedPrintOutcome}|null>(null);
  const choices=$derived([{value:'printed',label:t('web.Printing.reportPrinted')},{value:'not_printed',label:t('web.Printing.reportNotPrinted')},{value:'unknown',label:t('web.Printing.reportUnknown')}]);
  async function resolve(event:SubmitEvent){
    event.preventDefault();if(busy)return;
    if(!submitted){if(!acknowledged||!['printed','not_printed','unknown'].includes(outcome))return;submitted={job:{...job},outcome:outcome as ReportedPrintOutcome};}
    await onResolve(submitted.job,submitted.outcome);
  }
</script>
<form class="resolution-form" onsubmit={resolve}>
  <PairingChoice id={`resolution-outcome-${job.id}`} label={t('web.Printing.observedOutcome')} value={outcome} options={choices} disabled={busy||Boolean(submitted)} onChange={value=>{outcome=value;}}/>
  <div class="acknowledgement"><Checkbox id={`resolution-ack-${job.id}`} bind:checked={acknowledged} disabled={busy||Boolean(submitted)}/><Label for={`resolution-ack-${job.id}`}>{t('web.Printing.acknowledgeUncertainty')}</Label></div>
  {#if submitted}<p>{t('web.Printing.resolutionRetryHelp')}</p>{/if}
  <Button type="submit" variant="outline" disabled={busy||!acknowledged||!outcome}>{t(submitted?'web.Printing.retryResolution':'web.Printing.resolveOutcome')}</Button>
</form>
<style>.resolution-form{display:grid;gap:var(--space-3);margin-block:var(--space-3);max-width:32rem}.acknowledgement{display:flex;gap:var(--space-3);align-items:start}p{overflow-wrap:anywhere}</style>
